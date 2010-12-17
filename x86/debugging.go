package x86

var StartData = []X86{
	Section("data"),
	Symbol("goc.syscall"),
	Commented(GlobalInt(0),
		"This is a global variable for the address for syscalls"),
	Symbol("goc.args"),
	Commented(GlobalInt(0),
		"This is the number of args"),
	Symbol("goc.argsptr"),
	Commented(GlobalInt(0),
		"This is a pointer to the actual args"),

	Symbol("msg"),
	Commented(Ascii("Hello, world!\n"), "a non-null-terminated string"),
	Commented(SymbolicConstant(Symbol("len"), ". - msg"), "length of string"),
}

var StartText = []X86{
	Section("text"),
	Commented(GlobalSymbol("_start"), "this says where to start execution"),
	Call(Symbol("main_main")),
	Comment("And exit..."),
	Commented(MovL(Imm32(0), EBX), "first argument: exit code"),
	Commented(MovL(Imm32(1), EAX), "system call number (sys_exit)"),
	Int(Imm32(0x80)),
}

var Debugging = []X86{
	RawAssembly(`
#  Debug utility routines!

print:
	movl 4(%esp), %edx # read the length
	movl 8(%esp), %ecx # pop the pointer to the string
	movl $2, %ebx	# first argument: file handle (stderr)
	movl $4, %eax	# system call number (sys_write)
	int $128
	popl %eax # store the return address
  addl $8, %esp # get rid of the two arguments
	jmp *%eax # return from println

println:
	movl 4(%esp), %edx # read the length
	movl 8(%esp), %ecx # pop the pointer to the string
	movl $2, %ebx	# first argument: file handle (stderr)
	movl $4, %eax	# system call number (sys_write)
	int $128
  # now we want to write a newline...
  movl $10, 4(%esp) # a newline
	movl $1, %edx # the length
	movl %esp, %ecx # the pointer
  addl $4, %ecx
	movl $2, %ebx	# first argument: file handle (stderr)
	movl $4, %eax	# system call number (sys_write)
  int $128
	popl %eax # store the return address
  addl $8, %esp # get rid of the two arguments
	jmp *%eax # return from println

debug.print_eax:
	pushl %edx	# Save registers...
	pushl %ecx
	pushl %ebx
	pushl %eax
	movl %esp, %ecx	# second argument: pointer to data
	addl $-20, %ecx	# here I set up %ecx as my string pointer
	movl %eax, %ebx
	andl $15, %ebx
	shll $24, %ebx
	addl $805306368, %ebx	# least significant hex
	movl %eax, %edx
	andl $240, %edx
	addl $768, %edx
	shll $12, %edx
	addl %edx, %ebx	# second most significant hex
	movl %eax, %edx
	andl $3840, %edx
	addl $12288, %edx
	addl %edx, %ebx	# third most significant hex
	movl %eax, %edx
	andl $61440, %edx
	addl $196608, %edx
	shrl $12, %edx
	addl %edx, %ebx	# fourth most significant hex
	movl %ebx, 8(%ecx)	# Store four bytes of hex notation, which covers 16 bits of EAX
	movl %eax, %edx
	andl $983040, %edx
	addl $3145728, %edx
	shrl $16, %edx
	movl %edx, %ebx	# fifth most significant hex
	movl %eax, %edx
	andl $15728640, %edx
	addl $50331648, %edx
	shrl $12, %edx
	addl %edx, %ebx	# sixth most significant hex
	movl %eax, %edx
	andl $251658240, %edx
	addl $805306368, %edx
	shrl $8, %edx
	addl %edx, %ebx	# seventh most significant hex
	movl %eax, %edx
	shrl $4, %edx
	andl $251658240, %edx
	addl $805306368, %edx
	addl %edx, %ebx	# eighth most significant hex
	movl %ebx, 4(%ecx)	# Store four more bytes of hex notation, which covers the last 16 bits of EAX
	movl $10, 12(%ecx)	# Add newline
	movl $980967781, (%ecx)	# Add prefix
	movl $13, %edx	# third argument: data length
	movl $1, %ebx	# first argument: file handle (stdout)
	movl $4, %eax	# system call number (sys_write)
	int $128
	popl %eax	# Restore saved registers...
	popl %ebx
	popl %ecx
	popl %edx
	ret	# from debug.print_eax
		`),
}
