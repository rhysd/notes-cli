package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Output from stdout")
	fmt.Fprintln(os.Stderr, "Output from stderr")
	fmt.Println(os.Args[1:])
}
