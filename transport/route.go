package transport

import (
	"fmt"
	"net/http"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func (w *Wrapper) getFirstMatchedRoute(req *http.Request) *routev3.Route {
	routes, err := w.cache.GetRouteConfig("")
	if err != nil {
		panic(fmt.Errorf("fail to get routes: %w", err))
	}
	for _, rc := range routes {
		for _, vh := range rc.VirtualHosts {
			if !doesMatchVirtualHost(req, vh) {
				continue
			}

			for _, routev3 := range vh.Routes {
				if doesMatchRoutes(req, routev3.GetMatch()) {
					return routev3
				}
			}
		}
	}

	return nil
}

func doesMatchVirtualHost(req *http.Request, virtualHost *routev3.VirtualHost) bool {
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
