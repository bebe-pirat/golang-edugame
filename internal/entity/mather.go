package entity

import "strconv"

type Mather struct {
	infix             []string
	postfix           []string
	operationPriotiry map[string]int
	maxResult         int
}

func NewMather(infix_ []string, maxResult int) *Mather {
	return &Mather{
		infix:   infix_,
		postfix: make([]string, 0),
		operationPriotiry: map[string]int{
			"+": 1,
			"-": 1,
			"*": 2,
			"/": 2,
		},
		maxResult: maxResult,
	}
}

func (m *Mather) infixExprToPostfix() {
	output := make([]string, 0)
	stack := make([]string, 0)

	for _, token := range m.infix {
		_, err := strconv.Atoi(token)

		if err == nil {
			output = append(output, token)
			continue
		}

		switch token {
		case "(":
			stack = append(stack, token)

		case ")":
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				if top == "(" {
					break
				}
				output = append(output, top)
			}

		case "+", "-", "*", "/":
			for len(stack) > 0 {
				top := stack[len(stack)-1]

				if top == "(" || !m.isOperator(token) {
					break
				}

				if m.operationPriotiry[token] > m.operationPriotiry[top] {
					break
				}

				output = append(output, top)
				stack = stack[:len(stack)-1]
			}

			stack = append(stack, token)
		default:

		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if top != "(" {
			output = append(output, top)
		}
	}

	m.postfix = output
}

func (m *Mather) isOperator(token string) bool {
	_, exists := m.operationPriotiry[token]
	return exists
}

func (m *Mather) calculatePostfix() (int, error) {
	if len(m.postfix) <= 0 {
		return 0, &CalculationError{"Invalid expression"}
	}

	stack := make([]int, 0)

	for _, token := range m.postfix {
		integer, err := strconv.Atoi(token)

		if err == nil {
			stack = append(stack, integer)
			continue
		}

		if m.isOperator(token) {
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			b := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			result, err := m.calculateOperation(b, a, token)

			if err != nil {
				return 0, err
			}

			stack = append(stack, result)
		}
	}

	if len(stack) != 1 {
		return 0, &CalculationError{"Invalid expression"}
	}

	return stack[0], nil
}

func (m *Mather) calculateOperation(a, b int, op string) (int, error) {
	switch op {
	case "+":
		if a+b <= 0 {
			return 0, &CalculationError{"Under zero"}
		}
		return a + b, nil

	case "-":
		if a-b <= 0 {
			return 0, &CalculationError{"Under zero"}
		}
		return a - b, nil

	case "*":
		if a*b <= 0 {
			return 0, &CalculationError{"Under zero"}
		}
		return a * b, nil

	case "/":
		if b == 0 || a%b != 0 {
			return 0, &CalculationError{"Division by zero"}
		}
		return a / b, nil
	}
	return 0, &CalculationError{"Unknown operation"}
}

type CalculationError struct {
	Message string
}

func (e *CalculationError) Error() string {
	return "Calculation error: " + e.Message
}

func (m *Mather) Calculate() (int, error) {
	m.infixExprToPostfix()

	return m.calculatePostfix()
}
