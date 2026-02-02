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

// получить полный список типов уравнений для определенного класса
func (r *TypeRepository) GetListTypes(class int) ([]generator.EquationType, error) {
	query := `
        SELECT id, class, name, description, operation, num_operands, operand1_min, operand1_max, operand2_min, operand2_max, COALESCE(operand3_min, -1), COALESCE(operand3_max, -1), COALESCE(operand4_min, -1), COALESCE(operand4_max, -1), no_remainder, COALESCE(result_max, -1)
        FROM equation_types
        WHERE class = $1 AND is_available = TRUE
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
		SELECT * 
		FROM equation_types
		WHERE id = ?
	`, id).Scan(
		&t.ID,
		&t.Class,
		&t.Name,
		&t.Description,
		&t.Operation,
		&t.NumOperands,

		&t.Operands[0][0],
		&t.Operands[0][0],
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
