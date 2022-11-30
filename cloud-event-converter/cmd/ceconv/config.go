package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"

	"github.com/pelletier/go-toml"
)

var cfg *config

func init() {

	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		log.Fatal("must provide --config flag")
	}

	cfg = new(config)
	err := cfg.fromFile(cfgPath)
	if err != nil {
		log.Fatal(err.Error())
	}

}

type config struct {
	Server struct {
		Bind string `toml:"bind"`
		TLS  struct {
			Enabled bool   `toml:"enabled"`
			Cert    string `toml:"cert"`
			Key     string `toml:"key"`
		}
	} `toml:"server"`
	Rules []rule `toml:"rules"`
}

type rule struct {
	Endpoint  string   `toml:"direktivEndpoint"`
	Condition string   `toml:"condition"`
	Modifiers []string `toml:"modifiers"`
}

func (c *config) fromFile(path string) error {

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return c.load(b)
}

func (c *config) load(b []byte) error {

	err := toml.NewDecoder(bytes.NewReader(b)).Decode(c)
	if err != nil {
		return err
	}

	return nil
}
