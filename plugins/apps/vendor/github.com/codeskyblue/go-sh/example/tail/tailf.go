package main

import (
	"flag"
	"fmt"

	"github.com/codeskyblue/go-sh"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Usage: PROGRAM <file>")
		return
	}
	sh.Command("tail", "-f", flag.Arg(0)).Run()
}
