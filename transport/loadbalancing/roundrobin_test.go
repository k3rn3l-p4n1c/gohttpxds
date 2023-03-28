package loadbalancing_test

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/k3rn3l-p4n1c/gohttpxds/transport/loadbalancing"
	"github.com/stretchr/testify/assert"
)

func TestRoundRobinLoadBalancer(t *testing.T) {
	expectedChoice := []string{"Host 1", "Host 2", "Host 3", "Host 4", "Host 1"}
	cluster := &clusterv3.Cluster{
		Name:     "test",
		LbPolicy: clusterv3.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Hostname: "Host 1",
						},
					},
				}, {
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Hostname: "Host 2",
						},
					},
				}, {
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Hostname: "Host 3",
						},
					},
				}, {
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Hostname: "Host 4",
						},
					},
				}},
			}},
		},
	}

	chosenHosts := []string{}
	for i := 0; i <= 4; i++ {
		chosenHosts = append(chosenHosts, loadbalancing.ChooseEndpoint(cluster).Hostname)
	}

	assert.Equal(t, expectedChoice, chosenHosts)
}
