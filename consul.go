package consul

import (
	"github.com/google/wire"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

type Client struct {
	Config *consulApi.Config
	Client *consulApi.Client
}

// New
func New(o *consulApi.Config, logger *zap.Logger) (*Client, error) {

	// initialize consul
	var (
		consulCli *consulApi.Client
		err       error
	)
	if o.Address == "" {
		logger.Warn("The consul server address is not configured, and the provider will not take effect.")
		return nil, nil
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

var ProviderSet = wire.NewSet(New, NewOptions)
