package commands_test

import (
	"flag"

	"github.com/codelingo/lingo/app/commands"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util/testhelper"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *cmdSuite) TestInitCMD(c *gc.C) {
	// TODO(waigani) Do what the skip says.
	c.Skip("This test writes out a codelingo.yaml file in pwd. Test needs to write file to tmpdir and cleanup after.")
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	test := []string{"pwd"}
	set.Parse(test)

	ctx := cli.NewContext(app, set, nil)
	// ctx.GlobalString("name")

	initCMD := testhelper.Command("init", commands.All())
	c.Assert(initCMD.Run(ctx), jc.ErrorIsNil)
}
