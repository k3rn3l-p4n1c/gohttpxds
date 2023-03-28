package loadbalancing

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

type loadBalancer interface {
	Choose([]*endpointv3.LbEndpoint) *endpointv3.LbEndpoint
}

var loadBalancers map[string]loadBalancer

var lbPolicyConstructors map[clusterv3.Cluster_LbPolicy]func() loadBalancer = map[clusterv3.Cluster_LbPolicy]func() loadBalancer{
	clusterv3.Cluster_ROUND_ROBIN:                  func() loadBalancer { return &roundRobinLoadBalancer{} },
	clusterv3.Cluster_LEAST_REQUEST:                func() loadBalancer { return &nilLoadBalancer{} },
	clusterv3.Cluster_RING_HASH:                    func() loadBalancer { return &nilLoadBalancer{} },
	clusterv3.Cluster_RANDOM:                       func() loadBalancer { return &roundRobinLoadBalancer{} },
	clusterv3.Cluster_MAGLEV:                       func() loadBalancer { return &nilLoadBalancer{} },
	clusterv3.Cluster_CLUSTER_PROVIDED:             func() loadBalancer { return &nilLoadBalancer{} },
	clusterv3.Cluster_LOAD_BALANCING_POLICY_CONFIG: func() loadBalancer { return &nilLoadBalancer{} },
}

func init() {
	loadBalancers = make(map[string]loadBalancer)
}

func getOrCreateLoadBalancer(cluster *clusterv3.Cluster) loadBalancer {
	// todo check equality of clusters
	lb, found := loadBalancers[cluster.Name]
	if found {
		return lb
	}

	// todo implement other load balancers
	lb = lbPolicyConstructors[cluster.LbPolicy]()
	loadBalancers[cluster.Name] = lb
	return lb
}

func ChooseEndpoint(cluster *clusterv3.Cluster) *endpointv3.Endpoint {
	locality := chooseLocality(cluster.LoadAssignment.Endpoints)
	lb := getOrCreateLoadBalancer(cluster)
	return lb.Choose(locality.LbEndpoints).HostIdentifier.(*endpointv3.LbEndpoint_Endpoint).Endpoint
}

func chooseLocality(localityLbEndpoints []*endpointv3.LocalityLbEndpoints) *endpointv3.LocalityLbEndpoints {
	// todo: not implemented
	return localityLbEndpoints[0]
}
