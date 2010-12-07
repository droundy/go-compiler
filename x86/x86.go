package x86

import (
	"fmt"
)

func Assembly(code []X86) (out string) {
	// FIXME:  This is stupidly O(N^2)...
	for _,a := range code {
		out += a.X86() + "\n"
	}
	return
}

// An Assembly is something that can be converted into a line of
// assembly language.

type X86 interface {
	X86() string
}

// a W32 is a source (or sink) for a double-word (which I think of as
// a word, but x86 assembly doesn't), which could either be a register
// or a memory location.

type W32 interface {
	W32() string
}

// a W16 is a source (or sink) for a word (which I think of as a
// short, but x86 assembly doesn't), which could either be a register
// or a memory location.

type W16 interface {
	W16() string
}

// a W8 is a source (or sink) for a byte, which could either be a
// register or a memory location.

type W8 interface {
	W8() string
}

type Ptr interface {
	Ptr() string
}

// And now we come to the concrete types.

// A Comment is the most boring part of assembly, but pretty useful
// for trying to understand generated code.

type commentType struct {
	instr X86
	c string
}
func (s commentType) X86() string {
	if c,ok := s.instr.(commentType); ok {
		return "\n#  " + c.c + "\n"
	}
	return s.instr.X86() + "\t# " + s.c
}
func Comment(x string) X86 {
	return commentType{commentType{Symbol(""), x}, ""}
}
func Commented(instr X86, x string) X86 {
	return commentType{instr, x}
}

// A Symbol marks a location in the binary source file, which could be
// a function or a global variable or even a global constant.

type Symbol string

func (s Symbol) X86() string {
	return string(s) + ":"
}
func (s Symbol) W8() string {
	return "$" + string(s)
}
func (s Symbol) W16() string {
	return "$" + string(s)
}
func (s Symbol) W32() string {
	return "$" + string(s)
}
func (s Symbol) Ptr() string {
	return string(s)
}

// Memory is an in-memory reference

type Memory struct {
	Disp Ptr
	Base W32
	Index W32
	Scale Ptr
}
func (m Memory) W32() string {
	offstr := ""
	if m.Disp != nil {
		offstr = m.Disp.Ptr()
	}
	bstr := ""
	if m.Base != nil {
		bstr = m.Base.W32()
	}
	istr := ""
	if m.Index != nil {
		istr = ", " + m.Index.W32()
	}
	scalestr := ""
	if m.Scale != nil {
		scalestr = ", " + m.Scale.Ptr()
	} else if m.Base == nil && m.Index == nil {
		scalestr = ", 1"
	}
	return offstr + "(" + bstr + istr + scalestr + ")"
}
func (m Memory) Ptr() string {
	return m.W32()
}

// A Register refers to a general-purpose register, of which the x86
// has only eight, two of which are pretty much devoted to the stack.

type Register byte
const (
	EAX Register = iota
	EBX
	ECX
	EDX
  EDI
  ESI
  EBP
  ESP
)

func (r Register) String() string {
	switch r {
	case EAX: return "ax"
	case EBX: return "bx"
	case ECX: return "cx"
	case EDX: return "dx"
	case EDI: return "di"
	case ESI: return "si"
	case EBP: return "bp"
	case ESP: return "sp"
	}
	panic(fmt.Sprint("Bad register value: ", r))
}
func (r Register) W8() string {
	// I haven't come up with a great approach for handling the high
	// byte registers (with "h" at the end below).  For now, I just
	// won't use them.
	switch r {
	case EAX: return "%al"
	case EBX: return "%bl"
	case ECX: return "%cl"
	case EDX: return "%dl"
	}
	panic(fmt.Sprint("Bad 8-bit register value: ", r))
}
func (r Register) W16() string {
	return "%" + r.String()
}
func (r Register) W32() string {
	return "%e" + r.String()
}
func (r Register) Ptr() string {
	return "%e" + r.String()
}

// Imm32 represents an immediate 32-bit value

type Imm32 int32
func (i Imm32) W32() string {
	return "$" + fmt.Sprint(i)
}
func (i Imm32) Ptr() string {
	return fmt.Sprint(i)
}

// Imm16 represents an immediate 16-bit value

type Imm16 int16
func (i Imm16) W16() string {
	return "$" + fmt.Sprint(i)
}

// Imm8 represents an immediate 8-bit byte

type Imm8 int8
func (i Imm8) W8() string {
	return "$" + fmt.Sprint(i)
}

// OpL2 holds any two-argument instructions involving 32-bit arguments
// of with the latter is an "output" argument.  It shouldn't need to
// be exported, but it could also come in handy at some stage...

type OpL2 struct {
	name string
	src W32
	dest Ptr
}
func (o OpL2) X86() string {
	return "\t" + o.name + " " + o.src.W32() + ", " + o.dest.Ptr()
}

func MovL(src W32, dest Ptr) X86 {
	return OpL2{"movl", src, dest}
}

func AddL(src W32, dest Ptr) X86 {
	return OpL2{"addl", src, dest}
}

func AndL(src W32, dest Ptr) X86 {
	return OpL2{"andl", src, dest}
}

func ShiftLeftL(src W32, dest Ptr) X86 {
	return OpL2{"shll", src, dest}
}

func ShiftRightL(src W32, dest Ptr) X86 {
	return OpL2{"shrl", src, dest}
}

func IMulL(src W32, dest Ptr) X86 {
	return OpL2{"imull", src, dest}
}

// OpLL holds any two-argument instructions involving 32-bit arguments
// in which either could be immediate.  It shouldn't need to be
// exported, but it could also come in handy at some stage...

type OpLL struct {
	name string
	src1, src2 W32
}
func (o OpLL) X86() string {
	return "\t" + o.name + " " + o.src1.W32() + ", " + o.src2.W32()
}

func CmpL(src W32, dest W32) X86 {
	return OpLL{"cmpl", src, dest}
}

// OpL1 holds any instruction involving a single 32-bit argument.  It
// shouldn't need to be exported, but it could also come in handy at
// some stage...

type OpL1 struct {
	name string
	arg W32
}
func (o OpL1) X86() string {
	return "\t" + o.name + " " + o.arg.W32()
}

func Int(val W32) X86 {
	return OpL1{"int", val}
}

func PopL(dest W32) X86 {
	return OpL1{"popl", dest}
}

func PushL(src W32) X86 {
	return OpL1{"pushl", src}
}

type Op0 struct {
	name, comment string
}
func (o Op0) X86() string {
	return "\t" + o.name + "\t# " + o.comment
}

func Return(com string) X86 {
	return Op0{ "ret", com }
}

// OpL1 holds any instruction involving a single argument that must be
// an address.  It shouldn't need to be exported, but it could also
// come in handy at some stage...

type OpP1 struct {
	name string
	arg Ptr
}
func (o OpP1) X86() string {
	return "\t" + o.name + " " + o.arg.Ptr()
}

func Jne(src Ptr) X86 {
	return OpP1{"jne", src}
}

func Call(src Ptr) X86 {
	return OpP1{"call", src}
}

func Jmp(src Ptr) X86 {
	return OpP1{"jmp", src}
}

// A Section is... a section.

type Section string
func (s Section) X86() string {
	return "." + string(s)
}

// Ascii is just raw text...

type Ascii string
func (a Ascii) X86() (out string) {
	// FIXME:  This is stupidly O(N^2)...
	out = "\t.ascii\t\""
	for i := range a {
		switch a[i] {
		case '"':	out += `\"`
		case '\n': out += `\n`
		default: out += string([]byte{a[i]})
		}
	}
	out += `"`
	return
}

// Int is just raw 32-bit number...

type GlobalInt int32
func (a GlobalInt) X86() string {
	return "\t.int\t" + fmt.Sprint(a)
}

// SymbolicConstant defines a symbolic constant...

type symbolicConstant struct {
	name Symbol
	value string
}

func (a symbolicConstant) X86() string {
	return "\t" + string(a.name) + " = " + a.value
}

func SymbolicConstant(name Symbol, val string) X86 {
	return symbolicConstant{name, val}
}

type RawAssembly string
func (r RawAssembly) X86() string {
	return string(r)
}

func GlobalSymbol(name string) X86 {
	return RawAssembly(".global " + name + "\n" + name + ":")
}
