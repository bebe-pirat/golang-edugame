package repository

import (
	"database/sql"
	"edugame/internal/generator"
)

type TypeRepository struct {
	db *sql.DB
}

func NewTypeRepository(db *sql.DB) *TypeRepository {
	return &TypeRepository{db: db}
}

// GetAll получает все типы уравнений с диапазонами операндов
func (r *TypeRepository) GetAll() ([]generator.EquationType, error) {
	query := `
        SELECT id, class, name, description, operation, num_operands, 
               no_remainder, COALESCE(result_max, -1), is_available
        FROM equation_types
        ORDER BY class, name
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]generator.EquationType, 0)

	for rows.Next() {
		t := generator.EquationType{}
		err := rows.Scan(
			&t.ID,
			&t.Class,
			&t.Name,
			&t.Description,
			&t.Operation,
			&t.NumOperands,
			&t.NoRemainder,
			&t.ResultMax,
			&t.IsAvailable,
		)

		if err != nil {
			return types, err
		}

		// Загружаем диапазоны операндов для этого типа уравнения
		operands, err := r.getOperandRanges(t.ID)
		if err != nil {
			return types, err
		}
		t.Operands = operands

		types = append(types, t)
	}

	return types, nil
}

// getOperandRanges получает диапазоны операндов для типа уравнения
func (r *TypeRepository) getOperandRanges(equationTypeID int) ([]generator.OperandRange, error) {
	query := `
		SELECT operand_order, min_value, max_value
		FROM operand_ranges
		WHERE equation_type_id = $1
		ORDER BY operand_order
	`

	rows, err := r.db.Query(query, equationTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	operands := make([]generator.OperandRange, 0)
	for rows.Next() {
		var op generator.OperandRange
		err := rows.Scan(&op.Order, &op.MinValue, &op.MaxValue)
		if err != nil {
			return nil, err
		}
		operands = append(operands, op)
	}

	return operands, nil
}

// получить полный список типов уравнений для определенного класса
func (r *TypeRepository) GetListTypes(class int) ([]generator.EquationType, error) {
	query := `
        SELECT id, class, name, description, operation, num_operands, 
               no_remainder, COALESCE(result_max, -1), is_available
        FROM equation_types
        WHERE class = $1 
    `

	rows, err := r.db.Query(query, class)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]generator.EquationType, 0)

	for rows.Next() {
		t := generator.EquationType{}
		err := rows.Scan(
			&t.ID,
			&t.Class,
			&t.Name,
			&t.Description,
			&t.Operation,
			&t.NumOperands,
			&t.NoRemainder,
			&t.ResultMax,
			&t.IsAvailable,
		)

		if err != nil {
			return types, err
		}

		// Загружаем диапазоны операндов для этого типа уравнения
		operands, err := r.getOperandRanges(t.ID)
		if err != nil {
			return types, err
		}
		t.Operands = operands

		types = append(types, t)
	}

	return types, nil
}

func (r *TypeRepository) GetTypeById(id int) (generator.EquationType, error) {
	var t generator.EquationType

	err := r.db.QueryRow(`
		SELECT id, class, name, description, operation, num_operands, 
		       no_remainder, COALESCE(result_max, -1), is_available
		FROM equation_types
		WHERE id = $1
	`, id).Scan(
		&t.ID,
		&t.Class,
		&t.Name,
		&t.Description,
		&t.Operation,
		&t.NumOperands,
		&t.NoRemainder,
		&t.ResultMax,
		&t.IsAvailable,
	)

	if err != nil {
		return t, err
	}

	// Загружаем диапазоны операндов для этого типа уравнения
	operands, err := r.getOperandRanges(id)
	if err != nil {
		return t, err
	}
	t.Operands = operands

	return t, nil
}

// Create создает новый тип уравнения
func (r *TypeRepository) Create(et generator.EquationType) (*generator.EquationType, error) {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Вставляем тип уравнения
	query := `
		INSERT INTO equation_types (
			class, name, description, operation, num_operands,
			no_remainder, result_max, is_available
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, class, name, description, operation, num_operands,
		          no_remainder, result_max, is_available
	`

	var newEt generator.EquationType
	err = tx.QueryRow(query,
		et.Class, et.Name, et.Description, et.Operation, et.NumOperands,
		et.NoRemainder, nullIfMinusOne(et.ResultMax), true,
	).Scan(
		&newEt.ID,
		&newEt.Class,
		&newEt.Name,
		&newEt.Description,
		&newEt.Operation,
		&newEt.NumOperands,
		&newEt.NoRemainder,
		&newEt.ResultMax,
		&newEt.IsAvailable,
	)

	if err != nil {
		return nil, err
	}

	// Вставляем диапазоны операндов
	for _, op := range et.Operands {
		opQuery := `
			INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(opQuery, newEt.ID, op.Order, op.MinValue, op.MaxValue)
		if err != nil {
			return nil, err
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Загружаем полные данные с диапазонами операндов
	newEt.Operands = et.Operands

	return &newEt, nil
}

// Update обновляет тип уравнения
func (r *TypeRepository) Update(et generator.EquationType) (*generator.EquationType, error) {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Обновляем тип уравнения
	query := `
		UPDATE equation_types SET
			class = $1, name = $2, description = $3, operation = $4, num_operands = $5,
			no_remainder = $6, result_max = $7, is_available = $8
		WHERE id = $9
		RETURNING id, class, name, description, operation, num_operands,
		          no_remainder, result_max, is_available
	`

	var newEt generator.EquationType
	err = tx.QueryRow(query,
		et.Class, et.Name, et.Description, et.Operation, et.NumOperands,
		et.NoRemainder, nullIfMinusOne(et.ResultMax), et.IsAvailable, et.ID,
	).Scan(
		&newEt.ID,
		&newEt.Class,
		&newEt.Name,
		&newEt.Description,
		&newEt.Operation,
		&newEt.NumOperands,
		&newEt.NoRemainder,
		&newEt.ResultMax,
		&newEt.IsAvailable,
	)

	if err != nil {
		return nil, err
	}

	// Удаляем старые диапазоны операндов
	_, err = tx.Exec(`DELETE FROM operand_ranges WHERE equation_type_id = $1`, et.ID)
	if err != nil {
		return nil, err
	}

	// Вставляем новые диапазоны операндов
	for _, op := range et.Operands {
		opQuery := `
			INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(opQuery, et.ID, op.Order, op.MinValue, op.MaxValue)
		if err != nil {
			return nil, err
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Загружаем полные данные с диапазонами операндов
	newEt.Operands = et.Operands

	return &newEt, nil
}

// Delete удаляет тип уравнения
func (r *TypeRepository) Delete(id int) error {
	query := `DELETE FROM equation_types WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// ToggleAvailability переключает доступность типа уравнения
func (r *TypeRepository) ToggleAvailability(id int) error {
	query := `UPDATE equation_types SET is_available = NOT is_available WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func nullIfMinusOne(val int) sql.NullInt64 {
	if val == -1 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(val), Valid: true}
}
