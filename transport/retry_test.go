package transport

import (
	"net/http"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

func TestRoundTripWithRetry_NoRetryPolicy_ShouldNotRetry(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	counter := 0
	mockRoundTrip := func(*http.Request) (*http.Response, error) {
		counter++
		return &http.Response{Status: "500 Internal Server Error", StatusCode: 500}, nil
	}

	roundTripWithRetry(req, mockRoundTrip, &routev3.RetryPolicy{})

	assert.Equal(t, 1, counter, "round trip should called only once")
}

func TestRoundTripWithRetry_RetryOn5xxPolicy_ShouldRetryMax(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	counter := 0
	mockRoundTrip := func(*http.Request) (*http.Response, error) {
		counter++
		return &http.Response{Status: "500 Internal Server Error", StatusCode: 500}, nil
	}

	const retriesNum = 5

	roundTripWithRetry(req, mockRoundTrip, &routev3.RetryPolicy{NumRetries: &wrappers.UInt32Value{Value: retriesNum}, RetryOn: "5xx"})

	assert.Equal(t, retriesNum, counter, "round trip should called maximum")
}

func TestRoundTripWithRetry_ServerEventuallyReturns200_ShouldReturn200(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	counter := 0
	const retriesNum = 5

	mockRoundTrip := func(*http.Request) (*http.Response, error) {
		counter++
		if counter < 2 {
			return &http.Response{Status: "500 Internal Server Error", StatusCode: 500}, nil
		} else {
			return &http.Response{Status: "200 Ok", StatusCode: 200}, nil
		}
	}

	resp, err := roundTripWithRetry(req, mockRoundTrip, &routev3.RetryPolicy{NumRetries: &wrappers.UInt32Value{Value: retriesNum}, RetryOn: "5xx"})

	assert.Equal(t, 2, counter, "round trip should called 2 times")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
