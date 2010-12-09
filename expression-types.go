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
	case *ast.CallExpr:
		switch fn := e.Fun.(type) {
		case *ast.Ident:
			switch fn.Name {
			case "println":
				return *ast.NewType(ast.Tuple)
			default:
				// FIXME: this assumes all functions return no values, which
				// is clearly false.
				return *ast.NewType(ast.Tuple)
			}
		default:
			panic(fmt.Sprintf("Can't handle function of weird type %T", e.Fun))
		}
	case *ast.Ident:
		if e.Obj == nil {
			panic("There is no type information in " + e.Name)
		}
		if e.Obj.Type != nil {
			return *e.Obj.Type
		}
		panic("I don't know how to handle identifier " + e.Name)
	default:
		panic(fmt.Sprintf("I can't find type of expression %s of type %T\n", e0, e0))
	}
	return
}
