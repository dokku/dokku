package main

import "github.com/codeskyblue/go-sh"

func main() {
	sh.Command("less", "less.go").Run()
}
