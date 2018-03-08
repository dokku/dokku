package copy

import (
	"fmt"
	"os"
)

func ExampleCopy() {

	os.MkdirAll("testdata/example/foo/bar", os.ModePerm)
	defer os.RemoveAll("testdata")

	err := Copy("testdata/example", "testdata/example.copy")
	fmt.Println("Error:", err)
	info, _ := os.Stat("testdata/example.copy/foo/bar")
	fmt.Println("IsDir:", info.IsDir())

	// Output:
	// Error: <nil>
	// IsDir: true
}
