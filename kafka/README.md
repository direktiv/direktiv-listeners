# kafka listener

This listener connects to kafka brokers and forwards the messages as cloud events to
direktiv.

#### Build

```sh
docker build -t direktiv-kafka .
```

#### Environment variables

| Tables        | Are           |
| ------------- |:-------------:|
| KAFKA_CONNECTION          | Comma seperated list of brokers to connect to, e.g. 192.168.1.1:9092,myserver:8080     |
| KAFKA_TOPIC          | Comma seperated list of topics to subscribe to              |    
| KAFKA_GID          |   Groupd id of this instance            |      
| DEBUG          | Turn on debug logging ("true")              |       
| KAFKA_MAPPING          | Maps Kafka headers to cloud event fields              |       
| KAFKA_CONTEXT          | Maps additional Kafka headers values to the cloud event              |       
| DIREKTIV_URL          |  URL of the direktiv instance             |      
| DIREKTIV_TOKEN          | Authentication token for direktiv, e.g. 'Bearer abc12435' or 'Apikey myapikey'              |       

#### Additional information

Kafka can send as structured content. In that case this listner forwards the cloud event as-is to direktiv.

```text
------------------ Message -------------------

Topic Name: mytopic

------------------- key ----------------------

Key: mykey

------------------ headers -------------------

content-type: application/cloudevents+json; charset=UTF-8

------------------- value --------------------

{
    "specversion" : "1.0",
    "type" : "com.example.someevent",
    "source" : "/mycontext/subcontext",
    "id" : "1234-1234-1234",
    "time" : "2018-04-05T03:56:24Z",
    "datacontenttype" : "application/xml",
    "data" : {
        ... application data encoded in XML ...
    }
}

-----------------------------------------------
```

In binary format there are multiple steps involved to create the final cloudevent.
In the first step every ce_* header value will be used to populate the event.

After that every empty field will be populated with values defined in KAFKA_MAPPING, e.g. KAFKA_MAPPING="id=myid&source=mysource" would us the value in 'myid' as id and 'mysource' as the event source.

The last step adds the values listed in KAFKA_CONTEXT and adds it as context values to the event.


```text
------------------ Message -------------------

Topic Name: mytopic

------------------- key ----------------------

Key: mykey

------------------ headers -------------------

ce_specversion: "1.0"
ce_type: "com.example.someevent"
mysource: "/mycontext/subcontext"
myid: "1234-1234-1234"
ce_time: "2018-04-05T03:56:24Z"
content-type: application/avro
       .... further attributes ...

------------------- value --------------------

            ... application data encoded in Avro ...

-----------------------------------------------
```

#### Example workflow

```yaml
id: action
start:
  type: event
  event:
    type: com.au.test
states:
- id: notify
  type: noop
  log: .
```
