// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package mockserver

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
)

type Config struct {
	Listeners []Listener
}

type Listener struct {
	Name         string
	Address      string
	Port         uint32
	VirtualHosts []VirtualHost
}

type VirtualHost struct {
	Name    string
	Domains []string
	Routes  []Route
}

type Route struct {
	Name    string
	Prefix  string
	Cluster Cluster
}

type Cluster struct {
	Name      string
	Endpoints []Endpoint
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}

func Cast(c Config) map[resource.Type][]types.Resource {
	var endpoints []types.Resource
	var listeners []types.Resource
	var clusters []types.Resource
	var routes []types.Resource
	var virtualHosts []types.Resource

	var routesArray []Route

	for _, l := range c.Listeners {
		listeners = append(listeners, MakeHTTPListener(l.Name, l.Address, l.Port))
		var thisVirtualHosts []*route.VirtualHost
		for _, vh := range l.VirtualHosts {
			virtualHosts = append(virtualHosts, MakeVirtualHost(vh))
			thisVirtualHosts = append(thisVirtualHosts, MakeVirtualHost(vh))
			for _, r := range vh.Routes {
				routesArray = append(routesArray, r)
				endpoints = append(endpoints, MakeEndpoint(r.Cluster.Name, r.Cluster.Endpoints))
				clusters = append(clusters, MakeCluster(r.Cluster))
			}
		}
		routes = append(routes, MakeRoute(thisVirtualHosts))
	}
	// routes = []types.Resource{MakeRoute(routesArray)}

	return map[resource.Type][]types.Resource{
		resource.EndpointType:    endpoints,
		resource.ClusterType:     clusters,
		resource.RouteType:       routes,
		resource.ListenerType:    listeners,
		resource.VirtualHostType: virtualHosts,
	}
}

// import (
// 	"time"

// 	"google.golang.org/protobuf/types/known/anypb"
// 	"google.golang.org/protobuf/types/known/durationpb"

// 	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
// 	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
// 	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
// 	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
// 	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
// 	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
// 	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
// 	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
// 	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
// 	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
// 	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
// )

// const (
// 	ClusterName  = "example_proxy_cluster"
// 	RouteName    = "local_route"
// 	ListenerName = "listener_0"
// 	ListenerPort = 10000
// 	UpstreamHost = "jsonplaceholder.typicode.com"
// 	UpstreamPort = 80
// )

// func makeCluster(clusterName string) *cluster.Cluster {
// 	return &cluster.Cluster{
// 		Name:                 clusterName,
// 		ConnectTimeout:       durationpb.New(5 * time.Second),
// 		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
// 		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
// 		LoadAssignment:       makeEndpoint(clusterName),
// 		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
// 	}
// }

// func makeEndpoint(clusterName string, Endpoint) *endpoint.ClusterLoadAssignment {
// 	return &endpoint.ClusterLoadAssignment{
// 		ClusterName: clusterName,
// 		Endpoints: []*endpoint.LocalityLbEndpoints{{
// 			LbEndpoints: []*endpoint.LbEndpoint{{
// 				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
// 					Endpoint: &endpoint.Endpoint{
// 						Address: &core.Address{
// 							Address: &core.Address_SocketAddress{
// 								SocketAddress: &core.SocketAddress{
// 									Protocol: core.SocketAddress_TCP,
// 									Address:  UpstreamHost,
// 									PortSpecifier: &core.SocketAddress_PortValue{
// 										PortValue: UpstreamPort,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			}},
// 		}},
// 	}
// }

// func makeRoute(routeName string, clusterName string) *route.RouteConfiguration {
// 	return &route.RouteConfiguration{
// 		Name: routeName,
// 		VirtualHosts: []*route.VirtualHost{{
// 			Name:    "local_service",
// 			Domains: []string{"*"},
// 			Routes: []*route.Route{{
// 				Match: &route.RouteMatch{
// 					PathSpecifier: &route.RouteMatch_Prefix{
// 						Prefix: "/",
// 					},
// 				},
// 				Action: &route.Route_Route{
// 					Route: &route.RouteAction{
// 						ClusterSpecifier: &route.RouteAction_Cluster{
// 							Cluster: clusterName,
// 						},
// 						HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
// 							HostRewriteLiteral: UpstreamHost,
// 						},
// 					},
// 				},
// 			}},
// 		}},
// 	}
// }

// func makeHTTPListener(listenerName string, route string) *listener.Listener {
// 	routerConfig, _ := anypb.New(&router.Router{})
// 	// HTTP filter configuration
// 	manager := &hcm.HttpConnectionManager{
// 		CodecType:  hcm.HttpConnectionManager_AUTO,
// 		StatPrefix: "http",
// 		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
// 			Rds: &hcm.Rds{
// 				ConfigSource:    makeConfigSource(),
// 				RouteConfigName: route,
// 			},
// 		},
// 		HttpFilters: []*hcm.HttpFilter{{
// 			Name:       wellknown.Router,
// 			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
// 		}},
// 	}
// 	pbst, err := anypb.New(manager)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return &listener.Listener{
// 		Name: listenerName,
// 		Address: &core.Address{
// 			Address: &core.Address_SocketAddress{
// 				SocketAddress: &core.SocketAddress{
// 					Protocol: core.SocketAddress_TCP,
// 					Address:  "0.0.0.0",
// 					PortSpecifier: &core.SocketAddress_PortValue{
// 						PortValue: ListenerPort,
// 					},
// 				},
// 			},
// 		},
// 		FilterChains: []*listener.FilterChain{{
// 			Filters: []*listener.Filter{{
// 				Name: wellknown.HTTPConnectionManager,
// 				ConfigType: &listener.Filter_TypedConfig{
// 					TypedConfig: pbst,
// 				},
// 			}},
// 		}},
// 	}
// }

// func makeConfigSource() *core.ConfigSource {
// 	source := &core.ConfigSource{}
// 	source.ResourceApiVersion = resource.DefaultAPIVersion
// 	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
// 		ApiConfigSource: &core.ApiConfigSource{
// 			TransportApiVersion:       resource.DefaultAPIVersion,
// 			ApiType:                   core.ApiConfigSource_GRPC,
// 			SetNodeOnFirstMessageOnly: true,
// 			GrpcServices: []*core.GrpcService{{
// 				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
// 					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
// 				},
// 			}},
// 		},
// 	}
// 	return source
// }

// func GenerateSnapshot() *cache.Snapshot {
// 	snap, _ := cache.NewSnapshot("1",
// 		map[resource.Type][]types.Resource{
// 			resource.ClusterType:  {makeCluster(ClusterName)},
// 			resource.RouteType:    {makeRoute(RouteName, ClusterName)},
// 			resource.ListenerType: {makeHTTPListener(ListenerName, RouteName)},
// 		},
// 	)
// 	return snap
// }

func MakeCluster(c Cluster) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 c.Name,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       MakeEndpoint(c.Name, c.Endpoints),
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		EdsClusterConfig:     makeEDSCluster(),
	}
}

func makeEDSCluster() *cluster.Cluster_EdsClusterConfig {
	return &cluster.Cluster_EdsClusterConfig{
		EdsConfig: makeConfigSource(),
	}
}

func MakeEndpoint(clusterName string, eps []Endpoint) *endpoint.ClusterLoadAssignment {
	var endpoints []*endpoint.LbEndpoint

	for _, e := range eps {
		endpoints = append(endpoints, &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  e.UpstreamHost,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: e.UpstreamPort,
								},
							},
						},
					},
				},
			},
		})
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: endpoints,
		}},
	}
}

func MakeRoute(virtualHosts []*route.VirtualHost) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name:         "listener_0",
		VirtualHosts: virtualHosts,
	}

}

func MakeVirtualHost(virtualHost VirtualHost) *route.VirtualHost {
	rts := makeRoutes(virtualHost.Routes)

	return &route.VirtualHost{
		Name:    virtualHost.Name,
		Domains: virtualHost.Domains,
		Routes:  rts,
	}
}

func makeRoutes(routes []Route) []*route.Route {
	var rts []*route.Route

	for _, r := range routes {
		rts = append(rts, &route.Route{
			//Name: r.Name,
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: r.Prefix,
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: r.Cluster.Name,
					},
				},
			},
		})
	}
	return rts
}

func MakeHTTPListener(listenerName, address string, port uint32) *listener.Listener {
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: "listener_0",
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
		}},
	}
	pbst, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
