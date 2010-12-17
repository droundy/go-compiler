package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func ExprType(e0 ast.Expr, s *Stack) (t *ast.Type) {
	t = ast.NewType(ast.Basic)
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
				return ast.NewType(ast.Tuple)
			default:
				ftype := s.Lookup(fn.Name).Type()
				switch ftype.N {
				case 0:
					// This is a function with no return value: easy!
					return ast.NewType(ast.Tuple)
				case 1:
					return ftype.Params.Objects[0].Type
				default:
					panic("I don't yet do multiple return types...")
				}
			}
		default:
			panic(fmt.Sprintf("Can't handle function of weird type %T", e.Fun))
		}
	case *ast.Ident:
		return s.Lookup(e.Name).Type()
	default:
		panic(fmt.Sprintf("I can't find type of expression %s of type %T\n", e0, e0))
	}
	return
}

func TypeExpression(e ast.Expr) (t *ast.Type) {
	switch e := e.(type) {
	case *ast.Ident:
		switch e.Name {
		case "string":
			t = ast.NewType(ast.Basic)
			t.N = ast.String
			return
		default:
			panic("I don't understand type "+e.Name)
		}
	default:
		panic(fmt.Sprintf("I can't understand type expression %s of type %T\n", e, e))
	}
	panic(fmt.Sprintf("I don't understand the type expression %s", e))
}
