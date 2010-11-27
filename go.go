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
	fmt.Println("Creating a.out...")
	o,err := os.Open("a.out", os.O_WRONLY + os.O_CREAT + os.O_TRUNC, 0755)
	die(err)
	var h = elf.Header{ 4*elf.Page, []byte{}, []byte{} }
	die(h.WriteTo(o))
}
