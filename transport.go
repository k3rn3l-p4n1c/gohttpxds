package gohttpxds

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/k3rn3l-p4n1c/gohttpxds/internal/xdscache"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/rs/zerolog/log"
)

type TransportWrapper struct {
	Transport http.RoundTripper
	cache     xdscache.XDSCache
}

func (t *TransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "xds" {
		return t.Transport.RoundTrip(req)
	}

	route := t.getFirstMatchedRoute(req)
	if route == nil {
		body := "No route found"
		return &http.Response{
			Status:        "404 Not Found",
			StatusCode:    404,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
			ContentLength: int64(len(body)),
			Request:       req,
			Header:        make(http.Header, 0),
		}, nil
	}
	req = t.doAction(req, route)

	logRequest(req)
	return t.Transport.RoundTrip(req)
}

func (t *TransportWrapper) doAction(req *http.Request, r *route.Route) *http.Request {
	switch action := r.Action.(type) {
	case *route.Route_Route:
		return t.doRouteAction(req, action.Route)
	case *route.Route_Redirect:
		panic("not implemented")
	case *route.Route_DirectResponse:
		panic("not implemented")
	case *route.Route_FilterAction:
		panic("not implemented")
	case *route.Route_NonForwardingAction:
		panic("not implemented")
	default:
		panic("unknown route action type")
	}
}

func (t *TransportWrapper) doRouteAction(req *http.Request, ra *route.RouteAction) *http.Request {
	cluster, err := t.getCluster(ra)
	if err != nil {
		panic(fmt.Errorf("fail to find cluster: %w", err))
	}
	add := cluster[0].LoadAssignment.Endpoints[0].LbEndpoints[0].HostIdentifier.(*endpoint.LbEndpoint_Endpoint).Endpoint.Address.Address.(*core.Address_SocketAddress).SocketAddress
	host := add.Address
	// port := add.PortSpecifier.(*core.SocketAddress_PortValue).PortValue
	req.URL.Host = host // fmt.Sprintf("%s:%d", host, port)
	req.URL.Scheme = "http"
	req.Host = host

	return req
}

func (t *TransportWrapper) getCluster(ra *route.RouteAction) ([]*cluster.Cluster, error) {
	switch clusterSpecifier := ra.ClusterSpecifier.(type) {
	case *route.RouteAction_Cluster:
		name := clusterSpecifier.Cluster
		return t.cache.GetCluster(name)
	case *route.RouteAction_ClusterHeader:
		panic("not implemented")
	case *route.RouteAction_WeightedClusters:
		panic("not implemented")
	case *route.RouteAction_ClusterSpecifierPlugin:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

func (t *TransportWrapper) getFirstMatchedRoute(req *http.Request) *route.Route {
	routes, err := t.cache.GetRouteConfig("")
	if err != nil {
		panic(fmt.Errorf("fail to get routes: %w", err))
	}
	for _, rc := range routes {
		for _, vh := range rc.VirtualHosts {
			if !doesMatchVirtualHost(req, vh) {
				continue
			}

			for _, route := range vh.Routes {
				if doesMatchRoutes(req, route) {
					return route
				}
			}
		}
	}

	return nil
}

func doesMatchVirtualHost(req *http.Request, virtualHost *route.VirtualHost) bool {
	for _, domain := range virtualHost.Domains {
		if domain == "*" {
			return true
		}
		if domain == req.URL.Host {
			return true
		}
	}
	return false
}

func doesMatchRoutes(req *http.Request, route *route.Route) bool {
	return true // todo
}

func logRequest(req *http.Request) {
	log.Debug().Str("url", req.URL.String()).Msg("sending requests")
}
