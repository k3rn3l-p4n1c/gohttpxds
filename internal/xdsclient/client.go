package xdsclient

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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

	NodeId string
}

type XDSClient interface {
	WatchListener(string, func([]*listenerv3.Listener, error)) func()
	WatchRouteConfig(string, func([]*routev3.RouteConfiguration, error)) func()
	WatchCluster(string, func([]*clusterv3.Cluster, error)) func()

	Close()
}
