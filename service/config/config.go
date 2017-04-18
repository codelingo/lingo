package config

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/juju/errors"

	"gopkg.in/yaml.v2"

	"os"
	"github.com/hishboy/gocommons/lang"
)

func (c *Config) GetEnv() (string, error) {
	// return test when running go test
	if isTest, _ := regexp.MatchString("/_test/", os.Args[0]); isTest {
		return "test", nil
	}

	env, err := ioutil.ReadFile(c.envFile)
	if err != nil {
		if strings.Contains(err.Error(), "open /home/dev/.codelingo/configs/lingo-current-env: no such file or directory") {
			return "", errors.New("No lingo environment set. Please run `lingo use-env <env>` to set the environment.")
		}

		return "", errors.Trace(err)
	}

	trimmedEnv := strings.TrimSpace(string(env))
	return trimmedEnv, nil
}

func (c *Config) SetEnv(env string) error {
	err := ioutil.WriteFile(c.envFile, []byte(env), 0644)
	if err != nil {
		 return errors.Trace(err)
	}
	return nil
}

// TODO: switch Config to an interface type and refactor
type Config struct {
	envFile string
}

type FileConfig struct {
	config   *Config
	data     map[string]interface{}
	filename string
}

func New(envFile string) *Config {
	return &Config{
		envFile,
	}
}

func (c *Config) New(cfgFile string) (*FileConfig, error) {
	data, err := readYaml(cfgFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &FileConfig{
		config:   c,
		data:     data,
		filename: cfgFile,
	}, nil
}

func readYaml(cfgFile string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, errors.Errorf("problem reading %s: %v", cfgFile, err)
	}
	mapData := make(map[string]interface{})
	err = yaml.Unmarshal(data, mapData)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return mapData, nil
}

func (c *Config) Create(cfgFile string, data interface{}, perm os.FileMode) (*FileConfig, error) {
	var out []byte
	var err error
	if data != nil {
		out, err = yaml.Marshal(data)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	err = ioutil.WriteFile(cfgFile, out, perm)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &FileConfig{
		config:   c,
		data:     make(map[string]interface{}),
		filename: cfgFile,
	}, nil
}

type cfgInfo struct {
	info map[interface{}]interface{}
}

type yamlNode struct {
	depth int
	key string
	value interface{}
}

// TODO(waigani) err handling
func (c cfgInfo) walk(keyPath []string) map[string]interface{} {
	keyMap := make(map[string]interface{})
	stack := lang.NewStack()

	for k, v := range c.info {
		kString, ok := k.(string)
		if !ok {
			continue
		}
		newNode := &yamlNode{
			0,
			kString,
			v,
		}
		stack.Push(newNode)
	}

	for stack.Len() > 0 {
		node, ok := stack.Pop().(*yamlNode)
		if !ok {
			continue
		}

		if node.depth == len(keyPath) {
			keyMap[node.key] = node.value
			continue
		}

		value, ok := node.value.(map[interface{}]interface{})
		if !ok {
			continue
		}

		for k, v := range value {
			kString, ok := k.(string)
			if !ok {
				continue
			}

			if kString == keyPath[node.depth] {
				newNode := &yamlNode{
					node.depth + 1,
					node.key + "." + kString,
					v,
				}
				stack.Push(newNode)
			}
		}

	}

	return keyMap
}

func (c cfgInfo) walkSet(keyPath []string, value interface{}) error {
	keyLen := len(keyPath)
	switch {
	case keyLen < 1:
		return errors.New("key path is empty")
	case keyLen > 1:
		if c.info[keyPath[0]] == nil {
			c.info[keyPath[0]] = make(map[interface{}]interface{})
		}

		var ok bool
		c.info, ok = c.info[keyPath[0]].(map[interface{}]interface{})
		if !ok {
			return errors.Errorf("malformed config file. Expected map[interface{}]interface{}, got %T", c.info[keyPath[0]])
		}
		return c.walkSet(keyPath[1:], value)
	}
	c.info[keyPath[0]] = value
	return nil
}

func newCfgInfo(infoMap interface{}) (*cfgInfo, error) {
	if infoMap == nil {
		return nil, errors.New("infoMap is nil")
	}
	iMap := make(map[interface{}]interface{})
	if m, ok := infoMap.(map[interface{}]interface{}); ok {
		iMap = m
		// in the case of setting values the raw Config is passed through.
	} else if m, ok := infoMap.(map[string]interface{}); ok {
		for k, v := range m {
			iMap[k] = v
		}
	} else {
		return nil, errors.Errorf("unknown type for infoMap %T", infoMap)
	}
	return &cfgInfo{
		info: iMap,
	}, nil
}

func (fc *FileConfig) GetAll(key string) (map[string]interface{}, error) {
	keys := strings.Split(key, ".")

	infoM, err := newCfgInfo(fc.data)
	if err != nil {
		return nil, errors.Trace(err)
	}

	keysMap := infoM.walk(keys)

	return keysMap, nil
}

func (fc *FileConfig) GetValue(key string) (string, error) {
	env, err := fc.config.GetEnv()
	if err != nil {
		return "", errors.Trace(err)
	}

	keysMap, err := fc.GetAll(key)
	if err != nil {
		return "", errors.Trace(err)
	}

	var allVal string = ""
	var valErr error
	for k, v := range keysMap {
		keyEnv := strings.Split(k, ".")[0]
		switch keyEnv {
		case env:
			kString, ok := v.(string)
			if !ok {
				return "", errors.Errorf("Invalid value found for config %q, expected `string` but got `%T`", key, v)
			}
			return kString, nil
		case "all":
			kString, ok := v.(string)
			if !ok {
				valErr = errors.Errorf("Invalid value found for config %q, expected `string` but got `%T`", key, v)
			} else {
				allVal = kString
			}
		default:
			continue
		}
	}

	if allVal == "" {
		if valErr != nil {
			return "", valErr
		} else {
			return "", errors.Errorf("Could not find value for config %q", key)
		}
	} else {
		return allVal, nil
	}
}

func (fc *FileConfig) GetForEnv(env, key string) (interface{}, error) {
	keysMap, err := fc.GetAll(key)
	if err != nil {
		return "", errors.Trace(err)
	}

	for k, v := range keysMap {
		keyEnv := strings.Split(k, ".")[0]
		switch keyEnv {
		case env:
			return v, nil
		default:
			continue
		}
	}

	return "", errors.Errorf("Could not find value for config %q", key)
}

func (fc *FileConfig) SetForEnv(env string, key string, value interface{}) error {
	// Prepend the env to the given key
	key = env+"."+key

	mapData, err := readYaml(fc.filename)
	if err != nil {
		return errors.Trace(err)
	}
	keys := strings.Split(key, ".")
	infoM, err := newCfgInfo(mapData)
	if err != nil {
		return errors.Trace(err)
	}

	err = infoM.walkSet(keys, value)
	if err != nil {
		return errors.Trace(err)
	}

	data, err := yaml.Marshal(infoM.info)
	if err != nil {
		return errors.Trace(err)
	}

	fc.data = convertMapType(infoM.info)

	err = ioutil.WriteFile(fc.filename, data, 0755)
	if err != nil {
		return errors.Trace(err)
	}

	err = fc.Reload()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (fc *FileConfig) Set(key string, value interface{}) error {
	env, err := fc.config.GetEnv()
	if err != nil {
		return errors.Trace(err)
	}

	return fc.SetForEnv(env, key, value)
}

func (fc *FileConfig) Reload() error {
	newFc, err := fc.config.New(fc.filename)
	if err != nil {
		return errors.Trace(err)
	}
	fc.data = newFc.data
	return nil
}

func convertMapType(m map[interface{}]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range m {
		newMap[k.(string)] = v
	}
	return newMap
}
