package main

import (
	"fmt"
	"os"
	"github.com/droundy/go/elf"
	"github.com/droundy/go/x86"
)

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

	x86.Comment("And exit..."),
	x86.Commented(x86.MovL(x86.Imm32(0), x86.EBX), "first argument: exit code"),
	x86.Commented(x86.MovL(x86.Imm32(1), x86.EAX), "system call number (sys_exit)"),
	x86.Int(x86.Imm32(0x80)),
}

func main() {
	fmt.Println(x86.Assembly(hello))
	die(elf.AssembleAndLink("foo", []byte(x86.Assembly(hello))))
}
