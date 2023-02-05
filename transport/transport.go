package transport

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/k3rn3l-p4n1c/gohttpxds/internal/xdscache"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func New(transport http.RoundTripper, cache xdscache.XDSCache) http.RoundTripper {
	return &Wrapper{
		transport: transport,
		cache:     cache,
	}
}

type Wrapper struct {
	transport http.RoundTripper
	cache     xdscache.XDSCache
}

func (w *Wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "xds" {
		return w.transport.RoundTrip(req)
	}

	route := w.getFirstMatchedRoute(req)
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
	req = w.doAction(req, route)

	logRequest(req)
	return w.transport.RoundTrip(req)
}

func (w *Wrapper) doAction(req *http.Request, r *route.Route) *http.Request {
	switch action := r.Action.(type) {
	case *route.Route_Route:
		return w.doRouteAction(req, action.Route)
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
