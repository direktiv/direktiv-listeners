// Copyright 2012 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

/*
The developer of the trapserver code (https://github.com/jda) says "I'm working
on the best level of abstraction but I'm able to receive traps from a Cisco
switch and Net-SNMP".
Pull requests welcome.
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/sisatech/toml"
)

var cfg *config

type config struct {
	Server struct {
		Addr string `toml:"addr"`
	} `toml:"server"`
	Direktiv struct {
		URL       string `toml:"url"`
		Namespace string `toml:"namespace"`
		AuthToken string `toml:"authToken"`
	} `toml:"direktiv"`
	Event struct {
		Source string `toml:"source"`
		Type   string `toml:"type"`
	} `toml:"event"`
}

func init() {

	cfg = new(config)

	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		log.Fatal("please provide the path to a config file using the --config flag")
	}

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = toml.Unmarshal(b, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = cfg.validate()
	if err != nil {
		log.Fatal(err.Error())
	}

}

func (c *config) validate() error {

	if c.Direktiv.URL == "" {
		return fmt.Errorf("no direktiv url specified in config")
	}

	if c.Direktiv.Namespace == "" {
		return fmt.Errorf("no direktiv namespace specified in config")
	}

	if c.Direktiv.AuthToken == "" {
		return fmt.Errorf("no direktiv authentication token specified in config")
	}

	if c.Event.Source == "" {
		return fmt.Errorf("no cloud event source specified in config")
	}

	if c.Event.Type == "" {
		return fmt.Errorf("no cloud event type specified in config")
	}

	return nil
}
