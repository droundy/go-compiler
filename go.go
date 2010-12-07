package main

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"unicode"
	"go/ast"
	"go/token"
	"go/parser"
	"go/typechecker"
	"github.com/droundy/go/elf"
	"github.com/droundy/go/x86"
	"github.com/droundy/goopt"
)

type StringVisitor CompileVisitor
func (v StringVisitor) Visit(n0 interface{}) (w ast.Visitor) {
	if n,ok := n0.(*ast.BasicLit); ok && n.Kind == token.STRING {
		str,err := strconv.Unquote(string(n.Value))
		if err != nil {
			panic(err)
		}
		sanitize := func(rune int) int {
			if unicode.IsLetter(rune) {
				return rune
			}
			return -1
		}
		if _,ok := v.string_literals[str]; !ok {
			strname := "string_" + strings.Map(sanitize, str)
			*v.assembly = append(*v.assembly,
				x86.Symbol(strname),
				x86.Commented(x86.Ascii(str), "a non-null-terminated string"))
			v.string_literals[str] = strname
		}
	}
	return v
}

type CompileVisitor struct {
	assembly *[]x86.X86
	string_literals map[string]string
}
func (v *CompileVisitor) Append(xs... x86.X86) {
	*v.assembly = append(*v.assembly, xs...)
}
func (v *CompileVisitor) Visit(n0 interface{}) (w ast.Visitor) {
	// The following only handles functions (not methods)
	if n,ok := n0.(*ast.FuncDecl); ok && n.Recv == nil {
		v.Append(x86.Commented(x86.GlobalSymbol("main_"+n.Name.Name), "from where?"))
		for _,statement := range n.Body.List {
			v.CompileStatement(statement)
		}
		v.Append(x86.Return("from main_"+n.Name.Name))
		return nil // No need to peek inside the func declaration!
	}
	return v
}
func (v *CompileVisitor) CompileStatement(statement ast.Stmt) {
	switch s := statement.(type) {
	case *ast.EmptyStmt:
		// It is empty, I can handle that!
	case *ast.ExprStmt:
		v.CompileExpression(s.X)
	default:
		panic(fmt.Sprintf("I can't handle statements such as: %T", statement))
	}
}
func (v *CompileVisitor) CompileExpression(exp ast.Expr) {
	switch e := exp.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			str,err := strconv.Unquote(string(e.Value))
			if err != nil {
				panic(err)
			}
			n,ok := v.string_literals[str]
			if !ok {
				panic(fmt.Sprintf("I don't recognize the string: %s", string(e.Value)))
			}
			v.Append(
				x86.Commented(x86.PushL(x86.Symbol(n)), "Pushing string literal "+string(e.Value)),
				x86.PushL(x86.Imm32(len(str))))
		default:
			panic(fmt.Sprintf("I don't know how to deal with literal: %s", e))
		}
	case *ast.CallExpr:
		if fn,ok := e.Fun.(*ast.Ident); ok {
			switch fn.Name {
			case "println":
				if len(e.Args) != 1 {
					panic(fmt.Sprintf("println expects just one argument, not %d", len(e.Args)))
				}
				v.CompileExpression(e.Args[0])
				v.Append(x86.RawAssembly("\tcall println"))
			default:
				// This must not be a built-in function...
				if len(e.Args) != 0 {
					panic("I don't know how to handle functions with arguments yet...")
				}
				v.Append(x86.RawAssembly("\tcall main_"+fn.Name+" # FIXME this assumes no return value!"))
			}
		} else {
			panic(fmt.Sprintf("I don't know how to deal with complicated function: %s", e.Fun))
		}
	default:
		panic(fmt.Sprintf("I can't handle expressions such as: %T", exp))
	}
}

func main() {
	goopt.Parse(func() []string { return nil })
	if len(goopt.Args) > 0 {
		x,err := parser.ParseFiles(goopt.Args, 0)
		die(err)
		fmt.Fprintln(os.Stderr, "Parsed: ", *x["main"])
		die(typechecker.CheckPackage(x["main"], nil))
		fmt.Fprintln(os.Stderr, "Checked: ", *x["main"])
		//for _,a := range x["main"].Files {
		//	die(printer.Fprint(os.Stdout, a))
		//}

		aaa := x86.StartData
		var cv = CompileVisitor{assembly: &aaa, string_literals: make(map[string]string)}
		ast.Walk(StringVisitor(cv), x["main"])

		cv.Append(x86.StartText...)
		ast.Walk(&cv, x["main"])

		// Here we just add a crude debug library
		cv.Append(x86.Debugging...)
		ass := x86.Assembly(*cv.assembly)
		//fmt.Println(ass)
		die(elf.AssembleAndLink(goopt.Args[0][:len(goopt.Args[0])-3], []byte(ass)))
	}
}


func die(err os.Error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
