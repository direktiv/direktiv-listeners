package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Linode struct {
		LinodeAuthToken string `yaml:"linodeAuthToken"`
	} `yaml:"linode"`
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
		log.Fatal("Linode receiver needs a configuration file")
	}

	readConfig(os.Args[1])

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: gcConfig.Linode.LinodeAuthToken})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linodeClient := linodego.NewClient(oauth2Client)
	// linodeClient.SetDebug(true)

	m := make(map[string]bool)

	for {

		timeNow := time.Now().Add(-time.Second * 10).UTC()

		f := linodego.Filter{}
		f.AddField(linodego.Gte, "created", timeNow.Format("2006-01-02T15:04:05"))
		fStr, err := f.MarshalJSON()

		if err != nil {
			log.Fatal(err)
		}

		listOpts := linodego.NewListOptions(0, string(fStr))

		events, err := linodeClient.ListEvents(context.Background(), listOpts)

		if err != nil {
			log.Printf("Failed to get events list: %v", err)
			continue
		}

		for _, event := range events {

			if _, seen := m[strconv.Itoa(event.ID)]; !seen {

				//log.Printf("Event: %s / %s / %s", event.Action, event.Created, strconv.Itoa(event.ID))

				ce := cloudevents.NewEvent()
				id := uuid.New()
				ce.SetID(id.String())
				//ce.SetID(event.ID)
				ce.SetSource("direktiv/listener/linode")
				ce.SetType(string(event.Action))
				ce.SetTime(*event.Created)
				err = ce.SetData(event)
				if err != nil {
					log.Printf("Failed to set cloudevent data: %v", err)
					continue
				}

				data, err := ce.MarshalJSON()
				if err != nil {
					log.Printf("Failed to marshal cloudevent: %v", err)
					continue
				}

				fmt.Printf("%s,\n", data)

				eventTime := event.Created

				if eventTime.After(timeNow) {
					err = sendCloudEvent(ce)
					if err != nil {
						log.Printf("Failed to send cloudevent: %v", err)
						continue
					}
				}

				m[strconv.Itoa(event.ID)] = true

			}

		}

		time.Sleep(2 * time.Second)

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
