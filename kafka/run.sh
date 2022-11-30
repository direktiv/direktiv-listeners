#!/bin/bash

curl -X POST http://localhost/api/namespaces/test



DEBUG=true KAFKA_CONNECTION=localhost:9092 KAFKA_MAPPING="id=myid&source=mysource" DIREKTIV_URL=http://localhost/api/namespaces/test/event KAFKA_TOPIC=hello go run cmd/main.go
