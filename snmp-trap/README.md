# SNMP Trap

Captures SNMP 'TRAP' messages, and forwards them on to the specified Direktiv backend.

### Example 

```toml
[event]
  source = "dell/unity"
  type = "dell.unity.alert"

[server]
  addr = ":8080"

[direktiv]
  url = "https://example.direktiv.io"
  namespace = "example"
  authToken = "eyEXAMPLETOKENFOOBAR..."
```

