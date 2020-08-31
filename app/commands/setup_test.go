package commands_test

import (
	"flag"

	"github.com/codelingo/lingo/app/commands"

	ap "github.com/codelingo/lingo/app"

	"github.com/codelingo/lingo/app/util/testhelper"
	jc "github.com/juju/testing/checkers"
	"github.com/urfave/cli"
	gc "gopkg.in/check.v1"
)

func (s *cmdSuite) TestSetUpCMD(c *gc.C) {
	c.Skip("prompt for username and password")
	app := ap.New()
	set := flag.NewFlagSet("test", 0)
	// TODO(waigani) prompt for username and password
	test := []string{"pwd"}
	set.String("username", "testuser", "ignored")
	set.String("password", "123456", "ignored")
	set.Parse(test)

	ctx := cli.NewContext(app, set, nil)
	// ctx.GlobalString("name")

	setupCMD := testhelper.Command("setup", commands.All())
	c.Assert(setupCMD.Run(ctx), jc.ErrorIsNil)
}
