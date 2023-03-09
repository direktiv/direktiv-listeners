package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Bind         string `yaml:"bind"`
		TLS          bool   `yaml:"tls"`
		CertFilePath string `yaml:"certFile"`
		KeyFilePath  string `yaml:"keyFile"`
	} `yaml:"server"`
	GitLab struct {
		Token string `yaml:"token"`
	} `yaml:"gitlab"`
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
		log.Fatal("Gitlab receiver needs a configuration file")
	}

	readConfig(os.Args[1])

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		gitlabToken := req.Header.Get("X-Gitlab-Token")
		if gitlabToken != gcConfig.GitLab.Token {
			code := http.StatusForbidden
			http.Error(w, http.StatusText(code), code)
			log.Printf("Received a request with a missing or invalid gitlab token.")
			return
		}

		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			log.Printf("Failed to read request payload: %v.", err)
			return
		}

		m := make(map[string]interface{})
		err = json.Unmarshal(payload, &m)
		if err != nil {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			log.Printf("Failed to unmarshal request payload: %v.", err)
			return
		}

		// log.Printf("Event: %s / %s / %s", event.Action, event.Created, strconv.Itoa(event.ID))

		ce := cloudevents.NewEvent()
		id, err := uuid.Parse(req.Header.Get("X-Gitlab-Event-UUID"))
		if err != nil {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			log.Printf("Invalid event ID: %v.", err)
			return
		}

		ce.SetID(id.String())
		ce.SetSource("direktiv/listener/gitlab")
		ce.SetType(req.Header.Get("X-Gitlab-Event"))
		ce.SetTime(time.Now())
		ce.SetDataContentType("application/json")
		err = ce.SetData(m)
		if err != nil {
			log.Printf("Failed to set cloudevent data: %v", err)
			return
		}

		err = sendCloudEvent(ce)
		if err != nil {
			log.Printf("Failed to send cloudevent: %v", err)
			return
		}

	})

	if gcConfig.Server.TLS {
		log.Fatal(http.ListenAndServeTLS(gcConfig.Server.Bind, gcConfig.Server.CertFilePath, gcConfig.Server.KeyFilePath, nil))
	} else {
		log.Fatal(http.ListenAndServe(gcConfig.Server.Bind, nil))
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
