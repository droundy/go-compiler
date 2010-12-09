package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func ExprType(e0 ast.Expr) (t ast.Type) {
	t = *ast.NewType(ast.Basic)
	switch e := e0.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			t.N = ast.String
		default:
			panic(fmt.Sprintf("I don't handle basic literals such as %s", e))
		}
	default:
		panic(fmt.Sprintf("I can't find type of expression %s of type %T\n", e0, e0))
	}
	return
}
