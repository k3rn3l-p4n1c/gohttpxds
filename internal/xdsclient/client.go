package xdsclient

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/grpc"
)

// ServerConfig contains the configuration to connect to a server, including
// URI, creds, and transport API version (e.g. v2 or v3).
type ServerConfig struct {
	// ServerURI is the management server to connect to.
	//
	// The bootstrap file contains an ordered list of xDS servers to contact for
	// this authority. The first one is picked.
	ServerURI string
	// Creds contains the credentials to be used while talking to the xDS
	// server, as a grpc.DialOption.
	Creds grpc.DialOption
}

type XDSClient interface {
	WatchListener(string, func([]*listener.Listener, error)) func()
	WatchRouteConfig(string, func([]*route.RouteConfiguration, error)) func()
	WatchCluster(string, func([]*cluster.Cluster, error)) func()

	Close()
}
