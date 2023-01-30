package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/k3rn3l-p4n1c/gohttpxds"
	"github.com/k3rn3l-p4n1c/gohttpxds/test/example"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go example.Run(ctx, example.GenerateSnapshot())

	req, err := http.NewRequest("GET", "xds://test/todos/1", nil)
	if err != nil {
		panic(err.Error())
	}

	client, err := gohttpxds.NewHttpClient("127.0.0.1:18000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err.Error())
	}

	// client := &http.Client{}

	time.Sleep(1 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	assert.Equal(t, 200, resp.StatusCode)
	fmt.Printf("%v\n", resp)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%v\n", string(body))
}
