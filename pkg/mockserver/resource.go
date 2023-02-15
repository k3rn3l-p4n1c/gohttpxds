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

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Config struct {
	Listeners []Listener
}

type Listener struct {
	Name        string
	Address     string
	Port        uint32
	RouteConfig RouteConfig
}

type RouteConfig struct {
	Name         string
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

func makeCluster(c Cluster) *clusterv3.Cluster {
	return &clusterv3.Cluster{
		Name:                 c.Name,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_LOGICAL_DNS},
		LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: c.Name,
			Endpoints:   makeEndpoint(c.Endpoints),
		},
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
	}
}

func makeEndpoint(endpoints []Endpoint) []*endpointv3.LocalityLbEndpoints {
	var result []*endpointv3.LocalityLbEndpoints
	for _, e := range endpoints {
		result = append(result, &endpointv3.LocalityLbEndpoints{
			LbEndpoints: []*endpointv3.LbEndpoint{{
				HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
					Endpoint: &endpointv3.Endpoint{
						Address: &corev3.Address{
							Address: &corev3.Address_SocketAddress{
								SocketAddress: &corev3.SocketAddress{
									Protocol: corev3.SocketAddress_TCP,
									Address:  e.UpstreamHost,
									PortSpecifier: &corev3.SocketAddress_PortValue{
										PortValue: e.UpstreamPort,
									},
								},
							},
						},
					},
				},
			}},
		})
	}
	return result
}

func makeRoute(routeConfig RouteConfig) *routev3.RouteConfiguration {
	return &routev3.RouteConfiguration{
		Name:         routeConfig.Name,
		VirtualHosts: makeVirtualHost(routeConfig.VirtualHosts),
	}
}

func makeVirtualHost(virtualHosts []VirtualHost) []*routev3.VirtualHost {
	var result []*routev3.VirtualHost
	for _, vh := range virtualHosts {
		result = append(result, &routev3.VirtualHost{
			Name:    vh.Name,
			Domains: vh.Domains,
			Routes:  makeRoutes(vh.Routes),
		})
	}

	return result
}

func makeRoutes(routes []Route) []*routev3.Route {
	var result []*routev3.Route
	for _, r := range routes {
		result = append(result, &routev3.Route{
			Name: r.Name,
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: r.Prefix,
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: r.Cluster.Name,
					},
					// HostRewriteSpecifier: &routev3.RouteAction_HostRewriteLiteral{
					// 	HostRewriteLiteral: r.Cluster.Endpoints[0].UpstreamHost,
					// },
				},
			},
		})
	}
	return result
}

func makeHTTPListener(listenerName string, address string, port uint32, route string) *listenerv3.Listener {
	routerConfig, _ := anypb.New(&routerv3.Router{})
	// HTTP filter configuration
	manager := &hcmv3.HttpConnectionManager{
		CodecType:  hcmv3.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
			Rds: &hcmv3.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: route,
			},
		},
		HttpFilters: []*hcmv3.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}
	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	return &listenerv3.Listener{
		Name: listenerName,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{{
			Filters: []*listenerv3.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

func makeConfigSource() *corev3.ConfigSource {
	source := &corev3.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &corev3.ConfigSource_ApiConfigSource{
		ApiConfigSource: &corev3.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   corev3.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*corev3.GrpcService{{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}

func GenerateSnapshot(config Config) *cache.Snapshot {
	var listeners []types.Resource
	var clusters []types.Resource
	var routes []types.Resource

	for _, l := range config.Listeners {
		listeners = append(listeners, makeHTTPListener(l.Name, l.Address, l.Port, l.RouteConfig.Name))
		routes = append(routes, makeRoute(l.RouteConfig))

		for _, vh := range l.RouteConfig.VirtualHosts {
			for _, r := range vh.Routes {
				clusters = append(clusters, makeCluster(r.Cluster))
			}
		}
	}

	snap, _ := cache.NewSnapshot("1",
		map[resource.Type][]types.Resource{
			resource.ClusterType:  clusters,
			resource.RouteType:    routes,
			resource.ListenerType: listeners,
		},
	)
	return snap
}
