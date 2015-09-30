package kat

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
}

func NewConfig() *Config {
	c := Config{}
	c.initialize()
	return &c
}

func (c *Config) initialize() {
}

func (c *Config) Load(configfile string) error {
	_, err := toml.DecodeFile(configfile, c)
	return err
}

func (c *Config) String() string {
	return "kat"
}
