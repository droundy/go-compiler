package main

import (
	"fmt"
	"os"
	"github.com/droundy/go/elf"
)

func die(err os.Error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	die(elf.AssembleAndLink("foo", []byte(`
.data

msg:
	.ascii	"Hello, world!\n"	# our dear string
	len = . - msg			# length of our dear string

.text

	              # we must export the entry point to the ELF linker or
.global _start	# loader. They conventionally recognize _start as their
	              # entry point. Use ld -e foo to override the default.

_start:

# write our string to stdout

	movl	$len,%edx	# third argument: message length
	movl	$msg,%ecx	# second argument: pointer to message to write
	movl	$1,%ebx		# first argument: file handle (stdout)
	movl	$4,%eax		# system call number (sys_write)
	int	$0x80		# call kernel

# and exit

	movl	$0,%ebx		# first argument: exit code
	movl	$1,%eax		# system call number (sys_exit)
	int	$0x80		# call kernel
`)))
}
