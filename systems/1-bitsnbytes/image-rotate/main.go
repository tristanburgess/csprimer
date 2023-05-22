package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	srcPath = flag.String("src_file", "", "the path to the source BMP file")
)

type BMP struct {
}

func rotateBMP(bmp *BMP) {

}

func serializeBMP(bmp BMP) []byte {
	bs := make([]byte, 0)

	return bs
}

func deserializeBMP(bs []byte) BMP {
	bmp := BMP{}

	return bmp
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	file, err := os.Open(*srcPath)
	check(err)
	bs, err := io.ReadAll(file)
	check(err)

	bmp := deserializeBMP(bs)
	rotateBMP(&bmp)
	bs = serializeBMP(bmp)

	file, err = os.Create(fmt.Sprintf("%s/%s", filepath.Base(*srcPath), "rotated.bmp"))
	check(err)
	_, err = file.Write(bs)
	check(err)
}
