package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/k3rn3l-p4n1c/gohttpxds"
	"github.com/k3rn3l-p4n1c/gohttpxds/test/example"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	config := example.Config{
		Listeners: []example.Listener{{
			Name:    "listener_0",
			Address: "0.0.0.0",
			Port:    18000,
			RouteConfig: example.RouteConfig{
				Name: "route_config_0",
				VirtualHosts: []example.VirtualHost{{
					Name:    "virtual_host_0",
					Domains: []string{"test"},
					Routes: []example.Route{{
						Name:   "route_0",
						Prefix: "/",
						Cluster: example.Cluster{
							Name: "cluster_0",
							Endpoints: []example.Endpoint{{
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
	go example.Run(ctx, example.GenerateSnapshot(config))

	req, err := http.NewRequest("GET", "xds://test/todos/1", nil)
	if err != nil {
		panic(err.Error())
	}

	client, err := gohttpxds.NewHttpClient("127.0.0.1:18000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err.Error())
	}

	// client := &http.Client{}

	time.Sleep(1 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	assert.Equal(t, 200, resp.StatusCode)
	fmt.Printf("%v\n", resp)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%v\n", string(body))
}
