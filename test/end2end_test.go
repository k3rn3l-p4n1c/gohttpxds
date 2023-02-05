package test

import (
	"context"
	"fmt"
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
	const port = 18000
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := mockserver.Run(ctx, port)
	if err != nil {
		panic(err.Error())
	}

	server.UpdateSnapshot(mockserver.Config{
		Listeners: []mockserver.Listener{{
			Name:    "listener_0",
			Address: "0.0.0.0",
			Port:    80,
			VirtualHosts: []mockserver.VirtualHost{{
				Name:    "virtualhost_0",
				Domains: []string{"test"},
				Routes: []mockserver.Route{{
					Name:   "route_1",
					Prefix: "",
					Cluster: mockserver.Cluster{
						Name: "cluster_1",
						Endpoints: []mockserver.Endpoint{{
							UpstreamHost: "jsonplaceholder.typicode.com",
							UpstreamPort: 80,
						}},
					},
				}},
			}},
		}},
	})

	gohttpxds.Register(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))

	time.Sleep(5 * time.Second)

	if resp, err := http.Get("xds://test/todos/1"); err != nil {
		panic(err.Error())
	} else {
		assert.Equal(t, 200, resp.StatusCode)
	}

	if resp, err := http.Get("xds://test2/todos/1"); err != nil {
		panic(err.Error())
	} else {
		assert.Equal(t, 404, resp.StatusCode)
	}
}
