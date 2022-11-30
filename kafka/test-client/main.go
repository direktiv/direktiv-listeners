package main

import (
	"context"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func main() {
	topic := "hello"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	ce := `{
    "specversion": "1.0",
    "id": "123",
    "type": "com.au.test",
    "source": "kafka",
		"data": "testdata"
  }
  `
	m := kafka.Message{
		Headers: make([]kafka.Header, 1),
		Value:   []byte(ce),
	}
	m.Headers[0].Key = "Content-type"
	m.Headers[0].Value = []byte("application/cloudevents")

	m2 := kafka.Message{
		Headers: make([]kafka.Header, 5),
		Value:   []byte("mydata"),
	}
	m2.Headers[0].Key = "Content-type"
	m2.Headers[0].Value = []byte("plain/text")
	m2.Headers[1].Key = "myid"
	m2.Headers[1].Value = []byte("456")
	m2.Headers[2].Key = "mysource"
	m2.Headers[2].Value = []byte("kafkasource")
	m2.Headers[3].Key = "val1"
	m2.Headers[3].Value = []byte("helloworld")

	i, err := conn.WriteMessages(
		m2,
		m,
	)

	fmt.Printf("written %d bytes\n", i)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}

}
