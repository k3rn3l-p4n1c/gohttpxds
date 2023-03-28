package loadbalancing

import (
	"sync"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

type roundRobinLoadBalancer struct {
	currentIndex int
	mtx          sync.Mutex
}

func (lb *roundRobinLoadBalancer) Choose(endpoints []*endpointv3.LbEndpoint) *endpointv3.LbEndpoint {
	lb.mtx.Lock()
	defer lb.mtx.Unlock()

	endpoint := endpoints[lb.currentIndex]

	lb.currentIndex = (lb.currentIndex + 1) % len(endpoints)

	return endpoint
}
