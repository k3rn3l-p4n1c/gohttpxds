package xdscache

import (
	"fmt"

	"github.com/k3rn3l-p4n1c/gohttpxds/internal/xdsclient"
	"google.golang.org/protobuf/proto"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/rs/zerolog/log"
)

func New(xdsClient xdsclient.XDSClient) XDSCache {
	return &xdsCache{
		xdsClient:              xdsClient,
		listeners:              make(map[string][]*listenerv3.Listener),
		routeConfigs:           make(map[string][]*routev3.RouteConfiguration),
		clusters:               make(map[string][]*clusterv3.Cluster),
		virtualHosts:           make(map[string][]*routev3.VirtualHost),
		clusterLoadAssignments: make(map[string][]*endpointv3.ClusterLoadAssignment),
	}
}

type xdsCache struct {
	xdsClient xdsclient.XDSClient

	listeners              map[string][]*listenerv3.Listener
	routeConfigs           map[string][]*routev3.RouteConfiguration
	clusters               map[string][]*clusterv3.Cluster
	virtualHosts           map[string][]*routev3.VirtualHost
	clusterLoadAssignments map[string][]*endpointv3.ClusterLoadAssignment
}

func (x *xdsCache) GetListener(name string) ([]*listenerv3.Listener, error) {
	resource, exists := x.listeners[name]
	if !exists {
		return nil, fmt.Errorf("resource not found")
	}

	return resource, nil
}
func (x *xdsCache) GetRouteConfig(name string) ([]*routev3.RouteConfiguration, error) {
	if name == "" {
		resources := []*routev3.RouteConfiguration{}
		for k := range x.routeConfigs {
			resources = append(resources, x.routeConfigs[k]...)
		}
		return resources, nil
	}
	resource, exists := x.routeConfigs[name]
	if !exists {
		return nil, fmt.Errorf("resource not found")
	}

	return resource, nil

}
func (x *xdsCache) GetCluster(name string) ([]*clusterv3.Cluster, error) {
	resource, exists := x.clusters[name]
	if !exists {
		return nil, fmt.Errorf("resource not found")
	}

	return resource, nil

}

func (x *xdsCache) WatchListener(name string) {
	x.xdsClient.WatchListener(name, x.listenerCallback)
}
func (x *xdsCache) WatchRouteConfig(name string) {
	x.xdsClient.WatchRouteConfig(name, x.routeConfigCallback)
}
func (x *xdsCache) WatchCluster(name string) {
	x.xdsClient.WatchCluster(name, x.clusterCallback)
}

func (x *xdsCache) listenerCallback(resources []*listenerv3.Listener, err error) {
	log.Debug().Int("count", len(resources)).Msg("new listeners received")
	for _, resource := range resources {
		x.listeners[resource.Name] = append(x.listeners[resource.Name], resource)
	}

	for l := range resources {
		for i := range resources[l].FilterChains {
			for j := range resources[l].FilterChains[i].Filters {
				manager := &hcmv3.HttpConnectionManager{}
				if err := proto.Unmarshal(resources[l].FilterChains[i].Filters[j].GetTypedConfig().GetValue(), manager); err != nil {
					panic(fmt.Errorf("failed to unmarshal resource: %w", err).Error())
				}

				if hcmrds, ok := manager.GetRouteSpecifier().(*hcmv3.HttpConnectionManager_Rds); ok {
					x.WatchRouteConfig(hcmrds.Rds.RouteConfigName)
				}
			}
		}
	}
}
func (x *xdsCache) routeConfigCallback(resources []*routev3.RouteConfiguration, err error) {
	log.Debug().Int("count", len(resources)).Msg("new routes received")

	for _, resource := range resources {
		x.routeConfigs[resource.Name] = append(x.routeConfigs[resource.Name], resource)
	}

}
func (x *xdsCache) clusterCallback(resources []*clusterv3.Cluster, err error) {
	log.Debug().Int("count", len(resources)).Msg("new clusters received")

	for _, resource := range resources {
		x.clusters[resource.Name] = append(x.clusters[resource.Name], resource)
	}

}
