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

var myfiles = token.NewFileSet()

type StringVisitor CompileVisitor
func (v StringVisitor) Visit(n0 interface{}) (w ast.Visitor) {
	//fmt.Printf("in StringVisitor, n0 is %s of type %T\n", n0, n0)
	if n,ok := n0.(*ast.BasicLit); ok && n != nil && n.Kind == token.STRING {
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
	Stack *Stack
}
func (v *CompileVisitor) Append(xs... x86.X86) {
	*v.assembly = append(*v.assembly, xs...)
}
func (v *CompileVisitor) FunctionPrologue(fn *ast.FuncDecl) {
	v.Stack = v.Stack.New(fn.Name.Name)
	ftype := ast.NewType(ast.Function)
	ftype.N = uint(fn.Type.Results.NumFields())
	ftype.Params = ast.NewScope(nil)
	fmt.Println("Working on function", fn.Name.Name)
	if fn.Type.Results != nil {
		for _,resultfield := range fn.Type.Results.List {
			names := []string{"_"}
			if resultfield.Names != nil {
				names = []string{}
				for _,i := range resultfield.Names {
					names = append(names, i.Name)
				}
			}
			t := TypeExpression(resultfield.Type)
			for _,n := range names {
				ftype.Params.Insert(&ast.Object{ ast.Fun, n, t, resultfield, 0 })
				v.Stack.DefineVariable(n, t)
			}
		}
	}
	fmt.Println("Stack size after results is", v.Stack.Size)
	for _,paramfield := range fn.Type.Params.List {
		names := []string{"_"}
		if paramfield.Names != nil {
			names = []string{}
			for _,i := range paramfield.Names {
				names = append(names, i.Name)
			}
		}
		t := TypeExpression(paramfield.Type)
		for _,n := range names {
			ftype.Params.Insert(&ast.Object{ ast.Fun, n, t, paramfield, 0 })
			v.Stack.DefineVariable(n, t)
		}
	}
	fmt.Println("Stack size after params is", v.Stack.Size)
	v.Stack.DefineVariable("return", IntType)
	fmt.Println("Stack size after return is", v.Stack.Size)
	v.Stack = v.Stack.New("_")
	DefineGlobal(fn.Name.Name, ftype)
	// symbol for the start name
	pos := myfiles.Position(fn.Pos())
	v.Append(x86.Commented(x86.GlobalSymbol("main_"+fn.Name.Name),
		fmt.Sprint(pos.Filename, ": line ", pos.Line)))
	// If we had arguments, we'd want to swap them with the return
	// address here...
}
func (v *CompileVisitor) FunctionPostlogue() {
	// First we roll back the stack from where we started...
	for v.Stack.Name == "_" {
		// We need to pop off any extra layers of stack we've added...
		if v.Stack.Size > 0 {
			v.Append(x86.Commented(x86.AddL(x86.Imm32(v.Stack.Size), x86.ESP),
				"We stored this much on the stack so far."))
		}
		// Now jump to the "real" postlogue.  This is a little stupid, but I
		// expect it'll come in handy when I implement defer (not to mention
		// panic/recover).
		v.Stack = v.Stack.Parent // We've popped off the arguments...
	}
	v.Append(x86.Jmp(x86.Symbol("return_" + v.Stack.Name)))
}

func (v *CompileVisitor) Visit(n0 interface{}) (w ast.Visitor) {
	// The following only handles functions (not methods)
	if n,ok := n0.(*ast.FuncDecl); ok && n.Recv == nil {
		v.FunctionPrologue(n)
		for _,statement := range n.Body.List {
			v.CompileStatement(statement)
		}
		v.FunctionPostlogue()
		v.Append(x86.GlobalSymbol("return_"+n.Name.Name))
		// Pop off function arguments...
		// FIXME this would also pop off return values...
		v.Append(x86.Commented(x86.PopL(x86.EAX), "Pop the return address"))
		fmt.Println("Function", v.Stack.Name, "has stack size", v.Stack.Size)
		v.Append(x86.Commented(x86.AddL(x86.Imm32(v.Stack.Size - 4), x86.ESP),
			"Popping "+v.Stack.Name+" arguments."))
		// Then we return!
		v.Append(x86.RawAssembly("\tjmp *%eax"))
		v.Stack = v.Stack.Parent
		return nil // No need to peek inside the func declaration!
	}
	return v
}
func (v *CompileVisitor) PopType(t ast.Type) {
	switch t.Form {
	case ast.Tuple:
		for _,o := range t.Params.Objects {
			v.PopType(*o.Type)
		}
	case ast.Basic:
		switch t.N {
		case ast.String:
			v.Append(x86.AddL(x86.Imm32(8), x86.ESP))
		default:
			panic(fmt.Sprintf("I don't know how to pop basic type %s", t))
		}
	default:
		panic(fmt.Sprintf("I don't know how to pop type %s", t.Form))
	}
}
func (v *CompileVisitor) CompileStatement(statement ast.Stmt) {
	switch s := statement.(type) {
	case *ast.EmptyStmt:
		// It is empty, I can handle that!
	case *ast.ExprStmt:
		v.CompileExpression(s.X)
		t := ExprType(s.X, v.Stack)
		switch t.Form {
		case ast.Tuple:
			
		}
	case *ast.ReturnStmt:
		if len(s.Results) != 0 {
			panic("I can't handle return statements with values just yet...")
		}
		v.FunctionPostlogue()
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
			pos := myfiles.Position(fn.Pos())
			switch fn.Name {
			case "println":
				if len(e.Args) != 1 {
					panic(fmt.Sprintf("println expects just one argument, not %d", len(e.Args)))
				}
				argtype := ExprType(e.Args[0], v.Stack)
				if argtype.N != ast.String || argtype.Form != ast.Basic {
					panic(fmt.Sprintf("Argument to println has type %s but should have type string!",
						argtype))
				}
				v.Stack = v.Stack.New("arguments")
				v.CompileExpression(e.Args[0])
				v.Append(x86.Commented(x86.Call(x86.Symbol("println")),
					fmt.Sprint(pos.Filename, ": line ", pos.Line)))
				v.Stack = v.Stack.Parent // A hack to let the callee clean up arguments
			default:
				// This must not be a built-in function...
				functype := v.Stack.Lookup(fn.Name)
				if functype.Type().Form != ast.Function {
					panic("Function "+ fn.Name + " is not actually a function!")
				}
				if functype.Type().N != 0 {
					panic("I can't yet handle functions with a return value such as "+fn.Name)
				}
				v.Stack = v.Stack.New("arguments")
				for i:=len(e.Args)-1; i>=0; i-- {
					v.CompileExpression(e.Args[i])
				}
				// FIXME: I assume here that there is no return value!
				v.Append(x86.Commented(x86.Call(x86.Symbol("main_"+fn.Name)),
					fmt.Sprint(pos.Filename, ": line ", pos.Line)))
				v.Stack = v.Stack.Parent // A hack to let the callee clean up arguments
			}
		} else {
			panic(fmt.Sprintf("I don't know how to deal with complicated function: %s", e.Fun))
		}
	case *ast.Ident:
		evar := v.Stack.Lookup(e.Name)
		switch SizeOnStack(evar.Type()) {
		case 4:
			v.Append(x86.Commented(x86.MovL(evar.InMemory(), x86.EAX), "Reading variable "+e.Name))
			v.Append(x86.PushL(x86.EAX))
			v.Stack.DefineVariable("_", evar.Type())
		case 8:
			v.Append(x86.Comment(fmt.Sprintf("The offset of %s is %s", e.Name, evar.InMemory())))
			v.Append(x86.Commented(x86.MovL(evar.InMemory().Add(4), x86.EAX),
				"Reading variable "+e.Name))
			v.Append(x86.MovL(evar.InMemory(), x86.EBX))
			v.Append(x86.PushL(x86.EAX))
			v.Append(x86.PushL(x86.EBX))
			v.Stack.DefineVariable("_", evar.Type())
		default:
			panic(fmt.Sprintf("I don't handle variables with length %s", SizeOnStack(evar.Type())))
		}
	default:
		panic(fmt.Sprintf("I can't handle expressions such as: %T value %s", exp, exp))
	}
}

func main() {
	goopt.Parse(func() []string { return nil })
	if len(goopt.Args) > 0 {
		x,err := parser.ParseFiles(myfiles, goopt.Args, 0)
		die(err)
		fmt.Fprintln(os.Stderr, "Parsed: ", *x["main"])
		die(typechecker.CheckPackage(myfiles, x["main"], nil))
		fmt.Fprintln(os.Stderr, "Checked: ", *x["main"])
		//for _,a := range x["main"].Files {
		//	die(printer.Fprint(os.Stdout, a))
		//}

		aaa := x86.StartData
		var bbb *Stack
		var cv = CompileVisitor{ &aaa, make(map[string]string), bbb.New("global")}
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
