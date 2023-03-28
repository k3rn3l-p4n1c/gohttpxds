package loadbalancing

import (
	"sync"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

type nilLoadBalancer struct {
	currentIndex int
	mtx          sync.Mutex
}

func (lb *nilLoadBalancer) Choose(endpoints []*endpointv3.LbEndpoint) *endpointv3.LbEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	return endpoints[0]
}
