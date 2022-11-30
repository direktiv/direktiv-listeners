package main

import (
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/emersion/go-message/mail"
)

type Attachment struct {
	Data        []byte `json:"data"`
	ContentType string `json:"type"`
	Name        string `json:"name"`
}

func errorHandler(msg string, err error) {
	log.Printf("ERRRROROO %v", err)
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
		log.Printf("using token to login")
		options = append(options,
			cehttp.WithHeader("Direktiv-Token", gcConfig.Direktiv.Token))
	} else if len(gcConfig.Direktiv.ApiKey) > 0 {
		log.Printf("using apikey to login")
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
		log.Printf("unable to create cloudevent client: " + err.Error())
		return err
	}

	_, _, err = c.Send(context.Background(), event)
	if err != nil {
		log.Printf("unable to send cloudevent client: " + err.Error())
		return err
	}

	return nil

}

func handleAttachments(mr *mail.Reader) ([]*Attachment, string, error) {

	attachments := make([]*Attachment, 0)
	var message string

	for {
		p, err := mr.NextPart()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, "", err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := ioutil.ReadAll(p.Body)
			if len(string(b)) > 0 {
				message = string(b)
			}
		case *mail.AttachmentHeader:

			ct, _, err := h.ContentType()
			if err != nil {
				return nil, "", err
			}

			filename, err := h.Filename()
			if err != nil {
				return nil, "", err
			}

			// get body
			b, err := ioutil.ReadAll(p.Body)
			if err != nil {
				return nil, "", err
			}

			a := &Attachment{
				Data:        b,
				ContentType: ct,
				Name:        filename,
			}

			attachments = append(attachments, a)
		}
	}

	return attachments, message, nil

}
