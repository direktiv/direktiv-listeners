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

type Extension struct {
	OrganizationName string
	OrganizationId   string
	ProjectName      string
	ProjectId        string
	Hostname         string
	// IPAddress        string
}

var gcConfig Config
var eventExt Extension

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Equinix receiver needs a configuration file")
	}

	readConfig(os.Args[1])

	// Create a new client connection to the Equinix Metal API
	client, err := packngo.NewClient(packngo.WithAuth("packngo lib", gcConfig.Equinix.PacketAuthToken))
	if err != nil {
		log.Fatal(err)
	}

	// Get the details of the organization
	orgOpts := &packngo.GetOptions{}

	orgInfo, response, err := client.Organizations.Get(gcConfig.Equinix.OrganizationID, orgOpts)
	if err != nil {
		log.Fatal(err)
	}
	_ = response.Body.Close()

	log.Printf("Equinix Organization ID: < %s >", gcConfig.Equinix.OrganizationID)

	eventExt.OrganizationName = orgInfo.Name
	eventExt.OrganizationId = orgInfo.ID
	eventExt.Hostname = ""

	// Create a map which we'll use to track which events have been seen by the listener
	m := make(map[string]bool)

	// Record the time at which the first query is made to ensure we only send "new" events later
	timeNow := time.Now()

	for {

		// Get the events for the organization
		listOpts := &packngo.ListOptions{}

		projectList, response, err := client.Projects.List(listOpts)
		if err != nil {
			log.Printf("Failed to get projects list: %v", err)
			continue
		}
		_ = response.Body.Close()

		for _, project := range projectList {
			// log.Printf("Project: %s / %s ", project.ID, project.Name)

			eventExt.ProjectName = project.Name
			eventExt.ProjectId = project.ID

			projEvents, response, err := client.Projects.ListEvents(project.ID, listOpts)
			if err != nil {
				log.Printf("Failed to get events list: %v", err)
				continue
			}
			_ = response.Body.Close()

			// Send the first batch of events (Orgnization level events)
			m, err = createCloudEvent(projEvents, m, timeNow, eventExt)
			if err != nil {
				log.Printf("Failed to construct cloudevent: %v", err)
				continue
			}
			_ = response.Body.Close()

			// Get the list of devices for the specific project
			deviceList, response, err := client.Devices.List(project.ID, listOpts)
			if err != nil {
				log.Printf("Failed to get devices list: %v", err)
				continue
			}
			_ = response.Body.Close()

			for _, device := range deviceList {

				// Get the list of event for the specific devices
				deviceEvents, response, err := client.Devices.ListEvents(device.ID, listOpts)
				if err != nil {
					log.Printf("Failed to get device events list: %v", err)
					continue
				}
				_ = response.Body.Close()

				eventExt.Hostname = device.Hostname

				// Send the second batch of events (device level events)
				m, err = createCloudEvent(deviceEvents, m, timeNow, eventExt)
				if err != nil {
					log.Printf("Failed to construct cloudevent: %v", err)
					continue
				}
			}
		}

		time.Sleep(10 * time.Second)

	}

}

func createCloudEvent(eventList []packngo.Event, mapEvents map[string]bool, lastTime time.Time, extension Extension) (map[string]bool, error) {

	for _, event := range eventList {

		if _, seen := mapEvents[event.ID]; !seen {

			log.Printf("Event: %s / %s / %s", event.ID, event.Type, event.Interpolated)

			ce := cloudevents.NewEvent()
			id := uuid.New()
			ce.SetID(id.String())
			ce.SetSource("direktiv/listener/equinix/" + extension.OrganizationName + "/" + extension.ProjectName)
			ce.SetType(event.Type)
			err := ce.SetData(event)
			if err != nil {
				log.Printf("Event data error: %s / %s / %s", event.ID, event.Type, err.Error())
				continue
			}
			ce.SetExtension("orgname", extension.OrganizationName)
			ce.SetExtension("orgid", extension.OrganizationId)
			ce.SetExtension("projname", extension.ProjectName)
			ce.SetExtension("projid", extension.ProjectId)
			ce.SetExtension("hostname", extension.Hostname)

			data, err := ce.MarshalJSON()
			if err != nil {
				log.Printf("JSON marshal error: %v", err)
				continue
			}
			fmt.Printf("%s,", data)

			eventTime := event.CreatedAt

			if eventTime.After(lastTime) {
				err = sendCloudEvent(ce)
				if err != nil {
					log.Printf("failed to send cloudevent: %v", err)
					continue
				}
			}
			mapEvents[event.ID] = true
		}
	}
	return mapEvents, nil
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
