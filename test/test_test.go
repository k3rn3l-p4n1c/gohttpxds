package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/k3rn3l-p4n1c/gohttpxds"
	"github.com/k3rn3l-p4n1c/gohttpxds/pkg/mockserver"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test(t *testing.T) {
	nodeId := "testNode"

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	config := mockserver.Config{
		Listeners: []mockserver.Listener{{
			Name:    "listener_0",
			Address: "0.0.0.0",
			Port:    18000,
			RouteConfig: mockserver.RouteConfig{
				Name: "route_config_0",
				VirtualHosts: []mockserver.VirtualHost{{
					Name:    "virtual_host_0",
					Domains: []string{"test"},
					Routes: []mockserver.Route{{
						Name:   "route_0",
						Prefix: "/",
						Cluster: mockserver.Cluster{
							Name: "cluster_0",
							Endpoints: []mockserver.Endpoint{{
								UpstreamHost: "jsonplaceholder.typicode.com",
								UpstreamPort: 80,
							}},
						},
					}},
				}},
			},
		}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockServer := mockserver.New(context.Background(), nodeId, 18000)
	mockServer.StartRunning(context.Background())

	gohttpxds.Register("127.0.0.1:18000", grpc.WithTransportCredentials(insecure.NewCredentials()), nodeId)

	if resp, err := http.Get("xds://test/todos/1"); err != nil {
		panic(err.Error())
	} else {
		assert.Equal(t, 404, resp.StatusCode)
	}

	mockServer.SetConfig(ctx, config)
	time.Sleep(1 * time.Second)

	if resp, err := http.Get("xds://test2/todos/1"); err != nil {
		panic(err.Error())
	} else {
		assert.Equal(t, 404, resp.StatusCode)
	}

	if resp, err := http.Get("xds://test/todos/1"); err != nil {
		panic(err.Error())
	} else {
		assert.Equal(t, 200, resp.StatusCode)
	}

}
