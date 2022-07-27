package consul

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/duke-git/lancet/cryptor"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ProviderSet
var ProviderSet = wire.NewSet(New, NewOptions)

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

// New
func New(o *consulApi.Config) (*Client, error) {

	// initialize consul
	var (
		consulCli *consulApi.Client
		err       error
	)
	if o.Address == "" {
		return nil, errors.New("consul server address is undefined")
	}

	consulCli, err = consulApi.NewClient(o)
	if err != nil {
		return nil, errors.Wrap(err, "create consul client error")
	}

	c := &Client{
		Config: o,
		Client: consulCli,
	}

	return c, nil
}

// ServiceRegister
func (p *Client) ServiceRegister(service, addr string, tags []string, meta map[string]string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	check := &consulApi.AgentServiceCheck{
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "60m",
		TCP:                            u.Host,
	}
	id := cryptor.Md5String(service)
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return err
	}
	svcReg := &consulApi.AgentServiceRegistration{
		ID:                id,
		Name:              service,
		Tags:              tags,
		Port:              int(port),
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
func (p *Client) ServiceDeregister(service string) error {
	return p.Client.Agent().ServiceDeregister(service)
}

// Resolver
func (p *Client) Resolver(ctx context.Context, service, tag string) (addr string, err error) {
	var lastIndex uint64
	services, metainfo, err := p.Client.Health().Service(service, tag, true, &api.QueryOptions{
		WaitIndex: lastIndex,
	})
	if err != nil {
		return "", err
	}
	lastIndex = metainfo.LastIndex

	for _, s := range services {
		return fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port), nil
	}
	return "", fmt.Errorf("error retrieving instances from consul")
}
