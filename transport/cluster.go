package transport

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func (w *Wrapper) getCluster(ra *routev3.RouteAction) ([]*clusterv3.Cluster, error) {
	switch clusterSpecifier := ra.ClusterSpecifier.(type) {
	case *routev3.RouteAction_Cluster:
		name := clusterSpecifier.Cluster
		return w.cache.GetCluster(name)
	case *routev3.RouteAction_ClusterHeader:
		panic("not implemented")
	case *routev3.RouteAction_WeightedClusters:
		panic("not implemented")
	case *routev3.RouteAction_ClusterSpecifierPlugin:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}
