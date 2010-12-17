package main

import (
	"fmt"
	"go/ast"
	"github.com/droundy/go/x86"
)

type Variable interface {
	InMemory() x86.Memory
	Type() *ast.Type
	Name() string
}

type StackVariable struct {
	T *ast.Type
	N string
	Offset int
}

func (v *StackVariable) InMemory() x86.Memory {
	return x86.Memory{ x86.Imm32(v.Offset), x86.ESP, nil, nil }
}
func (v *StackVariable) Type() *ast.Type {
	return v.T
}
func (v *StackVariable) Name() string {
	return v.N
}

type GlobalVariable struct {
	T *ast.Type
	N string
}

func (g *GlobalVariable) InMemory() x86.Memory {
	return x86.Memory{x86.Symbol(g.N),nil,nil,nil}
}
func (v *GlobalVariable) Type() *ast.Type {
	return v.T
}
func (v *GlobalVariable) Name() string {
	return v.N
}

// All global variables are accessible via Globals.
var Globals = make(map[string]GlobalVariable)

func DefineGlobal(name string, t *ast.Type) {
	Globals[name] = GlobalVariable{ t, name }
}

// Stack variable scope is visible through type.

type Stack struct {
	Parent *Stack
	Vars map[string]StackVariable
	Size int
	ReturnSize int
	Name string
}

// DefineVariable returns the offset to be subtracted from the stack
// pointer
func (s *Stack) DefineVariable(name string, t *ast.Type, synonymns ...string) int {
	if _,ok := s.Vars[name]; ok && name != "_" {
		panic(fmt.Sprintf("Cannot define already existing variable %s", name))
	}
	off := SizeOnStack(t)
	s.Size += off
	s.Vars[name] = StackVariable{ t, name, s.Size }
	for _,n := range synonymns {
		s.Vars[n] = StackVariable{ t, name, s.Size }
	}
	return off
}

// Pop returns the offset to be added to the stack pointer
func (s *Stack) Pop(t *ast.Type) int {
	off := SizeOnStack(t)
	s.Size -= off
	return off
}

// PopTo returns code to save data from the stack into the variable.
// It also changes the stack size accordingly.
func (s *Stack) PopTo(name string) x86.X86 {
	v := s.Lookup(name)
	off := SizeOnStack(v.Type())
	if TypeToSize(v.Type()) != off {
		panic("I can't yet handle types with sizes that aren't a multiple of 4")
	}
	s.Size -= off
	switch off {
	case 4:
		return x86.RawAssembly(x86.Assembly([]x86.X86{
			x86.PopL(x86.EAX),
			x86.Commented(x86.MovL(x86.EAX, v.InMemory()), "Popping to variable "+v.Name()),
		}))
	case 8:
		return x86.RawAssembly(x86.Assembly([]x86.X86{
			x86.PopL(x86.EAX),
			x86.Commented(x86.MovL(x86.EAX, v.InMemory().Add(4)), "Popping to variable "+v.Name()),
			x86.PopL(x86.EAX),
			x86.Commented(x86.MovL(x86.EAX, v.InMemory()), "Popping to variable "+v.Name()),
		}))
	default:
		panic(fmt.Sprintf("I don't pop variables with length %s", SizeOnStack(v.Type())))
	}
	panic("This can't happen")
}

func (s *Stack) Lookup(name string) (out Variable) {
	offtotal := 0
	for {
		if s == nil {
			if v,ok := Globals[name]; ok {
				return &v
			}
			panic("There is no variable named "+name)
		}
		offtotal += s.Size
		if v,ok := s.Vars[name]; ok {
			fmt.Println("For", name, "offtotal is", offtotal, "and v.Offset is", v.Offset)
			v.Offset = offtotal - v.Offset
			// Because we do this just here, we can't save this Variable for
			// later, when the stack pointer might have changed.
			return &v
		}
		s = s.Parent
	}
	panic("This can never happen")
	return
}

func (s *Stack) New(name string) *Stack {
	n := Stack{ s, make(map[string]StackVariable), 0, 0, name }
	return &n
}
