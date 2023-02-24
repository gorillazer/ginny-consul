package consul

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(NewBuilder())
}

func TestName(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name string
		arg  args
		fn   func(param interface{})
		// mockFunc func() (patches *Patches)
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		// if tt.mockFunc != nil {
		// 	patch := tt.mockFunc()
		// 	defer patch.Reset()
		// }
		t.Run(tt.name, func(t *testing.T) {

			client, err := grpc.Dial("consul://192.168.0.11:8500/pb.Say.grpc?scheme=http&tags=&health=fallbackToUnhealthy",
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			assert.NoError(t, err)
			assert.NotEmpty(t, client)
			err = client.Invoke(context.Background(), "/pb.Say/Hello", nil, nil)
			assert.NoError(t, err)

			cli, err := NewClient(context.Background(), &api.Config{
				Address:  "192.168.0.11:8500",
				Scheme:   "http",
				Token:    "",
				WaitTime: 10 * time.Minute,
			})
			assert.NoError(t, err)
			addr, err := cli.Resolver(context.Background(), "pb.Say.grpc", nil)
			assert.NoError(t, err)
			assert.Equal(t, addr, "192.168.0.16:9000")
		})
	}
}
