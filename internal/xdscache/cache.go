package xdscache

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	// endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type XDSCache interface {
	GetListener(string) ([]*listener.Listener, error)
	GetRouteConfig(string) ([]*route.RouteConfiguration, error)
	GetCluster(string) ([]*cluster.Cluster, error)

	WatchListener(string)
	WatchRouteConfig(string)
	WatchCluster(string)
}
