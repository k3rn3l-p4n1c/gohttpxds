package transport

import (
	"fmt"
	"net/http"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/k3rn3l-p4n1c/gohttpxds/transport/loadbalancing"
)

func (w *Wrapper) doRouteAction(req *http.Request, ra *routev3.RouteAction) *http.Request {
	cluster, err := w.getCluster(ra)
	if err != nil {
		panic(fmt.Errorf("fail to find cluster: %w", err))
	}

	add := loadbalancing.ChooseEndpoint(cluster[0]).Address.Address.(*corev3.Address_SocketAddress).SocketAddress
	host := add.Address
	port := add.PortSpecifier.(*corev3.SocketAddress_PortValue).PortValue
	req.URL.Host = fmt.Sprintf("%s:%d", host, port)
	req.URL.Scheme = "http"
	req.Host = fmt.Sprintf("%s:%d", host, port)

	return req
}
