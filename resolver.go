package consul

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/duke-git/lancet/random"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

type consulResolver struct {
	cc           resolver.ClientConn
	client       *consul.Client
	ctx          context.Context
	cancel       context.CancelFunc
	healthFilter healthFilter
	service      string
	tags         []string
}

// newConsulResolver
func newConsulResolver(
	service string,
	tags []string,
	cc resolver.ClientConn,
	client *consul.Client,
	healthFilter healthFilter,
) *consulResolver {
	ctx, cancel := context.WithCancel(context.Background())
	return &consulResolver{
		cc:           cc,
		client:       client,
		ctx:          ctx,
		cancel:       cancel,
		service:      service,
		tags:         tags,
		healthFilter: healthFilter,
	}
}

func (c *consulResolver) ResolveNow(o resolver.ResolveNowOptions) {
	var lastReportedAddrs []resolver.Address
	go c.watcher(lastReportedAddrs)
}

func (c *consulResolver) Resolver(ctx context.Context) (string, error) {
	var (
		err               error
		lastReportedAddrs = []resolver.Address{}
	)
	// rand
	if len(lastReportedAddrs) == 0 {
		opts := (&consul.QueryOptions{}).WithContext(c.ctx)
		lastReportedAddrs, _, err = c.query(opts)
		if err != nil {
			return "", fmt.Errorf("error retrieving instances from consul: %s, %v, %w", c.service, c.tags, err)
		}
		go c.watcher(lastReportedAddrs)
	}
	i := random.RandInt(0, len(lastReportedAddrs))
	if lastReportedAddrs[i].Addr != "" {
		return lastReportedAddrs[i].Addr, nil
	}

	return "", fmt.Errorf("error retrieving instances from consul: %s, %v", c.service, c.tags)
}

func (c *consulResolver) Close() {
	c.cancel()
}

func (c *consulResolver) start() {
	var lastReportedAddrs []resolver.Address
	go c.watcher(lastReportedAddrs)
}

func (c *consulResolver) watcher(lastReportedAddrs []resolver.Address) {
	opts := (&consul.QueryOptions{}).WithContext(c.ctx)
	for range time.NewTicker(50 * time.Millisecond).C {
		var (
			err           error
			addrs         []resolver.Address
			lastWaitIndex = opts.WaitIndex
		)
		addrs, opts.WaitIndex, err = c.query(opts)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if c.cc != nil {
				c.cc.ReportError(err)
			}
			// fmt.Println(time.Now(), "load nodes error", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if opts.WaitIndex < lastWaitIndex {
			opts.WaitIndex = 0
			continue
		}
		if addressesEqual(addrs, lastReportedAddrs) {
			continue
		}

		if c.cc != nil {
			_ = c.cc.UpdateState(resolver.State{Addresses: addrs})
		}
		lastReportedAddrs = addrs
	}

}

func (c *consulResolver) query(opts *consul.QueryOptions) ([]resolver.Address, uint64, error) {
	services, meta, err := c.client.Health().ServiceMultipleTags(c.service, c.tags,
		c.healthFilter == healthFilterOnlyHealthy, opts)
	if err != nil {
		return nil, 0, err
	}

	if c.healthFilter == healthFilterFallbackToUnhealthy {
		services = filterPreferOnlyHealthy(services)
	}

	result := make([]resolver.Address, 0, len(services))
	for _, v := range services {
		addr := v.Service.Address
		if addr == "" {
			addr = v.Node.Address
		}
		result = append(result, resolver.Address{
			Addr: net.JoinHostPort(addr, fmt.Sprint(v.Service.Port)),
		})
	}

	return result, meta.LastIndex, nil
}

// filterPreferOnlyHealthy if entries contains services with passing health
// check only entries with passing health are returned.
// Otherwise entries is returned unchanged.
func filterPreferOnlyHealthy(entries []*consul.ServiceEntry) []*consul.ServiceEntry {
	healthy := make([]*consul.ServiceEntry, 0, len(entries))

	for _, e := range entries {
		if e.Checks.AggregatedStatus() == consul.HealthPassing {
			healthy = append(healthy, e)
		}
	}

	if len(healthy) != 0 {
		return healthy
	}

	return entries
}

func addressesEqual(a, b []resolver.Address) bool {
	if a == nil && b != nil {
		return false
	}

	if a != nil && b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Addr != b[i].Addr {
			return false
		}
	}

	return true
}
