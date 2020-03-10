package main

import (
	"fmt"
	"time"

	sh "github.com/codeskyblue/go-sh"
)

func main() {
	c := sh.Command("sleep", "3")
	c.Start()
	err := c.WaitTimeout(time.Second * 1)
	if err != nil {
		fmt.Printf("timeout should happend: %v\n", err)
	}
	// timeout should be a session
	out, err := sh.Command("sleep", "2").SetTimeout(time.Second).Output()
	fmt.Printf("output:(%s), err(%v)\n", string(out), err)

	out, err = sh.Command("echo", "hello").SetTimeout(time.Second).Output()
	fmt.Printf("output:(%s), err(%v)\n", string(out), err)
}
