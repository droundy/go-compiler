package main

import (
	"fmt"
	"go/ast"
)

func TypeToSize(t *ast.Type) (out int) {
	switch t.Form {
	case ast.Tuple:
		for _,o := range t.Params.Objects {
			out += TypeToSize(o.Type)
		}
		return out
	case ast.Basic:
		switch t.N {
		case ast.String:
			return 8
		case ast.Int:
			return 4
		default:
			panic(fmt.Sprintf("I don't know size of basic type %s", t))
		}
	default:
		panic(fmt.Sprintf("I don't know how to pop type %s", t.Form))
	}
	return
}

func SizeOnStack(t *ast.Type) (out int) {
	out = TypeToSize(t)
	if out & 3 != 0 {
		return 4*(out/4 + 1)
	}
	return
}

var IntType *ast.Type = ast.NewType(ast.Basic)

func init() {
	IntType.N = ast.Int
}
