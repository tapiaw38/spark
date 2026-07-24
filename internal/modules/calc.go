package modules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

// CalcSearch evaluates math expressions
func CalcSearch(query string) []Result {
	if len(query) < 1 {
		return nil
	}

	// Try to evaluate as math expression
	result, err := evalExpr(query)
	if err != nil {
		return nil
	}

	resultStr := formatNumber(result)
	return []Result{{
		Type:  "calc",
		Title: fmt.Sprintf("= %s", resultStr),
		Desc:  "Copy to clipboard",
		Icon:  "accessories-calculator",
		Action: func() {
			copyToClipboard(resultStr)
		},
	}}
}

func evalExpr(expr string) (float64, error) {
	// ponytail: use Go parser for safe math eval
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}
	return eval(node)
}

func eval(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		return strconv.ParseFloat(n.Value, 64)
	case *ast.ParenExpr:
		return eval(n.X)
	case *ast.BinaryExpr:
		left, err := eval(n.X)
		if err != nil {
			return 0, err
		}
		right, err := eval(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
	case *ast.UnaryExpr:
		val, err := eval(n.X)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -val, nil
		}
		return val, nil
	}
	return 0, fmt.Errorf("unsupported expression")
}

func formatNumber(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%.6g", f)
}
