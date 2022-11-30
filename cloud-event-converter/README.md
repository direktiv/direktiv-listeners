# Cloud Event Converter

This app is designed to receive and modify cloud events according to rules provided in a server config file. Modifications will only be made if a received cloud event request matches a rule's "conditions". Conditions and "modifiers" are both accepted as `jq` queries.

## Example

```toml
[server]
  bind = ":8080"

[[rules]]
  direktivEndpoint = "https://example.direktiv.io/api/namespaces/example/event"
  condition = '.type == "greetingcloudevent"'
  modifiers = ['.data.name = "Trent"', 'del(.subject)', 'del(.time)', 'del(.comexampleextension1)']
```

This config file instructs the app to host a server on port 8080. Cloud events that hit the server and have the `type` field of their JSON body set to `greetingcloudevent` will be caught by the defined rule, and have all listed modifiers performed on it before it is forwarded to the specified direktiv endpoint. 

### Incoming Request

```json
{
    "specversion" : "1.0",
    "type" : "greetingcloudevent",
    "source" : "Direktiv",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
	  "data": {
			"name": "adjksalskd"
		}
}
```

### Modified payload

The following payload will be sent to the URL defined by the `direktivEndpoint` field in the rule that caught this event.

```json
{
    "specversion" : "1.0",
    "type" : "greetingcloudevent",
    "source" : "Direktiv",
    "id" : "A234-1234-1234",
	  "data": {
			"name": "Trent"
		}
}
```

