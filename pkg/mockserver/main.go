// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package mockserver

import (
	"context"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/rs/zerolog/log"
)

type MockServer struct {
	cache cache.SnapshotCache
}

func Run(ctx context.Context, port uint) (*MockServer, error) {
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, nil)
	mockserver := &MockServer{cache: cache}

	log.Debug().Msg("will serve snapshot")

	// Run the xDS server
	cb := &test.Callbacks{Debug: log.Debug().Enabled()}
	srv := server.NewServer(ctx, cache, cb)
	go RunServer(ctx, srv, port)
	return mockserver, nil
}

func (m *MockServer) Update(config Config) {
	m.updateSnapshot(config)
}
