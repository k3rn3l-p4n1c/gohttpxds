package transport

import (
	"fmt"
	"net/http"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func (w *Wrapper) getFirstMatchedRoute(req *http.Request) *route.Route {
	routes, err := w.cache.GetRouteConfig("")
	if err != nil {
		panic(fmt.Errorf("fail to get routes: %w", err))
	}
	for _, rc := range routes {
		for _, vh := range rc.VirtualHosts {
			if !doesMatchVirtualHost(req, vh) {
				continue
			}

			for _, route := range vh.Routes {
				if doesMatchRoutes(req, route.GetMatch()) {
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
