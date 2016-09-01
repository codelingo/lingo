package config

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/juju/errors"

	"gopkg.in/yaml.v2"

	"os"
)

// ENV will return environment
func ENV() string {

	// return test when running go test
	if isTest, _ := regexp.MatchString("/_test/", os.Args[0]); isTest {
		return "test"
	}

	if env := os.Getenv("CODELINGO_ENV"); env != "" {
		return env
	}
	return "all"
}

type Config map[string]interface{}

func New(cfgFile string) (*Config, error) {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, errors.Errorf("problem reading %s: %v", cfgFile, err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return cfg, nil
}

type cfgInfo struct {
	info map[interface{}]interface{}
}

// TODO(waigani) err handling
// TODO(waigani) not concurrently safe
func (c cfgInfo) walk(keyPath []string) (string, error) {
	if l := len(keyPath); l > 1 {
		if c.info[keyPath[0]] == nil {
			return "", errors.Errorf("config %q not found", strings.Join(keyPath, "."))
		}

		var ok bool
		c.info, ok = c.info[keyPath[0]].(map[interface{}]interface{})
		if !ok {
			return "", errors.Errorf("malformed config file. Expected map[interface{}]interface{}, got %T", c.info[keyPath[0]])
		}
		return c.walk(keyPath[1:])
	}

	if result := c.info[keyPath[0]]; result != nil {
		if r, ok := result.(string); ok {
			return r, nil
		}
		return "", errors.Errorf("config %q is not a string", strings.Join(keyPath, "."))
	}

	return "", errors.Errorf("config %q not found", strings.Join(keyPath, "."))
}

func newCfgInfo(infoMap interface{}) (*cfgInfo, error) {
	if infoMap == nil {
		return nil, errors.New("infoMap is nil")
	}
	iMap := infoMap.(map[interface{}]interface{})
	return &cfgInfo{
		info: iMap,
	}, nil
}

func (c Config) Get(key string) (string, error) {
	keys := strings.Split(key, ".")
	var infoBlocks []*cfgInfo

	// first get info blocks
	if env := ENV(); env != "all" && env != "" {
		infoM, err := newCfgInfo(c[env])
		if err == nil {
			// TODO(waigani) log
			infoBlocks = append(infoBlocks, infoM)
		}
	}
	infoM, err := newCfgInfo(c["all"])
	if err == nil {
		// TODO(waigani) log
		infoBlocks = append(infoBlocks, infoM)
	}

	var val string
	for _, inf := range infoBlocks {
		val, err = inf.walk(keys)
		if err == nil && val != "" {
			return val, nil
		}
	}
	if err != nil {
		return "", errors.Trace(err)
	}
	return "", errors.Errorf("config %q not found", key)
}
