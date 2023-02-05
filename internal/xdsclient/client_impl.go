package xdsclient

import (
	"context"
	"fmt"
	"log"
	"sync"

	xdsCore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	xds "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	lds "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	rds "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	xdsresource "github.com/k3rn3l-p4n1c/gohttpxds/internal/xdsclient/resource"
	"github.com/k3rn3l-p4n1c/gohttpxds/pkg/event"
)

func New(config ServerConfig) (XDSClient, error) {
	conn, err := grpc.Dial(config.ServerURI, config.Creds)
	if err != nil {
		return nil, fmt.Errorf("fail to dial xds server: %w", err)
	}

	return &clientImpl{
		conn:            conn,
		serverConfig:    config,
		done:            &event.Event{},
		rdsClient:       rds.NewRouteDiscoveryServiceClient(conn),
		ldsClient:       lds.NewListenerDiscoveryServiceClient(conn),
		cdsClient:       cds.NewClusterDiscoveryServiceClient(conn),
		listernersNames: make(map[string]struct{}),
		clustersNames:   make(map[string]struct{}),
		routesNames:     make(map[string]struct{}),
	}, nil
}

type clientImpl struct {
	serverConfig  ServerConfig
	conn          *grpc.ClientConn
	done          *event.Event
	resourceTypes resourceTypeRegistry
	rdsClient     rds.RouteDiscoveryServiceClient
	ldsClient     lds.ListenerDiscoveryServiceClient
	cdsClient     cds.ClusterDiscoveryServiceClient

	listernersNames map[string]struct{}
	clustersNames   map[string]struct{}
	routesNames     map[string]struct{}
}

func (c *clientImpl) addListener(resourceName string) {
	if resourceName == "" {
		return
	}
	_, exists := c.listernersNames[resourceName]
	if !exists {
		c.listernersNames[resourceName] = struct{}{}
	}
}

func (c *clientImpl) GetListeners() []string {
	listeners := make([]string, 0, len(c.listernersNames))
	for k := range c.listernersNames {
		listeners = append(listeners, k)
	}
	return listeners
}

func (c *clientImpl) addCluster(resourceName string) {
	if resourceName == "" {
		return
	}
	_, exists := c.clustersNames[resourceName]
	if !exists {
		c.clustersNames[resourceName] = struct{}{}
	}
}

func (c *clientImpl) GetClusters() []string {
	clusters := make([]string, 0, len(c.clustersNames))
	for k := range c.clustersNames {
		clusters = append(clusters, k)
	}
	return clusters
}

func (c *clientImpl) addRoute(resourceName string) {
	if resourceName == "" {
		return
	}
	_, exists := c.routesNames[resourceName]
	if !exists {
		c.routesNames[resourceName] = struct{}{}
	}
}

func (c *clientImpl) GetRoutes() []string {
	routes := make([]string, 0, len(c.routesNames))
	for k := range c.routesNames {
		routes = append(routes, k)
	}
	return routes
}

func (c *clientImpl) WatchListener(resourceName string, callback func([]*listener.Listener, error)) func() {
	c.addListener(resourceName)
	streamClient, err := c.ldsClient.StreamListeners(context.TODO())
	if err != nil {
		panic(fmt.Errorf("failed to stream: %w", err).Error())
	}
	genericCallback := func(resources []*any.Any, err error) {
		if err != nil {
			callback(nil, err)
			return
		}
		listeners := make([]*listener.Listener, len(resources))
		for i := range resources {
			l := &listener.Listener{}
			if err := proto.Unmarshal(resources[i].GetValue(), l); err != nil {
				panic(fmt.Errorf("failed to unmarshal resource: %w", err).Error())
			}
			listeners[i] = l
		}
		callback(listeners, nil)
	}

	return c.watchResources(c.GetListeners, streamClient, genericCallback)
}

func (c *clientImpl) WatchRouteConfig(resourceName string, callback func([]*route.RouteConfiguration, error)) func() {
	streamClient, err := c.rdsClient.StreamRoutes(context.TODO())
	if err != nil {
		panic(fmt.Errorf("failed to stream: %w", err).Error())
	}
	genericCallback := func(resources []*any.Any, err error) {
		if err != nil {
			callback(nil, err)
			return
		}
		routeConfigs := make([]*route.RouteConfiguration, len(resources))
		for i := range resources {
			rc := &route.RouteConfiguration{}
			if err := proto.Unmarshal(resources[i].GetValue(), rc); err != nil {
				panic(fmt.Errorf("failed to unmarshal resource: %w", err).Error())
			}
			routeConfigs[i] = rc
		}
		callback(routeConfigs, nil)
	}

	return c.watchResources(c.GetRoutes, streamClient, genericCallback)
}

func (c *clientImpl) WatchCluster(resourceName string, callback func([]*cluster.Cluster, error)) func() {
	streamClient, err := c.cdsClient.StreamClusters(context.TODO())
	if err != nil {
		panic(fmt.Errorf("failed to stream: %w", err).Error())
	}
	genericCallback := func(resources []*any.Any, err error) {
		if err != nil {
			callback(nil, err)
			return
		}
		clusters := make([]*cluster.Cluster, len(resources))
		for i := range resources {
			c := &cluster.Cluster{}
			if err := proto.Unmarshal(resources[i].GetValue(), c); err != nil {
				panic(fmt.Errorf("failed to unmarshal resource: %w", err).Error())
			}
			clusters[i] = c
		}
		callback(clusters, nil)
	}

	return c.watchResources(c.GetClusters, streamClient, genericCallback)

}

type streamClient interface {
	Send(*xds.DiscoveryRequest) error
	Recv() (*xds.DiscoveryResponse, error)
	grpc.ClientStream
}

func (c *clientImpl) watchResources(getResourceNames func() []string, sc streamClient, callback func([]*any.Any, error)) func() {
	var cancel chan struct{}

	go func() {
		for {
			req := &xds.DiscoveryRequest{
				Node: &xdsCore.Node{
					Id: c.serverConfig.NodeId,
				},
				ResourceNames: getResourceNames(),
				VersionInfo:   "2",
			}
			sc.Send(req)
			resp, err := sc.Recv()
			if err != nil {
				callback(nil, err)
				continue
			}

			callback(resp.GetResources(), nil)
		}
	}()
	return func() {
		cancel <- struct{}{}
	}
}

// Close closes the gRPC connection to the management server.
//
// TODO: ensure that all underlying transports are closed before this function
// returns.
func (c *clientImpl) Close() {
	if c.done.HasFired() {
		return
	}
	c.done.Fire()

	log.Printf("Shutdown")
}

// A registry of xdsresource.Type implementations indexed by their corresponding
// type URLs. Registration of an xdsresource.Type happens the first time a watch
// for a resource of that type is invoked.
type resourceTypeRegistry struct {
	mu    sync.Mutex
	types map[string]xdsresource.Type
}

func newResourceTypeRegistry() *resourceTypeRegistry {
	return &resourceTypeRegistry{types: make(map[string]xdsresource.Type)}
}

func (r *resourceTypeRegistry) get(url string) xdsresource.Type {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.types[url]
}

func (r *resourceTypeRegistry) maybeRegister(rType xdsresource.Type) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	url := rType.TypeURL()
	typ, ok := r.types[url]
	if ok && typ != rType {
		return fmt.Errorf("attempt to re-register a resource type implementation for %v", rType.TypeEnum())
	}
	r.types[url] = rType
	return nil
}
