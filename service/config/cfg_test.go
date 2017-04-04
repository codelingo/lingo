package config_test

import (
	"testing"

	"github.com/codelingo/lingo/service/config"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	jc "github.com/juju/testing/checkers"
	. "gopkg.in/check.v1"
	"github.com/codelingo/lingo/app/util"
	"path/filepath"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type suite struct{}

var _ = Suite(&suite{})

func (s *suite) TestGetCfg(c *C) {
	c.Skip("Assert address")

	configHome, err := util.ConfigHome()
	c.Assert(err, jc.ErrorIsNil)
	envFilepath := filepath.Join(configHome, commonConfig.EnvCfgFile)

	cfg := config.New(envFilepath)
	testCfg, err := cfg.New("test_cfg.yaml")
	c.Assert(err, jc.ErrorIsNil)

	addr, err := testCfg.GetValue("gitserver.remote.name")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(addr, DeepEquals, "")
}

func (s *suite) TestSetCfg(c *C) {
	configHome, err := util.ConfigHome()
	c.Assert(err, jc.ErrorIsNil)
	envFilepath := filepath.Join(configHome, commonConfig.EnvCfgFile)

	cfg := config.New(envFilepath)
	testCfg, err := cfg.New("test_cfg.yaml")
	c.Assert(err, jc.ErrorIsNil)

	err = testCfg.Set("blah.gitserver.remote.name.nested.super", "new-name")
	err = testCfg.Set("blah.gitserver.remote.name.nested.x", "b")
	c.Assert(err, jc.ErrorIsNil)

	// TODO(waigani) assert config
}
