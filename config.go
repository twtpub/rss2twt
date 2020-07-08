package main

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Root    string
	BaseURL string
	Feeds   map[string]string // name -> url

	path string // path to config file that was loaded used by .Save()
}

func (conf *Config) Parse(data []byte) error {
	return yaml.Unmarshal(data, conf)
}

func (conf *Config) Save() error {
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(conf.path, data, 0644)
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if err := conf.Parse(data); err != nil {
		return nil, err
	}
	conf.path = filename
	return conf, nil
}
