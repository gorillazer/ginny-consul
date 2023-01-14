package consul

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/duke-git/lancet/cryptor"
	"github.com/duke-git/lancet/random"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/syncmap"
)

// ProviderSet
var (
	ProviderSet = wire.NewSet(NewClient, NewOptions)
	NodeMap     = syncmap.Map{}
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
		DeregisterCriticalServiceAfter: "60m",
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
func (p *Client) Resolver(ctx context.Context, service, tag string) (addr string, err error) {
	key := fmt.Sprintf("%s.%s", service, tag)

	var (
		flag  = false
		nodes = []*consulApi.AgentService{}
	)
	data, ok := NodeMap.Load(key)
	if !ok {
		flag = true
	}
	if n, ok := data.([]*consulApi.AgentService); ok {
		nodes = n
	} else {
		flag = true
	}
	if flag || len(nodes) == 0 {
		nodes, err = p.loadNodes(ctx, key, service, tag)
		if err != nil {
			return "", err
		}
		p.hotReloadNodes(ctx, key, service, tag)
	}

	// rand
	if len(nodes) > 0 {
		i := random.RandInt(0, len(nodes))
		for k, v := range nodes {
			if k == i && v.Address != "" {
				return fmt.Sprintf("%s:%d", v.Address, v.Port), nil
			}
		}
	}

	return "", fmt.Errorf("error retrieving instances from consul: %s, %s", service, tag)
}

// loadNodes ...
func (p *Client) loadNodes(ctx context.Context, key, service, tag string) ([]*consulApi.AgentService, error) {
	services, _, err := p.Client.Health().Service(service, tag, true, &api.QueryOptions{
		WaitIndex: 0,
	})
	if err != nil {
		return nil, err
	}

	nodes := []*consulApi.AgentService{}
	for _, v := range services {
		if v.Service != nil {
			nodes = append(nodes, v.Service)
		}
	}
	NodeMap.Store(key, nodes)
	return nodes, nil
}

// hotReloadNodes ...
func (p *Client) hotReloadNodes(ctx context.Context, key, service, tag string) {
	pCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func(c context.Context, k, s, t string) {
		for range time.NewTicker(time.Second * 10).C {
			_, err := p.loadNodes(c, k, s, t)
			if err != nil {
				// fmt.Println(time.Now(), "load nodes error", err)
				continue
			}
		}
	}(pCtx, key, service, tag)
}
