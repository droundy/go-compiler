package main

import (
	"fmt"
	"go/ast"
	"github.com/droundy/go/x86"
)

type Variable interface {
	InMemory() x86.Memory
	Type() *ast.Type
}

type StackVariable struct {
	T *ast.Type
	Name string
	Offset int
}

func (v *StackVariable) InMemory() x86.Memory {
	return x86.Memory{ x86.Imm32(v.Offset), x86.ESP, nil, nil }
}
func (v *StackVariable) Type() *ast.Type {
	return v.T
}

type GlobalVariable struct {
	T *ast.Type
	Name string
}

func (g *GlobalVariable) InMemory() x86.Memory {
	return x86.Memory{x86.Symbol(g.Name),nil,nil,nil}
}
func (v *GlobalVariable) Type() *ast.Type {
	return v.T
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
	Name string
}

// DefineVariable returns the offset to be subtracted from the stack
// pointer
func (s *Stack) DefineVariable(name string, t *ast.Type) int {
	if _,ok := s.Vars[name]; ok && name != "_" {
		panic(fmt.Sprintf("Cannot define already existing variable %s", name))
	}
	off := SizeOnStack(t)
	s.Size += off
	s.Vars[name] = StackVariable{ t, name, s.Size }
	return off
}

// Pop returns the offset to be added to the stack pointer
func (s *Stack) Pop(t *ast.Type) int {
	off := SizeOnStack(t)
	s.Size -= off
	return off
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
	n := Stack{ s, make(map[string]StackVariable), 0, name }
	return &n
}
