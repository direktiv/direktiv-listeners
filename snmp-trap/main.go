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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	g "github.com/gosnmp/gosnmp"
	"github.com/rs/xid"
)

func main() {

	tl := g.NewTrapListener()
	tl.OnNewTrap = th
	tl.Params = g.Default
	tl.Params.Logger = g.NewLogger(log.New(ioutil.Discard, "", 0))

	log.Printf("starting trap listener")

	err := tl.Listen(cfg.Server.Addr)
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}

}

func th(packet *g.SnmpPacket, addr *net.UDPAddr) {

	guid := xid.New()

	event := cloudevents.NewEvent()
	event.SetSource(cfg.Event.Source)
	event.SetType(cfg.Event.Type)
	event.SetID(guid.String())

	data := make(map[string]string)

	for _, v := range packet.Variables {
		switch v.Type {
		case g.OctetString:
			b := v.Value.([]byte)
			event.Context.SetExtension(strings.ReplaceAll(v.Name, ".", ""), string(b))
			data[v.Name] = string(b)
		default:
			log.Printf("trap hit switch default: %+v\n", v)
		}
	}

	event.SetData(cloudevents.ApplicationJSON, data)
	sendEvent(event)

}

func sendEvent(event cloudevents.Event) {

	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("error marshalling event: %s", err.Error())
		return
	}

	u := fmt.Sprintf("%s/api/namespaces/%s/event", cfg.Direktiv.URL, cfg.Direktiv.Namespace)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		log.Printf("failed to create event request: %s", err.Error())
		return
	}
	defer req.Body.Close()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Direktiv.AuthToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed to send request: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("postEvent returned status code %d", resp.StatusCode)
		return
	}

}
