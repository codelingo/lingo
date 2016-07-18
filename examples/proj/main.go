package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) >= 2 {
		dl := Download{}
		dl.GetURL(os.Args[1])
		fmt.Println(dl.CRC32())
	}
	return
}
