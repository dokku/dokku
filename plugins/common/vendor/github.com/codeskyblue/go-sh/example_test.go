package sh_test

import (
	"fmt"

	"github.com/codeskyblue/go-sh"
)

func ExampleCommand() {
	out, err := sh.Command("echo", "hello").Output()
	fmt.Println(string(out), err)
}

func ExampleCommandPipe() {
	out, err := sh.Command("echo", "-n", "hi").Command("wc", "-c").Output()
	fmt.Println(string(out), err)
}

func ExampleCommandSetDir() {
	out, err := sh.Command("pwd", sh.Dir("/")).Output()
	fmt.Println(string(out), err)
}

func ExampleTest() {
	if sh.Test("dir", "mydir") {
		fmt.Println("mydir exists")
	}
}
