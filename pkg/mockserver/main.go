package mockserver

import (
	"context"
	"os"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
)

var (
	l Logger
)

type MockServer struct {
	cache  cache.SnapshotCache
	server server.Server
	nodeID string
	port   uint
}

func New(ctx context.Context, nodeID string, port uint) *MockServer {
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)
	callback := &test.Callbacks{Debug: l.Debug}
	srv := server.NewServer(ctx, cache, callback)

	return &MockServer{
		cache:  cache,
		server: srv,
		nodeID: nodeID,
		port:   port,
	}
}

func (m *MockServer) StartRunning(ctx context.Context) {
	go RunServer(ctx, m.server, m.port)
}

func (m *MockServer) SetConfig(ctx context.Context, config Config) {
	snapshot := GenerateSnapshot(config)

	// Create the snapshot that we'll serve to Envoy
	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}
	l.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := m.cache.SetSnapshot(ctx, m.nodeID, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}
