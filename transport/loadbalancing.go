package transport

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

func chooseEndpoint(cluster *clusterv3.Cluster) *endpointv3.Endpoint {
	return cluster.LoadAssignment.Endpoints[0].LbEndpoints[0].HostIdentifier.(*endpointv3.LbEndpoint_Endpoint).Endpoint
}
