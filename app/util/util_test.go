package util_test

import (
	"testing"

	"github.com/codelingo/lingo/app/util"
	"github.com/waigani/xxx"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type utilSuite struct{}

var _ = Suite(&utilSuite{})

func (s *utilSuite) TestConfig(c *C) {
	cfg, err := util.LingoConfig()
	c.Assert(err, jc.ErrorIsNil)
	xxx.Dump(cfg)
}
