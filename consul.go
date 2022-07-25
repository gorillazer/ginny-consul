package consul

import (
	"fmt"

	"github.com/google/wire"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

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
func (p *Client) ServiceRegister(id, name, addr string, port int,
	tags []string, meta map[string]string) error {
	check := &consulApi.AgentServiceCheck{
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "60m",
		TCP:                            fmt.Sprintf("%s:%d", addr, port),
	}
	svcReg := &consulApi.AgentServiceRegistration{
		ID:                id,
		Name:              name,
		Tags:              []string{"grpc"},
		Port:              port,
		Address:           addr,
		EnableTagOverride: true,
		Check:             check,
		Checks:            nil,
	}

	err := p.Client.Agent().ServiceRegister(svcReg)
	if err != nil {
		return err
	}
	return nil
}

// ServiceDeregister
func (p *Client) ServiceDeregister(id string) error {
	return p.Client.Agent().ServiceDeregister(id)
}

var ProviderSet = wire.NewSet(New, NewOptions)
