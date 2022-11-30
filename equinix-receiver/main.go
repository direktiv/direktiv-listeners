package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
	"github.com/packethost/packngo"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Equinix struct {
		OrganizationID  string `yaml:"organizationId"`
		PacketAuthToken string `yaml:"packetAuthToken"`
	} `yaml:"equinix"`
	Direktiv struct {
		DirektivEndpoint   string `yaml:"endpoint"`
		Token              string `yaml:"token"`
		ApiKey             string `yaml:"apikey"`
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	} `yaml:"direktiv"`
}

var gcConfig Config

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Equinix receiver needs a configuration file")
	}

	readConfig(os.Args[1])

	client, err := packngo.NewClient(packngo.WithAuth("packngo lib", gcConfig.Equinix.PacketAuthToken))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Equinix Organization ID: < %s >", gcConfig.Equinix.OrganizationID)

	m := make(map[string]bool)

	timeNow := time.Now()

	for {

		listOpts := &packngo.ListOptions{}

		events, response, err := client.Organizations.ListEvents(gcConfig.Equinix.OrganizationID, listOpts)
		if err != nil {
			log.Fatal(err)
		}
		_ = response.Body.Close()

		for _, event := range events {

			if _, seen := m[event.ID]; !seen {

				log.Printf("Event: %s / %s / %s", event.ID, event.Type, event.Interpolated)

				ce := cloudevents.NewEvent()
				id := uuid.New()
				ce.SetID(id.String())
				//ce.SetID(event.ID)
				ce.SetSource("direktiv/listener/equinix")
				ce.SetType(event.Type)
				ce.SetData(event)

				data, err := ce.MarshalJSON()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%s", data)

				eventTime := event.CreatedAt

				if eventTime.After(timeNow) {
					err = sendCloudEvent(ce)
					if err != nil {
						log.Fatal(err)
					}
				}
				m[event.ID] = true

			}

		}

		time.Sleep(5 * time.Second)

	}

}

func sendCloudEvent(event cloudevents.Event) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: gcConfig.Direktiv.InsecureSkipVerify},
	}

	options := []cehttp.Option{
		cloudevents.WithTarget(gcConfig.Direktiv.DirektivEndpoint),
		cloudevents.WithStructuredEncoding(),
		cloudevents.WithHTTPTransport(tr),
	}

	if len(gcConfig.Direktiv.Token) > 0 {
		log.Printf("Using token to login...")
		options = append(options,
			cehttp.WithHeader("Direktiv-Token", gcConfig.Direktiv.Token))
	} else if len(gcConfig.Direktiv.ApiKey) > 0 {
		log.Printf("Using apikey to login...")
		options = append(options,
			cehttp.WithHeader("apikey", gcConfig.Direktiv.ApiKey))
	}

	t, err := cloudevents.NewHTTPTransport(
		options...,
	)
	if err != nil {
		return err
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Printf("Unable to create CloudEvent client: " + err.Error())
		return err
	}

	_, _, err = c.Send(context.Background(), event)
	if err != nil {
		log.Printf("Unable to send CloudEvent: " + err.Error())
		return err
	}

	return nil

}

func readConfig(cfile string) {

	// Open config file
	file, err := os.Open(cfile)
	if err != nil {
		log.Fatal("Can not find file or we do not have permission to read")
		return
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	if err := d.Decode(&gcConfig); err != nil {
		log.Fatal("Failed decoding config file")
	}

}
