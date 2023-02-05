package transport

import (
	"fmt"
	"net/http"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func (w *Wrapper) doRouteAction(req *http.Request, ra *route.RouteAction) *http.Request {
	cluster, err := w.getCluster(ra)
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
