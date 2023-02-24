package consul

import (
	"context"
	"net/url"
	"strconv"

	"github.com/duke-git/lancet/cryptor"
	"github.com/google/wire"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc/resolver"
)

// ProviderSet
var (
	ProviderSet = wire.NewSet(NewClient, NewOptions)
)

func init() {
	resolver.Register(NewBuilder())
}

// NewOptions
func NewOptions(v *viper.Viper) (*consulApi.Config, error) {
	var (
		err error
		o   = new(consulApi.Config)
	)
	if err = v.UnmarshalKey("consul", o); err != nil {
		return nil, errors.Wrapf(err, "viper unmarshal consul options error")
	}

	return o, nil
}

// Client
type Client struct {
	Config *consulApi.Config
	Client *consulApi.Client
}

// NewClient
func NewClient(ctx context.Context, conf *consulApi.Config) (*Client, error) {
	// initialize consul
	var (
		consulCli *consulApi.Client
		err       error
	)

	consulCli, err = consulApi.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "create consul client error")
	}

	c := &Client{
		Config: conf,
		Client: consulCli,
	}

	return c, nil
}

// ServiceRegister
func (p *Client) ServiceRegister(ctx context.Context, service, addr string, tags []string, meta map[string]string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	check := &consulApi.AgentServiceCheck{
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "24h",
		TCP:                            u.Host,
	}
	id := cryptor.Md5String(addr)
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return err
	}

	svcReg := &consulApi.AgentServiceRegistration{
		ID:                id,
		Name:              service,
		Tags:              tags,
		Port:              port,
		Address:           u.Hostname(),
		EnableTagOverride: true,
		Check:             check,
		Checks:            nil,
	}

	err = p.Client.Agent().ServiceRegister(svcReg)
	if err != nil {
		return err
	}
	return nil
}

// ServiceDeregister
func (p *Client) ServiceDeregister(ctx context.Context, service string) error {
	return p.Client.Agent().ServiceDeregister(service)
}

// Resolver
func (p *Client) Resolver(ctx context.Context, service string, tags []string) (addr string, err error) {
	r := newConsulResolver(service, tags, nil, p.Client, healthFilterOnlyHealthy)
	return r.Resolver(ctx)
}
