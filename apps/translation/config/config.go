package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Addr    string `yaml:"port"`
	RunMode string `yaml:"run_mode"`
}

type redisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Pass string `yaml:"pass"`
	Db   int    `yaml:"db"`
}

type mysqlConfig struct {
	Dsn      string `yaml:"dsn"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Db       string `yaml:"db"`
}

var c *Config

func GetConfig() *Config {
	return c
}

func ParseConfig(dist string) error {
	data, err := ioutil.ReadFile(dist)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析 YAML 失败: %w", err)
	}
	c = &config
	return nil
}
