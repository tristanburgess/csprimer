package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	prevState, err := termMakeCBreak(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer restore(int(os.Stdin.Fd()), prevState)
	for {
		b := make([]byte, 1)
		_, err = os.Stdin.Read(b)
		if err != nil {
			return
		}
		if b[0] >= byte('0') && b[0] <= byte('9') {
			for i := 0; i < int(b[0]-'0'); i++ {
				os.Stdout.Write([]byte{byte('\a')})
				time.Sleep(1 * time.Second)
			}
		}
	}
}
