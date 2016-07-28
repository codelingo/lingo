package config_test

import (
	"testing"

	"github.com/codelingo/lingo/service/config"
	jc "github.com/juju/testing/checkers"
	"github.com/waigani/xxx"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type suite struct{}

var _ = Suite(&suite{})

func (s *suite) TestCfg(c *C) {

	cfg, err := config.New("test_cfg.yaml")
	c.Assert(err, jc.ErrorIsNil)

	addr, err := cfg.Get("gitserver.remote.name")
	c.Assert(err, jc.ErrorIsNil)
	xxx.Print(addr)

}
