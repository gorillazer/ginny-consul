package consul

import (
	"context"
	"time"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

type resolverBuilder struct{}

func NewBuilder() *resolverBuilder {
	return &resolverBuilder{}
}

func (*resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	serviceName, scheme, tags, health, token, err := parseEndpoint(&target.URL)
	if err != nil {
		return nil, err
	}

	cli, err := NewClient(context.Background(), &api.Config{
		Address:  target.URL.Host,
		Scheme:   scheme,
		Token:    token,
		WaitTime: 10 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	r := newConsulResolver(serviceName, tags, cc, cli.Client, health)
	r.start()

	return r, nil
}

// Scheme returns the URI scheme for the resolver
func (*resolverBuilder) Scheme() string {
	return scheme
}
