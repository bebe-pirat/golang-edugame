package generator

import (
	"edugame/internal"
	"edugame/internal/entity"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type EquationType struct {
	ID          int
	Class       int
	Name        string
	Description string
	Operation   string
	NumOperands int

	Operands [4][2]int

	No_remainder bool
	Result_max   int
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
			vars[i] = strconv.Itoa(g.randSource.Intn(t.Operands[i][1]-t.Operands[i][0]) + t.Operands[i][0])
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

		log.Printf("dsa")
		m := entity.NewMather(expr, t.Result_max)
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
