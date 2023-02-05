package gohttpxds

import (
	"fmt"
	"net/http"

	"github.com/k3rn3l-p4n1c/gohttpxds/internal/xdscache"
	"github.com/k3rn3l-p4n1c/gohttpxds/internal/xdsclient"

	"github.com/k3rn3l-p4n1c/gohttpxds/transport"
	"google.golang.org/grpc"
)

func Register(serverURI string, creds grpc.DialOption) {
	httpXdsClient, err := NewHttpClient(serverURI, creds)
	if err != nil {
		panic(err.Error())
	}

	http.DefaultClient = httpXdsClient
}

func NewHttpClient(ServerURI string, Creds grpc.DialOption) (*http.Client, error) {
	xdsClient, err := xdsclient.New(xdsclient.ServerConfig{ServerURI: ServerURI, Creds: Creds})
	if err != nil {
		return nil, fmt.Errorf("fail to create xds client: %w", err)
	}
	xdsCache := xdscache.New(xdsClient)
	xdsCache.WatchCluster("")
	xdsCache.WatchListener("")
	return &http.Client{Transport: transport.New(http.DefaultTransport, xdsCache)}, nil
}
