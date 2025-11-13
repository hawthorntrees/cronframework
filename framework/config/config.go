package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

var (
	_config *Config
)

func Init(filePath string) *Config {
	cfg := load(filePath)
	cfg.completeDefaults()
	_config = cfg
	return cfg
}
func load(filePath string) *Config {
	path, err := getExecPath(filePath)
	if err == nil {
		filePath = path
	}
	_, err = os.Stat(filePath)
	var cfg Config
	if err == nil {
		file, err := os.ReadFile(filePath)
		if err != nil {
			panic(fmt.Errorf("加载配置文件失败%s,%v", filePath, err))
		}
		if err := yaml.Unmarshal(file, &cfg); err != nil {
			panic(fmt.Errorf("解析配置文件失败%s,%v", filePath, err))
		}
	} else if os.IsNotExist(err) {
		panic(fmt.Errorf("未找到配置文件: %s", filePath))
	}
	return &cfg
}

func (c *Config) completeDefaults() {
	completeServer(c)
	completeLog(c)
	completeCronTask(c)
	completeDatabases(c)
}
func GetTokenKey() string {
	return _config.Server.TokenKey
}
func GetExpireTime() time.Time {
	return time.Now().Add(_config.Server.SessionExpires)
}
func GetBasePath() string {
	return _config.Server.BashPath
}
func getExecPath(relativeFilePath string) (execPath string, err error) {
	ep, e := os.Executable()
	if e != nil {
		return "", e
	}
	p := filepath.Join(filepath.Dir(ep), relativeFilePath)
	_, e = os.Stat(p)
	if e != nil {
		return "", e
	}
	return p, nil
}
