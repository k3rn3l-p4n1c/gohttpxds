package transport

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

const (
	WEIGHTED_LOAD_BALANCING_ALGORITHM = "weighted"
)

func chooseEndpoint(cluster *clusterv3.Cluster, lbAlgorithm string) *endpointv3.Endpoint {
	if lbAlgorithm == WEIGHTED_LOAD_BALANCING_ALGORITHM {
		return weightedSelect(cluster)
	}
	return cluster.LoadAssignment.Endpoints[0].LbEndpoints[0].HostIdentifier.(*endpointv3.LbEndpoint_Endpoint).Endpoint
}

func weightedSelect(cluster *clusterv3.Cluster) *endpointv3.Endpoint {
	// get the list of endpoints from the current cluster
	var totalWeight int32
	endpoints := cluster.GetLoadAssignment().GetEndpoints()

	// get all lbEndpoints
	var lbEndpoints []*endpointv3.LbEndpoint
	for _, e := range endpoints {
		lbEndpoints = append(lbEndpoints, e.LbEndpoints...)
	}

	// Calculate the total weight for all lb endpoints
	for _, lbe := range lbEndpoints {
		totalWeight += int32(lbe.GetLoadBalancingWeight().Value)
	}

	// Calculate the probability for each endpoint
	var probabilities []float64
	for _, lbe := range lbEndpoints {
		probabilities = append(probabilities, float64(int32(lbe.GetLoadBalancingWeight().Value)/totalWeight))
	}
	// find max probility index
	maxProbailityIndex := 0
	for i, p := range probabilities {
		if float64(probabilities[maxProbailityIndex]) > p {
			maxProbailityIndex = i
		}
	}

	return lbEndpoints[maxProbailityIndex].HostIdentifier.(*endpointv3.LbEndpoint_Endpoint).Endpoint
}
