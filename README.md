# ginny-consul
consul provider for ginny.


Configuration is possible via the standard
[Consul Environment
Variables](https://developer.hashicorp.com/consul/commands#environment-variables)
and via the target URI passed to
[`grpc.Dial`](https://pkg.go.dev/google.golang.org/grpc#Dial).

To register the resolver with the grpc-go run:

```go
resolver.Register(consul.NewBuilder())
```

Afterwards it can be used by calling grpc.Dial() and passing an URI in the
following format:

```
consul://[<consul-server>]/<serviceName>[?<OPT>[&<OPT>]...]
```

`<OPT>` is one of:

| OPT    | Format                         | Default                                                                                            | Description                                                                                                                                                      |
| ------ | ------------------------------ | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| scheme | `http\|https`                  | default from [github.com/hashicorp/consul/api](https://pkg.go.dev/github.com/hashicorp/consul/api) | Establish connection to consul via http or https.                                                                                                                |
| tags   | `<tag>,[,<tag>]...`            |                                                                                                    | Filter service by tags                                                                                                                                           |
| health | `healthy\|fallbackToUnhealthy` | healthy                                                                                            | `healthy` resolves only to services with a passing health status.<br>`fallbackToUnhealthy` resolves to unhealthy ones if none exist with passing healthy status. |
| token  | `string`                       | default from [github.com/hashicorp/consul/api](https://pkg.go.dev/github.com/hashicorp/consul/api) | Authenticate Consul API Request with the token.                                                                                                                  |