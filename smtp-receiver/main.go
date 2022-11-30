package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"gopkg.in/yaml.v2"
)

var gcConfig Config
var r *rand.Rand

type Config struct {
	SMTP struct {
		Address string `yaml:"address"`
	} `yaml:"smtp"`
	Direktiv struct {
		DirektivEndpoint   string `yaml:"endpoint"`
		Token              string `yaml:"token"`
		ApiKey             string `yaml:"apikey"`
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	} `yaml:"direktiv"`
}

type backend struct{}

type session struct {
	to    []string
	event cloudevents.Event
	data  map[string]interface{}
}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts smtp.MailOptions) error {
	s.data["from"] = from
	return nil
}

func (s *session) Rcpt(to string) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	mr, err := mail.CreateReader(r)
	if err != nil {
		errorHandler("can not create reader", err)
		return err
	}

	subj, err := mr.Header.Subject()
	if err != nil {
		errorHandler("can not read subject", err)
		return err
	}
	s.data["subject"] = subj

	attachments, message, err := handleAttachments(mr)
	if err != nil {
		errorHandler("can not read attachments/body", err)
		return err
	}

	s.data["attachments"] = attachments
	s.data["message"] = message
	s.data["to"] = s.to

	s.event.SetData(s.data)
	err = sendCloudEvent(s.event)

	if err != nil {
		errorHandler("can not send event", err)
		return err
	}

	return err
}

func (s *session) Reset() {
	s.data = make(map[string]interface{})
	s.event = basicEvent()
}

func (s *session) Logout() error {
	return nil
}

func (bkd *backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return &session{
		data:  make(map[string]interface{}),
		event: basicEvent(),
	}, nil
}

func (bkd *backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &session{
		data:  make(map[string]interface{}),
		event: basicEvent(),
	}, nil
}

func basicEvent() cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetID(fmt.Sprintf("smtp-cloud-%v", r.Int63()))
	event.SetSource("direktiv/listener/smtp")
	event.SetType("smtp.message")
	return event
}

func main() {

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	be := &backend{}
	s := smtp.NewServer(be)

	if len(os.Args) < 2 {
		log.Fatal("smtp listener needs config file")
	}

	readConfig(os.Args[1])

	s.Addr = gcConfig.SMTP.Address
	s.Domain = "localhost"
	s.AllowInsecureAuth = true
	// s.Debug = os.Stdout

	log.Println("Starting SMTP server at", s.Addr)
	log.Fatal(s.ListenAndServe())

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
