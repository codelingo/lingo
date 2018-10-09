package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func main() {

	fmt.Print(os.Args[1:])
	var options string
	fmt.Scanln(&options)

	m := color.New(color.FgWhite, color.Faint).SprintfFunc()

	fmt.Print(m(options))

}
