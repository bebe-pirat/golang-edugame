package handler

import (
	"edugame/internal/repository"
	"edugame/internal/entity"
	"edugame/internal/generator"
	"html/template"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	schoolRepo    *repository.SchoolRepository
	classRepo     *repository.ClassRepository
	userRepo      *repository.UserRepository
	roleRepo      *repository.RoleRepository
	typeRepo      *repository.TypeRepository
	tmpl          *template.Template
}

func NewAdminHandler(
	schoolRepo *repository.SchoolRepository,
	classRepo *repository.ClassRepository,
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	typeRepo *repository.TypeRepository,
) *AdminHandler {
	tmpl := template.Must(template.ParseFiles(
		"internal/templates/admin/dashboard.html",
		"internal/templates/admin/schools.html",
		"internal/templates/admin/school_form.html",
		"internal/templates/admin/classes.html",
		"internal/templates/admin/class_form.html",
		"internal/templates/admin/users.html",
		"internal/templates/admin/user_form.html",
		"internal/templates/admin/equation_types.html",
		"internal/templates/admin/equation_type_form.html",
	))

	return &AdminHandler{
		schoolRepo: schoolRepo,
		classRepo:  classRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		typeRepo:   typeRepo,
		tmpl:       tmpl,
	}
}

// Dashboard - главная страница админки
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	schools, _ := h.schoolRepo.GetAll()
	classes, _ := h.classRepo.GetAll()
	users, _ := h.userRepo.GetAllUsers()
	types, _ := h.typeRepo.GetAll()

	data := map[string]interface{}{
		"Title":        "Админ-панель",
		"SchoolsCount": len(schools),
		"ClassesCount": len(classes),
		"UsersCount":   len(users),
		"TypesCount":   len(types),
	}

	h.tmpl.ExecuteTemplate(w, "admin/dashboard.html", data)
}

// ============= ШКОЛЫ =============

// Schools - список всех школ
func (h *AdminHandler) Schools(w http.ResponseWriter, r *http.Request) {
	schools, err := h.schoolRepo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка получения школ", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   "Управление школами",
		"Schools": schools,
	}

	h.tmpl.ExecuteTemplate(w, "admin/schools.html", data)
}

// SchoolForm - форма создания/редактирования школы
func (h *AdminHandler) SchoolForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	
	data := map[string]interface{}{
		"Title": "Новая школа",
		"School": nil,
	}

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			school, err := h.schoolRepo.GetByID(id)
			if err == nil {
				data["School"] = school
				data["Title"] = "Редактирование школы"
			}
		}
	}

	h.tmpl.ExecuteTemplate(w, "admin/school_form.html", data)
}

// SchoolCreate - создание школы
func (h *AdminHandler) SchoolCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/schools", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	address := r.FormValue("address")
	phone := r.FormValue("phone")
	email := r.FormValue("email")

	_, err := h.schoolRepo.Create(name, address, phone, email)
	if err != nil {
		http.Error(w, "Ошибка создания школы", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/schools", http.StatusSeeOther)
}

// SchoolUpdate - обновление школы
func (h *AdminHandler) SchoolUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/schools", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	address := r.FormValue("address")
	phone := r.FormValue("phone")
	email := r.FormValue("email")

	_, err = h.schoolRepo.Update(id, name, address, phone, email)
	if err != nil {
		http.Error(w, "Ошибка обновления школы", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/schools", http.StatusSeeOther)
}

// SchoolDelete - удаление школы
func (h *AdminHandler) SchoolDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/schools", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = h.schoolRepo.Delete(id)
	if err != nil {
		http.Error(w, "Ошибка удаления школы", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/schools", http.StatusSeeOther)
}

// ============= КЛАССЫ =============

// Classes - список всех классов
func (h *AdminHandler) Classes(w http.ResponseWriter, r *http.Request) {
	classes, err := h.classRepo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка получения классов", http.StatusInternalServerError)
		return
	}

	schools, _ := h.schoolRepo.GetAll()
	teachers, _ := h.userRepo.GetUserByRoleType("teacher")

	data := map[string]interface{}{
		"Title":    "Управление классами",
		"Classes":  classes,
		"Schools":  schools,
		"Teachers": teachers,
	}

	h.tmpl.ExecuteTemplate(w, "admin/classes.html", data)
}

// ClassForm - форма создания/редактирования класса
func (h *AdminHandler) ClassForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	
	schools, _ := h.schoolRepo.GetAll()
	teachers, _ := h.userRepo.GetUserByRoleType("teacher")

	data := map[string]interface{}{
		"Title":    "Новый класс",
		"Class":    nil,
		"Schools":  schools,
		"Teachers": teachers,
	}

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			class, err := h.classRepo.GetByID(id)
			if err == nil {
				data["Class"] = class
				data["Title"] = "Редактирование класса"
			}
		}
	}

	h.tmpl.ExecuteTemplate(w, "admin/class_form.html", data)
}

// ClassCreate - создание класса
func (h *AdminHandler) ClassCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/classes", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	grade, _ := strconv.Atoi(r.FormValue("grade"))
	teacherID, _ := strconv.Atoi(r.FormValue("teacher_id"))
	
	var schoolID *int
	if schoolIDStr := r.FormValue("school_id"); schoolIDStr != "" {
		id, _ := strconv.Atoi(schoolIDStr)
		schoolID = &id
	}

	_, err := h.classRepo.Create(name, grade, teacherID, schoolID)
	if err != nil {
		http.Error(w, "Ошибка создания класса", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/classes", http.StatusSeeOther)
}

// ClassUpdate - обновление класса
func (h *AdminHandler) ClassUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/classes", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	grade, _ := strconv.Atoi(r.FormValue("grade"))
	teacherID, _ := strconv.Atoi(r.FormValue("teacher_id"))
	
	var schoolID *int
	if schoolIDStr := r.FormValue("school_id"); schoolIDStr != "" {
		id, _ := strconv.Atoi(schoolIDStr)
		schoolID = &id
	}

	_, err = h.classRepo.Update(id, name, grade, teacherID, schoolID)
	if err != nil {
		http.Error(w, "Ошибка обновления класса", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/classes", http.StatusSeeOther)
}

// ClassDelete - удаление класса
func (h *AdminHandler) ClassDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/classes", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = h.classRepo.Delete(id)
	if err != nil {
		http.Error(w, "Ошибка удаления класса", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/classes", http.StatusSeeOther)
}

// ============= ПОЛЬЗОВАТЕЛИ =============

// Users - список всех пользователей
func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	roleFilter := r.URL.Query().Get("role")
	
	var users []entity.User
	var err error
	
	if roleFilter != "" {
		users, err = h.userRepo.GetUserByRoleType(roleFilter)
	} else {
		users, err = h.userRepo.GetAllUsers()
	}
	
	if err != nil {
		http.Error(w, "Ошибка получения пользователей", http.StatusInternalServerError)
		return
	}

	roles, _ := h.roleRepo.GetAll()
	schools, _ := h.schoolRepo.GetAll()

	data := map[string]interface{}{
		"Title":      "Управление пользователями",
		"Users":      users,
		"Roles":      roles,
		"Schools":    schools,
		"RoleFilter": roleFilter,
	}

	h.tmpl.ExecuteTemplate(w, "admin/users.html", data)
}

// UserForm - форма создания/редактирования пользователя
func (h *AdminHandler) UserForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	
	roles, _ := h.roleRepo.GetAll()
	schools, _ := h.schoolRepo.GetAll()
	classes, _ := h.classRepo.GetAll()

	data := map[string]interface{}{
		"Title":   "Новый пользователь",
		"User":    nil,
		"Roles":   roles,
		"Schools": schools,
		"Classes": classes,
	}

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			user, err := h.userRepo.GetByID(id)
			if err == nil {
				data["User"] = user
				data["Title"] = "Редактирование пользователя"
			}
		}
	}

	h.tmpl.ExecuteTemplate(w, "admin/user_form.html", data)
}

// UserCreate - создание пользователя
func (h *AdminHandler) UserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	fullName := r.FormValue("fullname")
	roleName := r.FormValue("role_id")
	
	var classID *int
	if classIDStr := r.FormValue("class_id"); classIDStr != "" {
		id, _ := strconv.Atoi(classIDStr)
		classID = &id
	}

	_, err := h.userRepo.Register(username, password, roleName, fullName, classID)
	if err != nil {
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// UserUpdate - обновление пользователя
func (h *AdminHandler) UserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	fullName := r.FormValue("fullname")
	email := r.FormValue("email")
	roleID, _ := strconv.Atoi(r.FormValue("role_id"))
	
	var schoolID *int
	if schoolIDStr := r.FormValue("school_id"); schoolIDStr != "" {
		id, _ := strconv.Atoi(schoolIDStr)
		schoolID = &id
	}

	_, err = h.userRepo.UpdateUser(id, username, fullName, email, roleID, schoolID)
	if err != nil {
		http.Error(w, "Ошибка обновления пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// UserDelete - удаление пользователя
func (h *AdminHandler) UserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = h.userRepo.DeleteUser(id)
	if err != nil {
		http.Error(w, "Ошибка удаления пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// ============= ТИПЫ УРАВНЕНИЙ =============

// EquationTypes - список всех типов уравнений
func (h *AdminHandler) EquationTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.typeRepo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка получения типов уравнений", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "Типы уравнений",
		"Types": types,
	}

	h.tmpl.ExecuteTemplate(w, "admin/equation_types.html", data)
}

// EquationTypeForm - форма создания/редактирования типа уравнения
func (h *AdminHandler) EquationTypeForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	
	data := map[string]interface{}{
		"Title": "Новый тип уравнения",
		"Type":  nil,
	}

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			et, err := h.typeRepo.GetTypeById(id)
			if err == nil {
				data["Type"] = et
				data["Title"] = "Редактирование типа уравнения"
			}
		}
	}

	h.tmpl.ExecuteTemplate(w, "admin/equation_type_form.html", data)
}

// EquationTypeCreate - создание типа уравнения
func (h *AdminHandler) EquationTypeCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
		return
	}

	// Парсинг данных формы
	class, _ := strconv.Atoi(r.FormValue("class"))
	name := r.FormValue("name")
	description := r.FormValue("description")
	operation := r.FormValue("operation")
	numOperands, _ := strconv.Atoi(r.FormValue("num_operands"))
	
	op1Min, _ := strconv.Atoi(r.FormValue("operand1_min"))
	op1Max, _ := strconv.Atoi(r.FormValue("operand1_max"))
	op2Min, _ := strconv.Atoi(r.FormValue("operand2_min"))
	op2Max, _ := strconv.Atoi(r.FormValue("operand2_max"))
	op3Min, _ := strconv.Atoi(r.FormValue("operand3_min"))
	op3Max, _ := strconv.Atoi(r.FormValue("operand3_max"))
	op4Min, _ := strconv.Atoi(r.FormValue("operand4_min"))
	op4Max, _ := strconv.Atoi(r.FormValue("operand4_max"))
	
	noRemainder := r.FormValue("no_remainder") == "on"
	
	resultMax, _ := strconv.Atoi(r.FormValue("result_max"))
	if resultMax == 0 {
		resultMax = -1
	}

	et := generator.EquationType{
		Class:       class,
		Name:        name,
		Description: description,
		Operation:   operation,
		NumOperands: numOperands,
		Operands: [4][2]int{
			{op1Min, op1Max},
			{op2Min, op2Max},
			{op3Min, op3Max},
			{op4Min, op4Max},
		},
		No_remainder: noRemainder,
		Result_max:   resultMax,
	}

	_, err := h.typeRepo.Create(et)
	if err != nil {
		http.Error(w, "Ошибка создания типа уравнения", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
}

// EquationTypeUpdate - обновление типа уравнения
func (h *AdminHandler) EquationTypeUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	// Парсинг данных формы
	class, _ := strconv.Atoi(r.FormValue("class"))
	name := r.FormValue("name")
	description := r.FormValue("description")
	operation := r.FormValue("operation")
	numOperands, _ := strconv.Atoi(r.FormValue("num_operands"))
	
	op1Min, _ := strconv.Atoi(r.FormValue("operand1_min"))
	op1Max, _ := strconv.Atoi(r.FormValue("operand1_max"))
	op2Min, _ := strconv.Atoi(r.FormValue("operand2_min"))
	op2Max, _ := strconv.Atoi(r.FormValue("operand2_max"))
	op3Min, _ := strconv.Atoi(r.FormValue("operand3_min"))
	op3Max, _ := strconv.Atoi(r.FormValue("operand3_max"))
	op4Min, _ := strconv.Atoi(r.FormValue("operand4_min"))
	op4Max, _ := strconv.Atoi(r.FormValue("operand4_max"))
	
	noRemainder := r.FormValue("no_remainder") == "on"
	
	resultMax, _ := strconv.Atoi(r.FormValue("result_max"))
	if resultMax == 0 {
		resultMax = -1
	}

	et := generator.EquationType{
		ID:          id,
		Class:       class,
		Name:        name,
		Description: description,
		Operation:   operation,
		NumOperands: numOperands,
		Operands: [4][2]int{
			{op1Min, op1Max},
			{op2Min, op2Max},
			{op3Min, op3Max},
			{op4Min, op4Max},
		},
		No_remainder: noRemainder,
		Result_max:   resultMax,
	}

	_, err = h.typeRepo.Update(et)
	if err != nil {
		http.Error(w, "Ошибка обновления типа уравнения", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
}

// EquationTypeDelete - удаление типа уравнения
func (h *AdminHandler) EquationTypeDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = h.typeRepo.Delete(id)
	if err != nil {
		http.Error(w, "Ошибка удаления типа уравнения", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
}

// ToggleEquationTypeAvailability - переключение доступности типа уравнения
func (h *AdminHandler) ToggleEquationTypeAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = h.typeRepo.ToggleAvailability(id)
	if err != nil {
		http.Error(w, "Ошибка переключения доступности", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/equation-types", http.StatusSeeOther)
}
