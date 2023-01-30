package xdscache

import (
	"fmt"

	"github.com/k3rn3l-p4n1c/gohttpxds/internal/xdsclient"
	"google.golang.org/protobuf/proto"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/rs/zerolog/log"
)

func New(xdsClient xdsclient.XDSClient) XDSCache {
	return &xdsCache{
		xdsClient:              xdsClient,
		listeners:              make(map[string][]*listener.Listener),
		routeConfigs:           make(map[string][]*route.RouteConfiguration),
		clusters:               make(map[string][]*cluster.Cluster),
		virtualHosts:           make(map[string][]*route.VirtualHost),
		clusterLoadAssignments: make(map[string][]*endpoint.ClusterLoadAssignment),
	}
}

type xdsCache struct {
	xdsClient xdsclient.XDSClient

	listeners              map[string][]*listener.Listener
	routeConfigs           map[string][]*route.RouteConfiguration
	clusters               map[string][]*cluster.Cluster
	virtualHosts           map[string][]*route.VirtualHost
	clusterLoadAssignments map[string][]*endpoint.ClusterLoadAssignment
}

func (x *xdsCache) GetListener(name string) ([]*listener.Listener, error) {
	resource, exists := x.listeners[name]
	if !exists {
		return nil, fmt.Errorf("resource not found")
	}

	return resource, nil
}
func (x *xdsCache) GetRouteConfig(name string) ([]*route.RouteConfiguration, error) {
	if name == "" {
		resources := []*route.RouteConfiguration{}
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
func (x *xdsCache) GetCluster(name string) ([]*cluster.Cluster, error) {
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

func (x *xdsCache) listenerCallback(resources []*listener.Listener, err error) {
	log.Debug().Int("count", len(resources)).Msg("new listeners received")
	for _, resource := range resources {
		x.listeners[resource.Name] = append(x.listeners[resource.Name], resource)
	}

	for l := range resources {
		for i := range resources[l].FilterChains {
			for j := range resources[l].FilterChains[i].Filters {
				manager := &hcm.HttpConnectionManager{}
				if err := proto.Unmarshal(resources[l].FilterChains[i].Filters[j].GetTypedConfig().GetValue(), manager); err != nil {
					panic(fmt.Errorf("failed to unmarshal resource: %w", err).Error())
				}

				if hcmrds, ok := manager.GetRouteSpecifier().(*hcm.HttpConnectionManager_Rds); ok {
					x.WatchRouteConfig(hcmrds.Rds.RouteConfigName)
				}
			}
		}
	}
}
func (x *xdsCache) routeConfigCallback(resources []*route.RouteConfiguration, err error) {
	log.Debug().Int("count", len(resources)).Msg("new routes received")

	for _, resource := range resources {
		x.routeConfigs[resource.Name] = append(x.routeConfigs[resource.Name], resource)
	}

}
func (x *xdsCache) clusterCallback(resources []*cluster.Cluster, err error) {
	log.Debug().Int("count", len(resources)).Msg("new clusters received")

	for _, resource := range resources {
		x.clusters[resource.Name] = append(x.clusters[resource.Name], resource)
	}

}
