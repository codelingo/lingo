package main

import (
	"os/exec"
)

func main() {
	cmd := exec.Command("echo", "hello")

	byt, err := cmd.CombinedOutput()
	if err != nil {
		panic(err.Error())
	}

	cmd.Run()

	print(string(byt))
}
