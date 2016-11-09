package config_test

import (
	"testing"

	"github.com/codelingo/lingo/service/config"
	jc "github.com/juju/testing/checkers"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type suite struct{}

var _ = Suite(&suite{})

func (s *suite) TestGetCfg(c *C) {
	c.Skip("Assert address")
	cfg, err := config.New("test_cfg.yaml")
	c.Assert(err, jc.ErrorIsNil)

	addr, err := cfg.Get("gitserver.remote.name")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(addr, DeepEquals, "")
}

func (s *suite) TestSetCfg(c *C) {
	cfg, err := config.New("test_cfg.yaml")
	c.Assert(err, jc.ErrorIsNil)

	err = cfg.Set("blah.gitserver.remote.name.nested.super", "new-name")
	err = cfg.Set("blah.gitserver.remote.name.nested.x", "b")
	c.Assert(err, jc.ErrorIsNil)

	// TODO(waigani) assert config
}
