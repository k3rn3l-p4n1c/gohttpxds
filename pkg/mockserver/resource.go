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
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
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

func makeCluster(c Cluster) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 c.Name,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: c.Name,
			Endpoints:   makeEndpoint(c.Endpoints),
		},
		DnsLookupFamily: cluster.Cluster_V4_ONLY,
	}
}

func makeEndpoint(endpoints []Endpoint) []*endpoint.LocalityLbEndpoints {
	var result []*endpoint.LocalityLbEndpoints
	for _, e := range endpoints {
		result = append(result, &endpoint.LocalityLbEndpoints{
			LbEndpoints: []*endpoint.LbEndpoint{{
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
			}},
		})
	}
	return result
}

func makeRoute(routeConfig RouteConfig) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name:         routeConfig.Name,
		VirtualHosts: makeVirtualHost(routeConfig.VirtualHosts),
	}
}

func makeVirtualHost(virtualHosts []VirtualHost) []*route.VirtualHost {
	var result []*route.VirtualHost
	for _, vh := range virtualHosts {
		result = append(result, &route.VirtualHost{
			Name:    vh.Name,
			Domains: vh.Domains,
			Routes:  makeRoutes(vh.Routes),
		})
	}

	return result
}

func makeRoutes(routes []Route) []*route.Route {
	var result []*route.Route
	for _, r := range routes {
		result = append(result, &route.Route{
			Name: r.Name,
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
					// HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
					// 	HostRewriteLiteral: r.Cluster.Endpoints[0].UpstreamHost,
					// },
				},
			},
		})
	}
	return result
}

func makeHTTPListener(listenerName string, address string, port uint32, route string) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: route,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}
	pbst, err := anypb.New(manager)
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
