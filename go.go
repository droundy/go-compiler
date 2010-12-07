package main

import (
	"os"
	"fmt"
	"strings"
	"unicode"
	"go/ast"
	"go/token"
	"go/parser"
	"go/printer"
	"go/typechecker"
	"github.com/droundy/go/elf"
	"github.com/droundy/go/x86"
	"github.com/droundy/goopt"
)

type StringVisitor CompileVisitor
func (v StringVisitor) Visit(n0 interface{}) (w ast.Visitor) {
	if n,ok := n0.(*ast.BasicLit); ok && n.Kind == token.STRING {
		str := string(n.Value)
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

type CallVisitor int
func (v CallVisitor) Visit(n0 interface{}) (w ast.Visitor) {
	if n,ok := n0.(*ast.CallExpr); ok && len(n.Args) == 1 {
		if fn,ok := n.Fun.(*ast.Ident); ok && fn.Name == "println" {
			fmt.Println("Found a println", n.Args[0])
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
			n,ok := v.string_literals[string(e.Value)]
			if !ok {
				panic(fmt.Sprintf("I don't recognize the string: %s", string(e.Value)))
			}
			v.Append(
				x86.Commented(x86.PushL(x86.Symbol(n)), "Pushing string literal "+string(e.Value)),
				x86.PushL(x86.Imm32(len(string(e.Value)))))
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
				panic(fmt.Sprintf("I don't know how to deal with function: %s", e.Fun))
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
		for _,a := range x["main"].Files {
			die(printer.Fprint(os.Stdout, a))
		}

		aaa := x86.StartData
		var cv = CompileVisitor{assembly: &aaa, string_literals: make(map[string]string)}
		ast.Walk(StringVisitor(cv), x["main"])
		ast.Walk(CallVisitor(0), x["main"])

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

var hello = []x86.X86{
	x86.Section("data"),
	x86.Symbol("goc.syscall"),
	x86.Commented(x86.GlobalInt(0),
		"This is a global variable for the address for syscalls"),
	x86.Symbol("goc.args"),
	x86.Commented(x86.GlobalInt(0),
		"This is the number of args"),
	x86.Symbol("goc.argsptr"),
	x86.Commented(x86.GlobalInt(0),
		"This is a pointer to the actual args"),

	x86.Symbol("msg"),
	x86.Commented(x86.Ascii("Hello, world!\n"), "a non-null-terminated string"),
	x86.Commented(x86.SymbolicConstant(x86.Symbol("len"), ". - msg"), "length of string"),
	x86.Section("text"),
	x86.Commented(x86.GlobalSymbol("_start"), "this says where to start execution"),

	x86.Comment("Search for mksyscall value in the ELF auxiliary vectors..."),
	x86.Commented(x86.MovL(x86.ESP, x86.EBP), "Save stack pointer for later."),
	x86.Commented(x86.MovL(
		x86.Memory{x86.Imm32(0), x86.ESP, nil, nil}, x86.EAX), "Read value of argc."),
	x86.Commented(x86.MovL(x86.EAX, x86.Symbol("goc.args")), "Save value of argc."),

	x86.Comment("Find first env var..."),
	x86.Commented(x86.MovL(x86.ESP, x86.EBX), "initialize a copy of the stack"),
	x86.Commented(x86.MovL(x86.Imm32(1), x86.EAX), "1 for extra NULL"),
	x86.Commented(x86.AddL(x86.Memory{x86.Symbol("goc.args"),nil,nil,nil}, x86.EAX),
		"Read # of args"),

	x86.Symbol("loopstart"),
	x86.Commented(x86.AddL(x86.Imm32(1), x86.EAX), "increment EAX"),
	x86.CmpL(x86.Imm32(0), x86.Memory{nil, x86.ESP, x86.EAX, x86.Imm32(4)}),
	x86.Jne(x86.Symbol("loopstart")),
	x86.Commented(x86.AddL(x86.Imm32(1), x86.EAX), "increment EAX one last time"),
	x86.Commented(x86.IMulL(x86.Imm32(4), x86.EAX), "multiply by four"),
	x86.Commented(x86.AddL(x86.ESP, x86.EAX), "add to stack pointer"),
	x86.MovL(x86.EAX, x86.Symbol("goc.argsptr")),

	//x86.Commented(x86.MovL(x86.Imm32(4), x86.EAX), "index to env var (in progress)"),
	//x86.Commented(x86.IMulL(x86.Memory{x86.Symbol("goc.args"),nil,nil,nil}, x86.EAX),
	//	"Multiply # of args by four to get offset"),
	//x86.Commented(x86.AddL(x86.Imm32(4), x86.EAX), "Add another four, for the extra blank word"),

	x86.Comment("Print integer..."),
	x86.Commented(x86.MovL(x86.Imm32(1), x86.EDX), "third argument: data length"),
	x86.Commented(x86.MovL(x86.Symbol("goc.args"), x86.ECX), "second argument: pointer to data"),
	x86.Commented(x86.AddL(x86.Imm32('0'), x86.Symbol("goc.args")),
		"make decimal from number"),
	x86.Commented(x86.MovL(x86.Imm32(1), x86.EBX), "first argument: file handle (stdout)"),
	x86.Commented(x86.MovL(x86.Imm32(4), x86.EAX), "system call number (sys_write)"),
	x86.Int(x86.Imm32(0x80)),
	x86.Commented(x86.AddL(x86.Imm32(-'0'), x86.Symbol("goc.args")),
		"change number back from decimal"),

	x86.Comment("Print string..."),
	x86.Commented(x86.MovL(x86.Symbol("len"), x86.EDX), "third argument: data length"),
	x86.Commented(x86.MovL(x86.Symbol("msg"), x86.ECX), "second argument: pointer to data"),
	x86.Commented(x86.MovL(x86.Imm32(1), x86.EBX), "first argument: file handle (stdout)"),
	x86.Commented(x86.MovL(x86.Imm32(4), x86.EAX), "system call number (sys_write)"),
	x86.Int(x86.Imm32(0x80)),

	x86.Commented(x86.MovL(x86.Imm32(4), x86.EAX), "system call number (sys_write)"),
	x86.Call(x86.Symbol("debug.print_eax")),

	x86.Comment("And exit..."),
	x86.Commented(x86.MovL(x86.Imm32(0), x86.EBX), "first argument: exit code"),
	x86.Commented(x86.MovL(x86.Imm32(1), x86.EAX), "system call number (sys_exit)"),
	x86.Int(x86.Imm32(0x80)),
}
