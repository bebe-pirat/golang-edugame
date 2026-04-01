package generator

import (
	"edugame/internal"
	"edugame/internal/entity"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type OperandRange struct {
	Order   int `json:"order"`
	MinValue int `json:"min_value"`
	MaxValue int `json:"max_value"`
}

type EquationType struct {
	ID          int             `json:"id"`
	Class       int             `json:"class"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Operation   string          `json:"operation"`
	NumOperands int             `json:"num_operands"`
	Operands    []OperandRange  `json:"operands"` // Динамический срез операндов
	NoRemainder bool            `json:"no_remainder"`
	ResultMax   int             `json:"result_max"`
	IsAvailable bool            `json:"is_available"`
}

type Equation struct {
	Text           string
	CorrectAnswer  string
	Class          int
	EquationTypeId int
}

type Generator struct {
	randSource *rand.Rand
}

func (g *Generator) GetRandSource() *rand.Rand {
	return g.randSource
}

func NewGenerator() *Generator {
	return &Generator{
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Generator) GenerateEquation(t EquationType) (Equation, error) {
	vars := make([]string, t.NumOperands)
	ops := make([]string, t.NumOperands-1)
	expr := make([]string, cap(vars)+cap(ops))
	var eqStr string = ""

	var correctAnswer int = 0
	var err error
	for {
		runes := []rune(t.Operation)
		for i := 0; i < t.NumOperands; i++ {
			// Используем динамический срез операндов вместо фиксированного массива
			operandRange := t.Operands[i]
			vars[i] = strconv.Itoa(g.randSource.Intn(operandRange.MaxValue-operandRange.MinValue) + operandRange.MinValue)
			expr[i*2] = vars[i]
			eqStr += vars[i]

			if i < cap(ops) {
				ops[i] = string(runes[g.randSource.Intn(len(runes))])
				if ops[i] == "/" {
					ops[i] = internal.DivSimbol
				} else if ops[i] == "*" {
					ops[i] = internal.MultSimbol
				}
				
				expr[i*2+1] = ops[i]
				eqStr += ops[i]
			}
		}
		eqStr += "= ?"

		log.Printf("Generating equation: %s", eqStr)
		m := entity.NewMather(expr, t.ResultMax)
		correctAnswer, err = m.Calculate()
		if err == nil {
			break
		} else {
			eqStr = ""
		}
	}

	return Equation{
		Text:           eqStr,
		CorrectAnswer:  strconv.Itoa(correctAnswer),
		Class:          t.Class,
		EquationTypeId: t.ID,
	}, nil
}
