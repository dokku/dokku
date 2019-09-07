package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	// spawn an orphaned zombie
	err := exec.Command("/bin/bash", "-c", "/bin/false &").Run()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", hello)
	fmt.Println("listening...")
	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "go")
}
