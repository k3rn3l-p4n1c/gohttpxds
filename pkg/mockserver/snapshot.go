package mockserver

import (
	"context"
	"strconv"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/rs/zerolog/log"
)

var (
	version = 0
	nodeID  = "testNode"
)

func (m *MockServer) updateSnapshot(config Config) error {
	resourceMap := Cast(config)
	version++
	snapshot, err := cache.NewSnapshot(
		strconv.Itoa(version),
		resourceMap,
	)
	if err != nil {
		log.Error().Err(err).Msg("fail to create snapshot")
		return err
	}

	if err := snapshot.Consistent(); err != nil {
		log.Error().Err(err).Msg("snapshot inconsistency")
		return err
	}
	log.Debug().Msg("will serve snapshot")

	// Add the snapshot to the cache
	if err := m.cache.SetSnapshot(context.TODO(), nodeID, snapshot); err != nil {
		log.Error().Err(err).Msg("snapshot error")
		return err
	}
	return nil
}
