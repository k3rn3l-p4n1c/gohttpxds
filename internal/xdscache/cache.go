package xdscache

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	// endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type XDSCache interface {
	GetListener(string) ([]*listenerv3.Listener, error)
	GetRouteConfig(string) ([]*routev3.RouteConfiguration, error)
	GetCluster(string) ([]*clusterv3.Cluster, error)

	WatchListener(string)
	WatchRouteConfig(string)
	WatchCluster(string)
}
