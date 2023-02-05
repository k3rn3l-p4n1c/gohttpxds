package transport

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func (w *Wrapper) getCluster(ra *route.RouteAction) ([]*cluster.Cluster, error) {
	switch clusterSpecifier := ra.ClusterSpecifier.(type) {
	case *route.RouteAction_Cluster:
		name := clusterSpecifier.Cluster
		return w.cache.GetCluster(name)
	case *route.RouteAction_ClusterHeader:
		panic("not implemented")
	case *route.RouteAction_WeightedClusters:
		panic("not implemented")
	case *route.RouteAction_ClusterSpecifierPlugin:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}
