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
	x86.Symbol("msg"),
	x86.Commented(x86.Ascii("Hello, world!\n"), "a non-null-terminated string"),
	x86.Commented(x86.SymbolicConstant(x86.Symbol("len"), ". - msg"), "length of string"),
	x86.Section("text"),
	x86.Commented(x86.GlobalSymbol("_start"), "this says where to start execution"),
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
