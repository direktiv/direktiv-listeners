package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv-listeners/kafka/internal"
)

const (
	kafkaConn    = "KAFKA_CONNECTION"
	kafkaTopic   = "KAFKA_TOPIC"
	kafkaGroupID = "KAFKA_GID"
	kafkaDebug   = "DEBUG"
)

func main() {

	log.Infof("starting kafka direktiv listener")

	conn := os.Getenv(kafkaConn)
	topic := os.Getenv(kafkaTopic)
	gid := os.Getenv(kafkaGroupID)

	// get kafka connection
	if conn == "" || topic == "" {
		log.Errorf("kafak topic and connection required")
		os.Exit(1)
	}

	if os.Getenv(kafkaDebug) == "true" {
		log.SetLevel(log.DebugLevel)
	}

	if gid == "" {
		gid = "direktiv-gid"
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	kl := internal.NewKafkaListener(conn, topic, gid)

	go func() {
		err := kl.Listen()
		if err != nil {
			log.Errorf("failed listenting to kafka: %v", err)
		}
		done <- true
	}()

	log.Infof("started listening")
	<-done

	log.Infof("stopping kafka listener")
	go kl.Stop()

	<-kl.Lifeline()
	log.Infof("kafka listener stopped")

}
