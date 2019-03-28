package main

import (
	"fmt"
	"log"
	"os"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/app"
)

func main() {

	err := app.New().Run(os.Args)
	if err != nil {
		if errors.Cause(err).Error() == "ui" {
			if e, ok := err.(*errors.Err); ok {
				log.Println(e.Underlying())
				fmt.Println(e.Underlying())
				os.Exit(1)
			}
		}

		fmt.Println(err.Error())
	}
}
