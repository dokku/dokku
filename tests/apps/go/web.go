package main

import (
    "fmt"
    "net/http"
    "os"
)

func main() {
    http.HandleFunc("/", hello)
    fmt.Println("listening...")
    err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
    if err != nil {
      panic(err)
    }
}

// hello handles HTTP requests to the root path and responds with "go".
func hello(res http.ResponseWriter, req *http.Request) {
    fmt.Fprintln(res, "go")
}
