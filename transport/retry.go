package transport

import (
	"errors"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

const (
	BaseIntervalMiliseconds int64 = 25
)

var (
	random *rand.Rand
)

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	random = rand.New(source)
}

func roundTripWithRetry(req *http.Request, roundTrip func(*http.Request) (*http.Response, error), retryPolicy *routev3.RetryPolicy) (*http.Response, error) {
	for retryNum := 1; ; retryNum++ {
		resp, err := roundTrip(req)

		if !shouldRetry(req, resp, err, retryPolicy) {
			return resp, err
		}

		if retryNum >= int(retryPolicy.GetNumRetries().GetValue()) {
			return resp, err
		}

		backoff(retryNum, retryPolicy.GetRetryBackOff())
	}

}

func shouldRetry(req *http.Request, resp *http.Response, responseError error, retryPolicy *routev3.RetryPolicy) bool {
	if contains(retryPolicy.GetRetriableStatusCodes(), resp.StatusCode) {
		return true
	}

	for _, retryOn := range strings.Split(retryPolicy.GetRetryOn(), ",") {
		switch retryOn {
		case "5xx":
			if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
				return true
			}
		case "gateway-error":
			if resp.StatusCode == 502 || resp.StatusCode == 503 || resp.StatusCode == 504 {
				return true
			}
		case "reset":
			if errors.Is(responseError, http.ErrServerClosed) ||
				errors.Is(responseError, net.ErrClosed) {
				return true
			}
			if err, ok := responseError.(net.Error); ok && err.Timeout() {
				return true
			}
		case "connect-failure":
			if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
				return true
			}
			if err, ok := responseError.(net.Error); ok && err.Timeout() {
				return true
			}
		case "retriable-4xx":
			if resp.StatusCode == 409 {
				return true
			}
		case "retriable-status-codes":
			for _, statusCode := range strings.Split(req.Header.Get("x-envoy-retriable-status-codes"), ",") {
				if strconv.Itoa(resp.StatusCode) == statusCode {
					return true
				}
			}
		case "retriable-headers":
			for _, headerName := range strings.Split(req.Header.Get("x-envoy-retriable-header-names"), ",") {
				if resp.Header.Get(headerName) != "" {
					return true
				}
			}
		}
	}

	return false
}

func contains(s []uint32, e int) bool {
	for _, a := range s {
		if int(a) == e {
			return true
		}
	}
	return false
}

func backoff(retryNum int, retryBackOff *routev3.RetryPolicy_RetryBackOff) {
	var baseIntervalMiliseconds = BaseIntervalMiliseconds
	if retryBackOff.GetBaseInterval() != nil {
		baseIntervalMiliseconds = retryBackOff.GetBaseInterval().AsDuration().Milliseconds()
	}
	var maxIntervalMiliseconds = 10 * baseIntervalMiliseconds

	m := math.Pow(2, float64(retryNum)) - 1
	bound := int64(math.Min(m*float64(baseIntervalMiliseconds), float64(maxIntervalMiliseconds)))

	time.Sleep(time.Duration(random.Int63n(bound)) * time.Millisecond)
}
