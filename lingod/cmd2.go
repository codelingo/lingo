package main

import (
	"os"
	"os/exec"
)

func main() {

	cmd := exec.Command("go", "run", "cmd1.go", "out")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}
