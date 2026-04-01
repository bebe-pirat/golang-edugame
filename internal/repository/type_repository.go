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

// GetAll получает все типы уравнений
func (r *TypeRepository) GetAll() ([]generator.EquationType, error) {
	query := `
        SELECT id, class, name, description, operation, num_operands, 
               operand1_min, operand1_max, operand2_min, operand2_max, 
               COALESCE(operand3_min, -1), COALESCE(operand3_max, -1), 
               COALESCE(operand4_min, -1), COALESCE(operand4_max, -1), 
               no_remainder, COALESCE(result_max, -1), is_available
        FROM equation_types
        ORDER BY class, name
    `

	types := make([]generator.EquationType, 0)

	rows, err := r.db.Query(query)
	if err != nil {
		return types, err
	}
	defer rows.Close()

	for rows.Next() {
		t := generator.EquationType{}
		var isAvailable bool
		err := rows.Scan(
			&t.ID,
			&t.Class,
			&t.Name,
			&t.Description,
			&t.Operation,
			&t.NumOperands,

			&t.Operands[0][0],
			&t.Operands[0][1],
			&t.Operands[1][0],
			&t.Operands[1][1],
			&t.Operands[2][0],
			&t.Operands[2][1],
			&t.Operands[3][0],
			&t.Operands[3][1],

			&t.No_remainder,
			&t.Result_max,
			&isAvailable,
		)

		if err != nil {
			return types, err
		}

		types = append(types, t)
	}

	return types, nil
}

// получить полный список типов уравнений для определенного класса
func (r *TypeRepository) GetListTypes(class int) ([]generator.EquationType, error) {
	query := `
        SELECT id, class, name, description, operation, num_operands, operand1_min, operand1_max, operand2_min, operand2_max, COALESCE(operand3_min, -1), COALESCE(operand3_max, -1), COALESCE(operand4_min, -1), COALESCE(operand4_max, -1), no_remainder, COALESCE(result_max, -1)
        FROM equation_types
        WHERE class = $1 
    `

	types := make([]generator.EquationType, 0)

	rows, err := r.db.Query(query, class)
	if err != nil {
		return types, err
	}

	defer rows.Close()

	for rows.Next() {
		t := generator.EquationType{}
		err := rows.Scan(
			&t.ID,
			&t.Class,
			&t.Name,
			&t.Description,
			&t.Operation,
			&t.NumOperands,

			&t.Operands[0][0],
			&t.Operands[0][1],
			&t.Operands[1][0],
			&t.Operands[1][1],
			&t.Operands[2][0],
			&t.Operands[2][1],
			&t.Operands[3][0],
			&t.Operands[3][1],

			&t.No_remainder,
			&t.Result_max,
		)

		if err != nil {
			return types, err
		}

		types = append(types, t)
	}

	return types, nil
}

func (r *TypeRepository) GetTypeById(id int) (generator.EquationType, error) {
	var t generator.EquationType

	err := r.db.QueryRow(`
		SELECT id, class, name, description, operation, num_operands, 
		       operand1_min, operand1_max, operand2_min, operand2_max, 
		       COALESCE(operand3_min, -1), COALESCE(operand3_max, -1), 
		       COALESCE(operand4_min, -1), COALESCE(operand4_max, -1), 
		       no_remainder, COALESCE(result_max, -1)
		FROM equation_types
		WHERE id = $1
	`, id).Scan(
		&t.ID,
		&t.Class,
		&t.Name,
		&t.Description,
		&t.Operation,
		&t.NumOperands,

		&t.Operands[0][0],
		&t.Operands[0][1],
		&t.Operands[1][0],
		&t.Operands[1][1],
		&t.Operands[2][0],
		&t.Operands[2][1],
		&t.Operands[3][0],
		&t.Operands[3][1],

		&t.No_remainder,
		&t.Result_max,
	)

	if err != nil {
		return t, err
	}

	return t, nil
}

// Create создает новый тип уравнения
func (r *TypeRepository) Create(et generator.EquationType) (*generator.EquationType, error) {
	query := `
		INSERT INTO equation_types (
			class, name, description, operation, num_operands,
			operand1_min, operand1_max, operand2_min, operand2_max,
			operand3_min, operand3_max, operand4_min, operand4_max,
			no_remainder, result_max, is_available
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, class, name, description, operation, num_operands,
		          operand1_min, operand1_max, operand2_min, operand2_max,
		          operand3_min, operand3_max, operand4_min, operand4_max,
		          no_remainder, result_max
	`

	var newEt generator.EquationType
	err := r.db.QueryRow(query,
		et.Class, et.Name, et.Description, et.Operation, et.NumOperands,
		et.Operands[0][0], et.Operands[0][1], et.Operands[1][0], et.Operands[1][1],
		nullIfMinusOne(et.Operands[2][0]), nullIfMinusOne(et.Operands[2][1]),
		nullIfMinusOne(et.Operands[3][0]), nullIfMinusOne(et.Operands[3][1]),
		et.No_remainder, nullIfMinusOne(et.Result_max), true,
	).Scan(
		&newEt.ID,
		&newEt.Class,
		&newEt.Name,
		&newEt.Description,
		&newEt.Operation,
		&newEt.NumOperands,
		&newEt.Operands[0][0],
		&newEt.Operands[0][1],
		&newEt.Operands[1][0],
		&newEt.Operands[1][1],
		&newEt.Operands[2][0],
		&newEt.Operands[2][1],
		&newEt.Operands[3][0],
		&newEt.Operands[3][1],
		&newEt.No_remainder,
		&newEt.Result_max,
	)

	if err != nil {
		return nil, err
	}

	return &newEt, nil
}

// Update обновляет тип уравнения
func (r *TypeRepository) Update(et generator.EquationType) (*generator.EquationType, error) {
	query := `
		UPDATE equation_types SET
			class = $1, name = $2, description = $3, operation = $4, num_operands = $5,
			operand1_min = $6, operand1_max = $7, operand2_min = $8, operand2_max = $9,
			operand3_min = $10, operand3_max = $11, operand4_min = $12, operand4_max = $13,
			no_remainder = $14, result_max = $15
		WHERE id = $17
		RETURNING id, class, name, description, operation, num_operands,
		          operand1_min, operand1_max, operand2_min, operand2_max,
		          operand3_min, operand3_max, operand4_min, operand4_max,
		          no_remainder, result_max
	`

	var newEt generator.EquationType
	err := r.db.QueryRow(query,
		et.Class, et.Name, et.Description, et.Operation, et.NumOperands,
		et.Operands[0][0], et.Operands[0][1], et.Operands[1][0], et.Operands[1][1],
		nullIfMinusOne(et.Operands[2][0]), nullIfMinusOne(et.Operands[2][1]),
		nullIfMinusOne(et.Operands[3][0]), nullIfMinusOne(et.Operands[3][1]),
		et.No_remainder, nullIfMinusOne(et.Result_max), et.ID,
	).Scan(
		&newEt.ID,
		&newEt.Class,
		&newEt.Name,
		&newEt.Description,
		&newEt.Operation,
		&newEt.NumOperands,
		&newEt.Operands[0][0],
		&newEt.Operands[0][1],
		&newEt.Operands[1][0],
		&newEt.Operands[1][1],
		&newEt.Operands[2][0],
		&newEt.Operands[2][1],
		&newEt.Operands[3][0],
		&newEt.Operands[3][1],
		&newEt.No_remainder,
		&newEt.Result_max,
	)

	if err != nil {
		return nil, err
	}

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
