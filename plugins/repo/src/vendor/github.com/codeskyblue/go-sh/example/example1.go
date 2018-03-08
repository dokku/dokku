package main

import (
	"fmt"
	"log"

	"github.com/codeskyblue/go-sh"
)

func main() {
	sh.Command("echo", "hello").Run()
	out, err := sh.Command("echo", "hello").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output is", string(out))

	var a int
	sh.Command("echo", "2").UnmarshalJSON(&a)
	fmt.Println("a =", a)

	s := sh.NewSession()
	s.Alias("hi", "echo", "hi")
	s.Command("hi", "boy").Run()

	fmt.Print("pwd = ")
	s.Command("pwd", sh.Dir("/")).Run()

	if !sh.Test("dir", "data") {
		sh.Command("echo", "mkdir", "data").Run()
	}

	sh.Command("echo", "hello", "world").
		Command("awk", `{print "second arg is "$2}`).Run()
	s.ShowCMD = true
	s.Command("echo", "hello", "world").
		Command("awk", `{print "second arg is "$2}`).Run()

	s.SetEnv("BUILD_ID", "123").Command("bash", "-c", "echo $BUILD_ID").Run()
	s.Command("bash", "-c", "echo current shell is $SHELL").Run()
}
