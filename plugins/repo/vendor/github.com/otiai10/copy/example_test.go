package copy

import (
	"fmt"
	"os"
	"strings"
)

func ExampleCopy() {

	err := Copy("testdata/example", "testdata.copy/example")
	fmt.Println("Error:", err)
	info, _ := os.Stat("testdata.copy/example")
	fmt.Println("IsDir:", info.IsDir())

	// Output:
	// Error: <nil>
	// IsDir: true
}

func ExampleOptions() {

	err := Copy(
		"testdata/example",
		"testdata.copy/example_with_options",
		Options{
			Skip: func(src string) bool {
				return strings.HasSuffix(src, ".git-like")
			},
			OnSymlink: func(src string) SymlinkAction {
				return Skip
			},
			AddPermission: 0200,
		},
	)
	fmt.Println("Error:", err)
	_, err = os.Stat("testdata.copy/example_with_options/.git-like")
	fmt.Println("Skipped:", os.IsNotExist(err))

	// Output:
	// Error: <nil>
	// Skipped: true

}
