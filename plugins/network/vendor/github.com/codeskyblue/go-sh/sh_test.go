package sh

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"testing"
)

func TestAlias(t *testing.T) {
	s := NewSession()
	s.Alias("gr", "echo", "hi")
	out, err := s.Command("gr", "sky").Output()
	if err != nil {
		t.Error(err)
	}
	if string(out) != "hi sky\n" {
		t.Errorf("expect 'hi sky' but got:%s", string(out))
	}
}

func ExampleSession_Command() {
	s := NewSession()
	out, err := s.Command("echo", "hello").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
	// Output: hello
}

func ExampleSession_Command_pipe() {
	s := NewSession()
	out, err := s.Command("echo", "hello", "world").Command("awk", "{print $2}").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
	// Output: world
}

func ExampleSession_Alias() {
	s := NewSession()
	s.Alias("alias_echo_hello", "echo", "hello")
	out, err := s.Command("alias_echo_hello", "world").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
	// Output: hello world
}

func TestEcho(t *testing.T) {
	out, err := Echo("one two three").Command("wc", "-w").Output()
	if err != nil {
		t.Error(err)
	}
	if strings.TrimSpace(string(out)) != "3" {
		t.Errorf("expect '3' but got:%s", string(out))
	}
}

func TestSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("ignore test on windows")
		return
	}
	session := NewSession()
	session.ShowCMD = true
	err := session.Call("pwd")
	if err != nil {
		t.Error(err)
	}
	out, err := session.SetDir("/").Command("pwd").Output()
	if err != nil {
		t.Error(err)
	}
	if string(out) != "/\n" {
		t.Errorf("expect /, but got %s", string(out))
	}
}

/*
	#!/bin/bash -
	#
	export PATH=/usr/bin:/bin
	alias ll='ls -l'
	cd /usr
	if test -d "local"
	then
		ll local | awk '{print $1, $NF}' | grep bin
	fi
*/
func Example(t *testing.T) {
	s := NewSession()
	//s.ShowCMD = true
	s.Env["PATH"] = "/usr/bin:/bin"
	s.SetDir("/bin")
	s.Alias("ll", "ls", "-l")

	if s.Test("d", "local") {
		//s.Command("ll", []string{"local"}).Command("awk", []string{"{print $1, $NF}"}).Command("grep", []string{"bin"}).Run()
		s.Command("ll", "local").Command("awk", "{print $1, $NF}").Command("grep", "bin").Run()
	}
}
