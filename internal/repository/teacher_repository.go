// internal/repository/teacher_repository.go
package repository

import (
	"database/sql"
	"edugame/internal/entity"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/lib/pq"
)

type TeacherRepository struct {
	db *sql.DB
}

func NewTeacherRepository(db *sql.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// Получить классы учителя
func (r *TeacherRepository) GetTeacherClass(teacherID int) (struct {
	ID    int
	Name  string
	Grade int
}, error) {
	query := `
		SELECT id, name, grade 
		FROM classes 
		WHERE teacher_id = $1
		ORDER BY grade, name
	`
	var class struct {
		ID    int
		Name  string
		Grade int
	}

	err := r.db.QueryRow(query, teacherID).Scan(&class.ID, &class.Name, &class.Grade)

	if err != nil {
		return class, err
	}

	return class, nil
}

// repository/teacher_repository.go

// GetAllClasses получает все классы из базы данных
func (r *TeacherRepository) GetAllClasses() ([]*entity.Class, error) {
	query := `
        SELECT id, name, grade, teacher_id 
        FROM classes 
        ORDER BY grade, name
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения классов: %w", err)
	}
	defer rows.Close()

	var classes []*entity.Class
	for rows.Next() {
		var class entity.Class
		err := rows.Scan(
			&class.ID,
			&class.Name,
			&class.Grade,
			&class.TeacherID,
		)
		if err != nil {
			continue // пропускаем ошибки чтения
		}
		classes = append(classes, &class)
	}

	return classes, nil
}

// Получить учеников класса
func (r *TeacherRepository) GetClassStudents(classID int) ([]struct {
	ID       int
	Username string
	FullName string
}, error) {
	query := `
		SELECT u.id, u.username, u.fullname
		FROM users u
		JOIN student_classes sc ON u.id = sc.student_id
		JOIN roles r ON u.role_id = r.id
		WHERE sc.class_id = $1 AND r.name = 'student'
		ORDER BY u.fullname
	`

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []struct {
		ID       int
		Username string
		FullName string
	}

	for rows.Next() {
		var student struct {
			ID       int
			Username string
			FullName string
		}
		if err := rows.Scan(&student.ID, &student.Username, &student.FullName); err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// Получить статистику по классу
func (r *TeacherRepository) GetClassStatistics(classID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Основная статистика класса
	query := `
        SELECT 
            COUNT(DISTINCT u.id) as student_count,
            COUNT(DISTINCT a.id) as total_attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_attempts
        FROM users u
        LEFT JOIN attempts a ON u.id = a.user_id
        JOIN student_classes sc ON u.id = sc.student_id
		JOIN roles r on r.id = u.role_id
        WHERE sc.class_id = $1 AND r.name = 'student'
    `

	var studentCount, totalAttempts, correctAttempts int
	err := r.db.QueryRow(query, classID).Scan(&studentCount, &totalAttempts, &correctAttempts)
	if err != nil {
		return nil, err
	}

	stats["student_count"] = studentCount
	stats["total_attempts"] = totalAttempts
	stats["correct_attempts"] = correctAttempts

	if totalAttempts > 0 {
		stats["accuracy_percent"] = float64(correctAttempts) / float64(totalAttempts) * 100
	} else {
		stats["accuracy_percent"] = 0
	}

	// Активность по дням (последние 7 дней)
	activityQuery := `
        SELECT 
            DATE(a.created_at) as date,
            COUNT(*) as attempts,
            SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) as correct
        FROM attempts a
        JOIN users u ON a.user_id = u.id
        JOIN student_classes sc ON u.id = sc.student_id
        WHERE sc.class_id = $1 
          AND a.created_at >= CURRENT_DATE - INTERVAL '7 days'
        GROUP BY DATE(a.created_at)
        ORDER BY date DESC
    `

	rows, err := r.db.Query(activityQuery, classID)
	if err == nil {
		defer rows.Close()

		var activity []map[string]interface{}
		for rows.Next() {
			var date time.Time
			var attempts, correct int

			if err := rows.Scan(&date, &attempts, &correct); err != nil {
				continue
			}

			activity = append(activity, map[string]interface{}{
				"date":     date.Format("02.01"),
				"attempts": attempts,
				"correct":  correct,
				"accuracy": func() float64 {
					if attempts > 0 {
						return float64(correct) / float64(attempts) * 100
					}
					return 0
				}(),
			})
		}
		stats["activity"] = activity
	}

	// Статистика по типам уравнений
	typeStatsQuery := `
        SELECT 
            et.name as type_name,
            COUNT(a.id) as attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct
        FROM equation_types et
        LEFT JOIN attempts a ON et.id = a.equation_type_id
        LEFT JOIN users u ON a.user_id = u.id
        LEFT JOIN student_classes sc ON u.id = sc.student_id
        WHERE sc.class_id = $1
        GROUP BY et.id, et.name
        HAVING COUNT(a.id) > 0
        ORDER BY attempts DESC
    `

	typeRows, err := r.db.Query(typeStatsQuery, classID)
	if err == nil {
		defer typeRows.Close()

		var typeStats []map[string]interface{}
		for typeRows.Next() {
			var typeName string
			var attempts, correct int

			if err := typeRows.Scan(&typeName, &attempts, &correct); err != nil {
				continue
			}

			typeStats = append(typeStats, map[string]interface{}{
				"type_name": typeName,
				"attempts":  attempts,
				"correct":   correct,
				"accuracy": func() float64 {
					if attempts > 0 {
						return float64(correct) / float64(attempts) * 100
					}
					return 0
				}(),
			})
		}
		stats["type_statistics"] = typeStats
	}

	// Топ учеников
	topStudentsQuery := `
        SELECT 
            u.id,
            u.fullname,
            COUNT(a.id) as total_attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_attempts
        FROM users u
        LEFT JOIN attempts a ON u.id = a.user_id
        JOIN student_classes sc ON u.id = sc.student_id
        WHERE sc.class_id = $1
        GROUP BY u.id, u.fullname
        ORDER BY correct_attempts DESC
        LIMIT 5
    `

	studentRows, err := r.db.Query(topStudentsQuery, classID)
	if err == nil {
		defer studentRows.Close()

		var topStudents []map[string]interface{}
		for studentRows.Next() {
			var studentID, total, correct int
			var fullname string

			if err := studentRows.Scan(&studentID, &fullname, &total, &correct); err != nil {
				continue
			}

			topStudents = append(topStudents, map[string]interface{}{
				"id":      studentID,
				"name":    fullname,
				"total":   total,
				"correct": correct,
				"accuracy": func() float64 {
					if total > 0 {
						return float64(correct) / float64(total) * 100
					}
					return 0
				}(),
			})
		}
		stats["top_students"] = topStudents
	}

	return stats, nil
}

// GetClassesStatistics получает статистику по всем классам
func (r *TeacherRepository) GetClassesStatistics() (map[int]map[string]interface{}, error) {
	// Получаем все классы
	classesQuery := `
        SELECT id, name, grade 
        FROM classes 
        ORDER BY grade, name
    `

	rows, err := r.db.Query(classesQuery)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить классы: %w", err)
	}
	defer rows.Close()

	classes := make(map[int]map[string]interface{})
	var classIDs []int

	for rows.Next() {
		var id, grade int
		var name string

		if err := rows.Scan(&id, &name, &grade); err != nil {
			continue
		}

		classes[id] = map[string]interface{}{
			"class_id":                 id,
			"class_name":               name,
			"grade":                    grade,
			"student_count":            0,
			"total_attempts":           0,
			"correct_attempts":         0,
			"accuracy_percent":         0.0,
			"avg_attempts_per_student": 0.0,
		}
		classIDs = append(classIDs, id)
	}

	if len(classIDs) == 0 {
		// Возвращаем общую статистику с правильными полями
		return map[int]map[string]interface{}{
			0: {
				"total_classes":         0,
				"total_students":        0,
				"total_attempts":        0,
				"total_correct":         0,
				"overall_accuracy":      0.0,
				"most_active_class":     nil,
				"best_performing_class": nil,
				// Дублирующие поля для шаблона
				"student_count":    0,
				"correct_attempts": 0,
				"accuracy_percent": 0.0,
				"top_students":     []map[string]interface{}{},
			},
		}, nil
	}

	// Основная статистика по всем классам
	statsQuery := `
        SELECT 
            sc.class_id,
            COUNT(DISTINCT u.id) as student_count,
            COUNT(DISTINCT a.id) as total_attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_attempts
        FROM student_classes sc
        JOIN users u ON sc.student_id = u.id 
		JOIN roles r ON u.role_id = r.id AND r.name = 'student'
        LEFT JOIN attempts a ON u.id = a.user_id
        WHERE sc.class_id = ANY($1)
        GROUP BY sc.class_id
    `

	statsRows, err := r.db.Query(statsQuery, pq.Array(classIDs))
	if err != nil {
		fmt.Printf("Предупреждение: не удалось получить основную статистику: %v\n", err)
	} else {
		defer statsRows.Close()

		for statsRows.Next() {
			var classID, studentCount, totalAttempts, correctAttempts int

			if err := statsRows.Scan(&classID, &studentCount, &totalAttempts, &correctAttempts); err != nil {
				continue
			}

			if classData, exists := classes[classID]; exists {
				classData["student_count"] = studentCount
				classData["total_attempts"] = totalAttempts
				classData["correct_attempts"] = correctAttempts

				if totalAttempts > 0 {
					classData["accuracy_percent"] = float64(correctAttempts) / float64(totalAttempts) * 100
				}

				if studentCount > 0 {
					classData["avg_attempts_per_student"] = float64(totalAttempts) / float64(studentCount)
				}
			}
		}
	}

	// Активность по классам (последние 7 дней)
	activityQuery := `
        SELECT 
            sc.class_id,
            DATE(a.created_at) as date,
            COUNT(*) as attempts,
            SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) as correct
        FROM attempts a
        JOIN users u ON a.user_id = u.id
        JOIN student_classes sc ON u.id = sc.student_id
        WHERE sc.class_id = ANY($1)
          AND a.created_at >= CURRENT_DATE - INTERVAL '7 days'
        GROUP BY sc.class_id, DATE(a.created_at)
    `

	activityRows, err := r.db.Query(activityQuery, pq.Array(classIDs))
	if err == nil {
		defer activityRows.Close()

		for activityRows.Next() {
			var classID int
			var date time.Time
			var attempts, correct int

			if err := activityRows.Scan(&classID, &date, &attempts, &correct); err != nil {
				continue
			}

			if classData, exists := classes[classID]; exists {
				// Безопасное добавление активности
				activity := classData["activity"]
				if activity == nil {
					activity = []map[string]interface{}{}
				}

				activityList, _ := activity.([]map[string]interface{})
				activityList = append(activityList, map[string]interface{}{
					"date":     date.Format("02.01"),
					"attempts": attempts,
					"correct":  correct,
					"accuracy": func() float64 {
						if attempts > 0 {
							return float64(correct) / float64(attempts) * 100
						}
						return 0
					}(),
				})
				classData["activity"] = activityList
			}
		}
	}

	// Статистика по типам уравнений для всех классов
	typeStatsQuery := `
        SELECT 
            sc.class_id,
            et.name as type_name,
            COUNT(a.id) as attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct
        FROM equation_types et
        CROSS JOIN student_classes sc
        LEFT JOIN users u ON sc.student_id = u.id
        LEFT JOIN attempts a ON et.id = a.equation_type_id AND a.user_id = u.id
        WHERE sc.class_id = ANY($1)
        GROUP BY sc.class_id, et.id, et.name
        HAVING COUNT(a.id) > 0
        ORDER BY sc.class_id, attempts DESC
    `

	typeRows, err := r.db.Query(typeStatsQuery, pq.Array(classIDs))
	if err == nil {
		defer typeRows.Close()

		for typeRows.Next() {
			var classID, attempts, correct int
			var typeName string

			if err := typeRows.Scan(&classID, &typeName, &attempts, &correct); err != nil {
				continue
			}

			if classData, exists := classes[classID]; exists {
				// Безопасное добавление статистики типов
				typeStats := classData["type_statistics"]
				if typeStats == nil {
					typeStats = []map[string]interface{}{}
				}

				typeStatsList, _ := typeStats.([]map[string]interface{})
				typeStatsList = append(typeStatsList, map[string]interface{}{
					"type_name": typeName,
					"attempts":  attempts,
					"correct":   correct,
					"accuracy": func() float64 {
						if attempts > 0 {
							return float64(correct) / float64(attempts) * 100
						}
						return 0
					}(),
				})
				classData["type_statistics"] = typeStatsList
			}
		}
	}

	// Топ учеников по каждому классу
	topStudentsQuery := `
        SELECT 
            sc.class_id,
            u.id,
            u.fullname,
            COUNT(a.id) as total_attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_attempts
        FROM users u
        LEFT JOIN attempts a ON u.id = a.user_id
        JOIN student_classes sc ON u.id = sc.student_id
        WHERE sc.class_id = ANY($1)
        GROUP BY sc.class_id, u.id, u.fullname
        ORDER BY sc.class_id, correct_attempts DESC
    `

	studentRows, err := r.db.Query(topStudentsQuery, pq.Array(classIDs))
	if err == nil {
		defer studentRows.Close()

		tempTopStudents := make(map[int][]map[string]interface{})

		for studentRows.Next() {
			var classID, studentID, total, correct int
			var fullname string

			if err := studentRows.Scan(&classID, &studentID, &fullname, &total, &correct); err != nil {
				continue
			}

			studentData := map[string]interface{}{
				"id":      studentID,
				"name":    fullname,
				"total":   total,
				"correct": correct,
				"accuracy": func() float64 {
					if total > 0 {
						return float64(correct) / float64(total) * 100
					}
					return 0
				}(),
			}

			tempTopStudents[classID] = append(tempTopStudents[classID], studentData)
		}

		// Добавляем топ-5 учеников для каждого класса
		for classID, students := range tempTopStudents {
			if classData, exists := classes[classID]; exists {
				limit := 5
				if len(students) < limit {
					limit = len(students)
				}
				classData["top_students"] = students[:limit]
			}
		}
	}

	// Общая сводка по всем классам
	overallStats := map[string]interface{}{
		"total_classes":         0,
		"total_students":        0,
		"total_attempts":        0,
		"total_correct":         0,
		"overall_accuracy":      0.0,
		"most_active_class":     nil,
		"best_performing_class": nil,
		// Дублирующие поля для шаблона
		"student_count":    0,
		"correct_attempts": 0,
		"accuracy_percent": 0.0,
		"top_students":     []map[string]interface{}{},
	}

	var totalStudents, totalAttempts, totalCorrect int
	var mostActiveClass map[string]interface{}
	var bestPerformingClass map[string]interface{}
	maxAttempts := 0
	maxAccuracy := 0.0
	var allTopStudents []map[string]interface{}

	for _, classData := range classes {
		studentCount, _ := classData["student_count"].(int)
		attempts, _ := classData["total_attempts"].(int)
		correct, _ := classData["correct_attempts"].(int)
		accuracy, _ := classData["accuracy_percent"].(float64)

		totalStudents += studentCount
		totalAttempts += attempts
		totalCorrect += correct

		if attempts > maxAttempts {
			maxAttempts = attempts
			mostActiveClass = map[string]interface{}{
				"class_id":   classData["class_id"],
				"class_name": classData["class_name"],
				"attempts":   attempts,
			}
		}

		if attempts > 0 && accuracy > maxAccuracy {
			maxAccuracy = accuracy
			bestPerformingClass = map[string]interface{}{
				"class_id":   classData["class_id"],
				"class_name": classData["class_name"],
				"accuracy":   accuracy,
			}
		}

		// Собираем топ учеников из всех классов
		if topStudents, ok := classData["top_students"].([]map[string]interface{}); ok && len(topStudents) > 0 {
			allTopStudents = append(allTopStudents, topStudents...)
		}
	}

	overallStats["total_classes"] = len(classIDs)
	overallStats["total_students"] = totalStudents
	overallStats["total_attempts"] = totalAttempts
	overallStats["total_correct"] = totalCorrect
	overallStats["most_active_class"] = mostActiveClass
	overallStats["best_performing_class"] = bestPerformingClass
	overallStats["student_count"] = totalStudents
	overallStats["correct_attempts"] = totalCorrect

	if totalAttempts > 0 {
		accuracy := float64(totalCorrect) / float64(totalAttempts) * 100
		overallStats["overall_accuracy"] = accuracy
		overallStats["accuracy_percent"] = accuracy
	}

	if len(allTopStudents) > 0 {
		sort.Slice(allTopStudents, func(i, j int) bool {
			accI, _ := allTopStudents[i]["accuracy"].(float64)
			accJ, _ := allTopStudents[j]["accuracy"].(float64)
			return accI > accJ
		})

		limit := 10
		if len(allTopStudents) < limit {
			limit = len(allTopStudents)
		}
		overallStats["top_students"] = allTopStudents[:limit]
	}

	// Добавляем общую статистику в результат
	result := map[int]map[string]interface{}{
		0: overallStats,
	}

	// Добавляем статистику по классам
	for id, data := range classes {
		result[id] = data
	}

	return result, nil
}

// Получить статистику ученика
func (r *TeacherRepository) GetStudentStatistics(studentID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Информация об ученике
	studentQuery := `
		SELECT u.username, u.fullname, u.created_at
		FROM users u
		WHERE u.id = $1
	`

	var username, fullname, createdAt string
	err := r.db.QueryRow(studentQuery, studentID).Scan(&username, &fullname, &createdAt)
	if err != nil {
		return nil, err
	}

	stats["student_info"] = map[string]interface{}{
		"id":         studentID,
		"username":   username,
		"fullname":   fullname,
		"created_at": createdAt,
	}

	// Общая статистика
	overallQuery := `
		SELECT 
			COUNT(*) as total_attempts,
			COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0) as correct_attempts
		FROM attempts
		WHERE user_id = $1
	`

	var totalAttempts, correctAttempts int
	err = r.db.QueryRow(overallQuery, studentID).Scan(&totalAttempts, &correctAttempts)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	stats["overall"] = map[string]interface{}{
		"total_attempts":   totalAttempts,
		"correct_attempts": correctAttempts,
		"accuracy": func() float64 {
			if totalAttempts > 0 {
				return float64(correctAttempts) / float64(totalAttempts) * 100
			}
			return 0
		}(),
	}

	// Статистика по типам уравнений
	typeStatsQuery := `
		SELECT 
			et.id as type_id,
			et.name as type_name,
			et.class,
			COUNT(a.id) as attempts,
			SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) as correct,
			MAX(a.created_at) as last_attempt
		FROM equation_types et
		LEFT JOIN attempts a ON et.id = a.equation_type_id AND a.user_id = $1
		GROUP BY et.id, et.name, et.class
		ORDER BY et.class, et.name
	`

	rows, err := r.db.Query(typeStatsQuery, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var typeStats []map[string]interface{}
	for rows.Next() {
		var typeID, class, attempts, correct int
		var typeName string
		var lastAttempt sql.NullTime

		if err := rows.Scan(&typeID, &typeName, &class, &attempts, &correct, &lastAttempt); err != nil {
			continue
		}

		typeStats = append(typeStats, map[string]interface{}{
			"type_id":   typeID,
			"type_name": typeName,
			"class":     class,
			"attempts":  attempts,
			"correct":   correct,
			"accuracy": func() float64 {
				if attempts > 0 {
					return float64(correct) / float64(attempts) * 100
				}
				return 0
			}(),
			"last_attempt": func() string {
				if lastAttempt.Valid {
					return lastAttempt.Time.Format("02.01.2006 15:04")
				}
				return "Не решал"
			}(),
		})
	}

	stats["type_statistics"] = typeStats

	// Последние попытки
	recentAttemptsQuery := `
		SELECT 
			a.id,
			a.equation_text,
			a.correct_answer,
			a.user_answer,
			a.is_correct,
			a.created_at,
			et.name as type_name
		FROM attempts a
		JOIN equation_types et ON a.equation_type_id = et.id
		WHERE a.user_id = $1
		ORDER BY a.created_at DESC
		LIMIT 10
	`

	attemptRows, err := r.db.Query(recentAttemptsQuery, studentID)
	if err == nil {
		defer attemptRows.Close()

		var recentAttempts []map[string]interface{}
		for attemptRows.Next() {
			var id int
			var equationText, correctAnswer, userAnswer, typeName string
			var isCorrect bool
			var createdAt time.Time

			if err := attemptRows.Scan(&id, &equationText, &correctAnswer, &userAnswer, &isCorrect, &createdAt, &typeName); err != nil {
				continue
			}

			recentAttempts = append(recentAttempts, map[string]interface{}{
				"id":             id,
				"equation_text":  equationText,
				"correct_answer": correctAnswer,
				"user_answer":    userAnswer,
				"is_correct":     isCorrect,
				"created_at":     createdAt.Format("02.01 15:04"),
				"type_name":      typeName,
			})
		}
		stats["recent_attempts"] = recentAttempts
	}

	return stats, nil
}

// Получить попытки ученика по типу уравнения
func (r *TeacherRepository) GetStudentAttemptsByType(studentID, typeID int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			a.id,
			a.equation_text,
			a.correct_answer,
			a.user_answer,
			a.is_correct,
			a.created_at
		FROM attempts a
		WHERE a.user_id = $1 AND a.equation_type_id = $2
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.Query(query, studentID, typeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []map[string]interface{}
	for rows.Next() {
		var id int
		var equationText, correctAnswer, userAnswer string
		var isCorrect bool
		var createdAt time.Time

		if err := rows.Scan(&id, &equationText, &correctAnswer, &userAnswer, &isCorrect, &createdAt); err != nil {
			continue
		}

		attempts = append(attempts, map[string]interface{}{
			"id":             id,
			"equation_text":  equationText,
			"correct_answer": correctAnswer,
			"user_answer":    userAnswer,
			"is_correct":     isCorrect,
			"created_at":     createdAt.Format("02.01.2006 15:04"),
			"status": func() string {
				if isCorrect {
					return "✅ Правильно"
				}
				return "❌ Неправильно"
			}(),
			"status_class": func() string {
				if isCorrect {
					return "correct"
				}
				return "incorrect"
			}(),
		})
	}

	return attempts, nil
}

// Получить подробную статистику по классу
func (r *TeacherRepository) GetDetailedClassStatistics(classID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	classStats, err := r.GetClassStatistics(classID)
	if err != nil {
		return nil, err
	}

	students, err := r.GetClassStudentsWithStats(classID)
	if err != nil {
		return nil, err
	}

	// Получаем детальную статистику по типам уравнений
	detailedTypeStats, err := r.GetClassTypeStatistics(classID)
	if err != nil {
		return nil, err
	}

	// Получаем таблицу успеваемости (каждый ученик по каждому типу)
	performanceTable, err := r.GetClassPerformanceTable(classID)
	if err != nil {
		return nil, err
	}

	stats["overall_stats"] = classStats
	stats["students"] = students
	stats["detailed_type_stats"] = detailedTypeStats
	stats["performance_table"] = performanceTable

	return stats, nil
}

// Получить учеников класса со статистикой
func (r *TeacherRepository) GetClassStudentsWithStats(classID int) ([]map[string]interface{}, error) {
	query := `
        SELECT 
            u.id,
            u.username,
            u.fullname,
            COALESCE(COUNT(a.id), 0) as total_attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_attempts,
            MAX(a.created_at) as last_activity
        FROM users u
        JOIN student_classes sc ON u.id = sc.student_id
        LEFT JOIN attempts a ON u.id = a.user_id
		JOIN roles r ON u.role_id = r.id
        WHERE sc.class_id = $1 AND r.name = 'student'
        GROUP BY u.id, u.username, u.fullname
        ORDER BY u.fullname
    `

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []map[string]interface{}
	for rows.Next() {
		var studentID, totalAttempts, correctAttempts int
		var username, fullname string
		var lastActivity sql.NullTime

		if err := rows.Scan(&studentID, &username, &fullname, &totalAttempts, &correctAttempts, &lastActivity); err != nil {
			continue
		}

		accuracy := 0.0
		if totalAttempts > 0 {
			accuracy = float64(correctAttempts) / float64(totalAttempts) * 100
		}

		students = append(students, map[string]interface{}{
			"id":               studentID,
			"username":         username,
			"fullname":         fullname,
			"total_attempts":   totalAttempts,
			"correct_attempts": correctAttempts,
			"accuracy":         accuracy,
			"last_activity": func() string {
				if lastActivity.Valid {
					return lastActivity.Time.Format("02.01.2006 15:04")
				}
				return "Нет активности"
			}(),
		})
	}

	return students, nil
}

// Получить детальную статистику по типам уравнений для класса
func (r *TeacherRepository) GetClassTypeStatistics(classID int) ([]map[string]interface{}, error) {
	query := `
        WITH class_attempts AS (
            SELECT 
                et.id as type_id,
                et.name as type_name,
                et.class as type_class,
                COUNT(a.id) as total_attempts,
                COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_attempts,
                COUNT(DISTINCT a.user_id) as students_attempted
            FROM equation_types et
            LEFT JOIN attempts a ON et.id = a.equation_type_id
            LEFT JOIN users u ON a.user_id = u.id
            LEFT JOIN student_classes sc ON u.id = sc.student_id AND sc.class_id = $1
            GROUP BY et.id, et.name, et.class
        ),
        class_students AS (
            SELECT COUNT(DISTINCT student_id) as total_students
            FROM student_classes
            WHERE class_id = $1
        )
        SELECT 
            ca.type_id,
            ca.type_name,
            ca.type_class,
            COALESCE(ca.total_attempts, 0) as total_attempts,
            COALESCE(ca.correct_attempts, 0) as correct_attempts,
            COALESCE(ca.students_attempted, 0) as students_attempted,
            COALESCE(cs.total_students, 0) as total_students,
            CASE 
                WHEN COALESCE(ca.total_attempts, 0) > 0 
                THEN ROUND(ca.correct_attempts::DECIMAL / ca.total_attempts * 100, 1)
                ELSE 0 
            END as accuracy_percent,
            CASE 
                WHEN COALESCE(cs.total_students, 0) > 0 
                THEN ROUND(COALESCE(ca.students_attempted, 0)::DECIMAL / cs.total_students * 100, 1)
                ELSE 0 
            END as coverage_percent
        FROM class_attempts ca
        CROSS JOIN class_students cs
        ORDER BY ca.type_class, ca.total_attempts DESC
    `

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var typeStats []map[string]interface{}
	for rows.Next() {
		var typeID, totalAttempts, correctAttempts, studentsAttempted, totalStudents int
		var typeName string
		var typeClass, accuracyPercent, coveragePercent float64

		if err := rows.Scan(&typeID, &typeName, &typeClass, &totalAttempts, &correctAttempts,
			&studentsAttempted, &totalStudents, &accuracyPercent, &coveragePercent); err != nil {
			continue
		}

		// Получаем статистику по ученикам для этого типа
		studentStats, _ := r.GetTypeStudentStats(classID, typeID)

		typeStats = append(typeStats, map[string]interface{}{
			"type_id":            typeID,
			"type_name":          typeName,
			"type_class":         int(typeClass),
			"total_attempts":     totalAttempts,
			"correct_attempts":   correctAttempts,
			"accuracy_percent":   accuracyPercent,
			"students_attempted": studentsAttempted,
			"total_students":     totalStudents,
			"coverage_percent":   coveragePercent,
			"student_stats":      studentStats,
		})
	}

	return typeStats, nil
}

// Получить статистику учеников по типу уравнения
func (r *TeacherRepository) GetTypeStudentStats(classID, typeID int) ([]map[string]interface{}, error) {
	query := `
        SELECT 
            u.id,
            u.fullname,
            COALESCE(COUNT(a.id), 0) as attempts,
            COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct,
            MAX(a.created_at) as last_attempt
        FROM users u
        JOIN student_classes sc ON u.id = sc.student_id
        LEFT JOIN attempts a ON u.id = a.user_id AND a.equation_type_id = $2
        WHERE sc.class_id = $1
        GROUP BY u.id, u.fullname
        ORDER BY u.fullname
    `

	rows, err := r.db.Query(query, classID, typeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentStats []map[string]interface{}
	for rows.Next() {
		var studentID, attempts, correct int
		var fullname string
		var lastAttempt sql.NullTime

		if err := rows.Scan(&studentID, &fullname, &attempts, &correct, &lastAttempt); err != nil {
			continue
		}

		accuracy := 0.0
		if attempts > 0 {
			accuracy = float64(correct) / float64(attempts) * 100
		}

		studentStats = append(studentStats, map[string]interface{}{
			"student_id": studentID,
			"fullname":   fullname,
			"attempts":   attempts,
			"correct":    correct,
			"accuracy":   accuracy,
			"last_attempt": func() string {
				if lastAttempt.Valid {
					return lastAttempt.Time.Format("02.01")
				}
				return "-"
			}(),
		})
	}

	return studentStats, nil
}

// Получить таблицу успеваемости (ученики × типы уравнений)
func (r *TeacherRepository) GetClassPerformanceTable(classID int) ([]map[string]interface{}, error) {
	query := `
        WITH student_type_stats AS (
            SELECT 
                u.id as student_id,
                u.fullname,
                et.id as type_id,
                et.name as type_name,
                COALESCE(COUNT(a.id), 0) as attempts,
                COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct
            FROM users u
            JOIN student_classes sc ON u.id = sc.student_id
            CROSS JOIN equation_types et
            LEFT JOIN attempts a ON u.id = a.user_id AND a.equation_type_id = et.id
            WHERE sc.class_id = $1
            GROUP BY u.id, u.fullname, et.id, et.name
        )
        SELECT 
            student_id,
            fullname,
            type_id,
            type_name,
            attempts,
            correct,
            CASE 
                WHEN attempts = 0 THEN '❌ Не решал'
                WHEN attempts > 0 AND correct = 0 THEN '🔴 Нет верных'
                WHEN correct::DECIMAL / NULLIF(attempts, 0) >= 0.8 THEN '🟢 Отлично'
                WHEN correct::DECIMAL / NULLIF(attempts, 0) >= 0.6 THEN '🟡 Хорошо'
                WHEN correct::DECIMAL / NULLIF(attempts, 0) >= 0.4 THEN '🟠 Удовлет.'
                ELSE '🔴 Слабо'
            END as performance
        FROM student_type_stats
        ORDER BY fullname, type_name
    `

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var performanceTable []map[string]interface{}
	for rows.Next() {
		var studentID, typeID, attempts, correct int
		var fullname, typeName, performance string

		if err := rows.Scan(&studentID, &fullname, &typeID, &typeName, &attempts, &correct, &performance); err != nil {
			continue
		}

		accuracy := 0.0
		if attempts > 0 {
			accuracy = float64(correct) / float64(attempts) * 100
		}

		performanceTable = append(performanceTable, map[string]interface{}{
			"student_id":  studentID,
			"fullname":    fullname,
			"type_id":     typeID,
			"type_name":   typeName,
			"attempts":    attempts,
			"correct":     correct,
			"accuracy":    accuracy,
			"performance": performance,
			"cell_class": func() string {
				switch {
				case attempts == 0:
					return "not-attempted"
				case attempts > 0 && correct == 0:
					return "no-correct"
				case accuracy >= 80:
					return "excellent"
				case accuracy >= 60:
					return "good"
				case accuracy >= 40:
					return "average"
				default:
					return "poor"
				}
			}(),
		})
	}

	return performanceTable, nil
}

// Получить имя класса по ID
func (r *TeacherRepository) GetClassName(classID int) (string, error) {
	query := `SELECT name FROM classes WHERE id = $1`
	var name string
	err := r.db.QueryRow(query, classID).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

type SessionResult struct {
	Hour    int `json:"hour"`
	Total   int `json:"total"`
	Correct int `json:"correct"`
}

type StudentResult struct {
	Date    string
	Results []SessionResult
}

type DailyResult struct {
	Results  []SessionDisplay `json:"results"`
	DayScore int              `json:"dayScore"`
}

type SessionDisplay struct {
	Correct  int    `json:"correct"`
	Total    int    `json:"total"`
	CSSClass string `json:"cssClass"`
}

type DateInfo struct {
	Date    string `json:"date"`
	Weekday string `json:"weekday"`
}

type StudentInfo struct {
	ID            int           `json:"id"`
	FullName      string        `json:"fullName"`
	DailyResults  []DailyResult `json:"dailyResults"`
	AverageScore  float64       `json:"averageScore"`
	TotalAttempts int           `json:"totalAttempts"`
	IsActive      bool          `json:"isActive"`
}

type ClassStats struct {
	AvgDailyAttempts int     `json:"avgDailyAttempts"`
	AvgScore         float64 `json:"avgScore"`
	TotalSessions    int     `json:"totalSessions"`
	PerfectSessions  int     `json:"perfectSessions"`
	TotalStudents    int     `json:"totalStudents"`
	TotalCorrect     int     `json:"totalCorrect"`
	TotalAttempts    int     `json:"totalAttempts"`
	OverallAccuracy  float64 `json:"overallAccuracy"`
}

type DailyClassResults struct {
	WeekStart string        `json:"weekStart"`
	WeekEnd   string        `json:"weekEnd"`
	Dates     []DateInfo    `json:"dates"`
	Students  []StudentInfo `json:"students"`
	Stats     ClassStats    `json:"stats"`
}

// Функция для получения понедельника недели
func getMondayOfWeek(t time.Time) time.Time {
	// Go: Sunday = 0, Monday = 1, ..., Saturday = 6
	weekday := int(t.Weekday())
	// Преобразуем к нашей нумерации: Monday = 0, Sunday = 6
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysSinceMonday := weekday - 1
	return t.AddDate(0, 0, -daysSinceMonday)
}

// GetDailyClassResults получает ежедневные результаты класса по 10 примерам за сессию
func (r *TeacherRepository) GetDailyClassResults(classID int, weeksOffset int) (*DailyClassResults, error) {
	// Определяем даты для недели, начиная с понедельника
	now := time.Now()

	// Находим понедельник текущей недели
	monday := getMondayOfWeek(now)
	startDate := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 6) // Воскресенье

	if weeksOffset != 0 {
		startDate = startDate.AddDate(0, 0, weeksOffset*7)
		endDate = endDate.AddDate(0, 0, weeksOffset*7)
	}

	// Получаем учеников класса
	students, err := r.GetClassStudents(classID)
	if err != nil {
		return nil, err
	}

	// Инициализируем структуру результата
	result := &DailyClassResults{
		WeekStart: startDate.Format("02.01"),
		WeekEnd:   endDate.Format("02.01"),
		Dates:     make([]DateInfo, 0, 7),
		Students:  make([]StudentInfo, 0, len(students)),
		Stats:     ClassStats{},
	}

	// Генерируем даты для заголовков таблицы
	weekdays := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		result.Dates = append(result.Dates, DateInfo{
			Date:    currentDate.Format("02.01"),
			Weekday: weekdays[i],
		})
	}

	// Получаем результаты для каждого ученика
	var totalSessions, perfectSessions, totalScore, totalAttempts int

	for _, student := range students {
		studentResults, err := r.GetStudentDailyResults(student.ID, startDate, endDate)
		if err != nil {
			continue
		}

		var studentTotalScore, studentSessions int
		dailyResults := make([]DailyResult, 0, 7)

		// Заполняем результаты по дням
		for i := 0; i < 7; i++ {
			currentDate := startDate.AddDate(0, 0, i)
			dateStr := currentDate.Format("2006-01-02")

			dailyResult := DailyResult{
				Results: make([]SessionDisplay, 0),
			}

			if sessions, exists := studentResults[dateStr]; exists && len(sessions) > 0 {
				var sessionDisplays []SessionDisplay
				var dayScore int

				for _, session := range sessions {
					cssClass := getCSSClassForResult(session.Correct)

					if session.Correct == 10 {
						perfectSessions++
					}

					sessionDisplays = append(sessionDisplays, SessionDisplay{
						Correct:  session.Correct,
						Total:    session.Total,
						CSSClass: cssClass,
					})

					dayScore += session.Correct
					studentSessions++
					totalSessions++
				}

				dailyResult.Results = sessionDisplays
				dailyResult.DayScore = dayScore
				studentTotalScore += dayScore
			}

			dailyResults = append(dailyResults, dailyResult)
		}

		// Рассчитываем средний балл ученика
		avgScore := 0.0
		if studentSessions > 0 {
			avgScore = float64(studentTotalScore) / float64(studentSessions)
		}

		totalScore += studentTotalScore
		totalAttempts += studentSessions * 10

		// Добавляем ученика в результаты
		result.Students = append(result.Students, StudentInfo{
			ID:            student.ID,
			FullName:      student.FullName,
			DailyResults:  dailyResults,
			AverageScore:  avgScore,
			TotalAttempts: studentSessions,
			IsActive:      studentSessions > 0,
		})
	}

	// Рассчитываем статистику
	avgDailyAttempts := 0
	avgScore := 0.0

	if len(students) > 0 {
		avgDailyAttempts = totalSessions / len(students)
	}

	if totalSessions > 0 {
		avgScore = float64(totalScore) / float64(totalSessions)
	}

	overallAccuracy := 0.0
	if totalAttempts > 0 {
		overallAccuracy = float64(totalScore) / float64(totalAttempts) * 100
	}

	result.Stats = ClassStats{
		AvgDailyAttempts: avgDailyAttempts,
		AvgScore:         avgScore,
		TotalSessions:    totalSessions,
		PerfectSessions:  perfectSessions,
		TotalStudents:    len(students),
		TotalCorrect:     totalScore,
		TotalAttempts:    totalAttempts,
		OverallAccuracy:  overallAccuracy,
	}

	return result, nil
}

// Вспомогательная функция для определения CSS класса
func getCSSClassForResult(correct int) string {
	switch {
	case correct == 10:
		return "perfect-result"
	case correct >= 8:
		return "good-result"
	case correct >= 6:
		return "average-result"
	default:
		return "poor-result"
	}
}

func (r *TeacherRepository) GetStudentDailyResults(studentID int, startDate, endDate time.Time) (map[string][]SessionResult, error) {
	// Запрос для получения результатов по 10 примеров за сессию
	query := `
        WITH session_groups AS (
            SELECT 
                DATE(a.created_at) as session_date,
                EXTRACT(HOUR FROM a.created_at) as session_hour,
                COUNT(*) as examples_in_session,
                SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) as correct_in_session
            FROM attempts a
            WHERE a.user_id = $1
              AND DATE(a.created_at) BETWEEN $2::date AND $3::date
            GROUP BY DATE(a.created_at), EXTRACT(HOUR FROM a.created_at)
            HAVING COUNT(*) = 10
        )
        SELECT 
            session_date,
            session_hour,
            examples_in_session,
            correct_in_session
        FROM session_groups
        ORDER BY session_date, session_hour
    `

	// Правильное форматирование дат
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	rows, err := r.db.Query(query, studentID, startDateStr, endDateStr)
	if err != nil {
		log.Printf("Query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Типизированная мапа
	results := make(map[string][]SessionResult)

	for rows.Next() {
		var sessionDate time.Time
		var sessionHour int
		var examples, correct int

		if err := rows.Scan(&sessionDate, &sessionHour, &examples, &correct); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		// Добавляем типизированную структуру
		results[sessionDate.Format("2006-01-02")] = append(results[sessionDate.Format("2006-01-02")], SessionResult{
			Hour:    sessionHour,
			Total:   examples,
			Correct: correct,
		})
	}

	// Проверяем ошибки итерации
	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return nil, err
	}

	log.Printf("Query params: studentID=%d, startDate=%s, endDate=%s",
		studentID, startDateStr, endDateStr)
	log.Printf("Results: %v", results)
	return results, nil
}
