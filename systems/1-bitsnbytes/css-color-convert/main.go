package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	srcPath = flag.String("src_path", "", "path to source .css file")
)

func main() {
	flag.Parse()

	file, err := os.Open(*srcPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	bs, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("bs: %#02v", bs)
}
