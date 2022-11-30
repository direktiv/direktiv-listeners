package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"


	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/rs/xid"
	kafka "github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"

	"github.com/vorteil/direktiv-listeners/utils"
)

const (
	kafkaMapping = "KAFKA_MAPPING"
	kafkaContext = "KAFKA_CONTEXT"

	minRead = 16384
	maxRead = 1048576
)

// KafkaListener ...
type KafkaListener struct {
	conn, topic, gid string
	lifeLine         chan bool
}

// NewKafkaListener returns a new kafka listener
func NewKafkaListener(conn, topic, gid string) *KafkaListener {

	kl := &KafkaListener{
		conn:     conn,
		topic:    topic,
		gid:      gid,
		lifeLine: make(chan bool),
	}

	return kl

}

// Stop stops kafka listener
func (kl *KafkaListener) Stop() {
	kl.lifeLine <- true
}

func headerValueForName(n string, headers []kafka.Header) string {

	for _, h := range headers {
		if string(h.Key) == n {
			return string(h.Value)
		}
	}

	return ""

}

func handleBinaryEvent(m *kafka.Message) error {

	// try to handle ce- headers
	event := cloudevents.NewEvent()
	ct := "application/x-binary"

	for _, h := range m.Headers {

		if strings.ToLower(h.Key) == "content-type" {
			ct = string(h.Value)
		}

		switch h.Key {
		case "ce_type":
			{
				event.Context.SetType(string(h.Value))
			}
		case "ce_source":
			{
				event.Context.SetSource(string(h.Value))
			}
		case "ce_id":
			{
				event.Context.SetID(string(h.Value))
			}
		case "ce_time":
			{
				// has to be RFC 3339
				t, err := time.Parse(time.RFC3339, string(h.Value))
				if err != nil {
					return err
				}
				event.Context.SetTime(t)
			}
		}

	}

	mappings, err := url.ParseQuery(os.Getenv(kafkaMapping))
	if err != nil {
		return err
	}

	setDefault := func(k, d, v string, fn func(string) error) error {
		if v == "" {
			mid, ok := mappings[k]
			if ok {
				mv := headerValueForName(mid[0], m.Headers)
				if mv != "" {
					d = mv
				}
			}
			return fn(d)
		}
		return nil
	}

	setDefault("id", xid.New().String(), event.Context.GetID(), event.Context.SetID)
	setDefault("type", "com.au.direktiv.kafka.message", event.Context.GetType(), event.Context.SetType)
	setDefault("source", "direktiv-kafka", event.Context.GetSource(), event.Context.SetSource)

	event.Context.SetDataContentType(ct)
	event.DataEncoded = m.Value

	for _, k := range strings.Split(os.Getenv(kafkaContext), ",") {
		hv := headerValueForName(k, m.Headers)
		if hv != "" {
			event.Context.SetExtension(k, hv)
		}
	}

	log.Debugf("%v", event.String())

	_, err = utils.SendCloudEvent(&event)
	return err

}

func handleCloudEvent(key, value string) error {

	log.Debugf("handle cloud event: %v, value: %v", key, value)

	// check if data is cloud event
	event := cloudevents.NewEvent()
	err := json.Unmarshal([]byte(value), &event)
	if err != nil {
		return err
	}

	if event.Context.GetID() == "" {
		return fmt.Errorf("cloudevent id is missing")
	}

	if event.Context.GetSource() == "" {
		return fmt.Errorf("cloudevent source is missing")
	}

	if event.Context.GetType() == "" {
		return fmt.Errorf("cloudevent type is missing")
	}

	_, err = utils.SendCloudEvent(&event)

	return err

}

// Listen starts listening to the configured kafka instance
func (kl *KafkaListener) Listen() error {

	var (
		err error
		m   kafka.Message
	)

	brokers := strings.Split(kl.conn, ",")
	topics := strings.Split(kl.topic, ",")

	log.Infof("listening to %s on %s", kl.topic, kl.conn)

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     kl.gid,
		GroupTopics: topics,
		MinBytes:    minRead,
		MaxBytes:    maxRead,
		StartOffset: kafka.LastOffset,
	})
	defer kr.Close()

	for {
		m, err = kr.ReadMessage(context.Background())
		if err != nil {
			log.Errorf("can not read message: %v", io.EOF)
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}

		isCE := false
		for _, h := range m.Headers {
			if strings.ToLower(h.Key) == "content-type" &&
				strings.HasPrefix(string(h.Value), "application/cloudevents") {
				isCE = true
			}
		}

		log.Debugf("message at topic:%v, partition:%v offset:%v	%s", m.Topic, m.Partition, m.Offset, string(m.Key))

		if isCE {
			err := handleCloudEvent(string(m.Key), string(m.Value))
			if err != nil {
				log.Errorf("message with key '%s' could not be send: %v", string(m.Key), err)
			}
			continue
		}

		// handle binary
		err := handleBinaryEvent(&m)
		if err != nil {
			log.Errorf("binary message with key '%s' could not be send: %v", string(m.Key), err)
		}

	}

	return err
}

// Lifeline returns the chan for stop/wait
func (kl *KafkaListener) Lifeline() chan bool {
	return kl.lifeLine
}
