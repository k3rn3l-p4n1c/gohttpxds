package transport

import (
	"log"
	"net/http"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/stretchr/testify/assert"
)

func TestDoesPathMatch_Prefix_ShouldPass(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	routeMatch := &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Prefix{
			Prefix: "/prefix",
		},
	}

	assert.True(t, doesPathMatch(req, routeMatch), "request should match")
}

func TestDoesPathMatch_WrongPrefix_ShouldFail(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	routeMatch := &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Prefix{
			Prefix: "/prefix2",
		},
	}

	assert.False(t, doesPathMatch(req, routeMatch), "request should not match")
}

func TestDoesPathMatch_Exact_ShouldPass(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	routeMatch := &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Path{
			Path: "/prefix/url",
		},
	}

	assert.True(t, doesPathMatch(req, routeMatch), "request should match")
}

func TestDoesPathMatch_Exact_ShouldFail(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	routeMatch := &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Path{
			Path: "/prefix/urll",
		},
	}

	assert.False(t, doesPathMatch(req, routeMatch), "request should not match")
}

func TestDoesPathMatch_Regex_ShouldPass(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url213", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	routeMatch := &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_SafeRegex{
			SafeRegex: &matcher.RegexMatcher{
				Regex: `.*url\d+$`,
			},
		},
	}

	assert.True(t, doesPathMatch(req, routeMatch), "request should match")
}

func TestDoesQueryParametersMatch_Present_ShouldPass(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url?q1=2&q2=3", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	queryParameters := []*route.QueryParameterMatcher{
		{
			Name: "q1",
			QueryParameterMatchSpecifier: &route.QueryParameterMatcher_PresentMatch{
				PresentMatch: true,
			},
		},
		{
			Name: "q2",
			QueryParameterMatchSpecifier: &route.QueryParameterMatcher_StringMatch{
				StringMatch: &matcher.StringMatcher{
					MatchPattern: &matcher.StringMatcher_Exact{
						Exact: "3",
					},
				},
			},
		},
	}

	assert.True(t, doesQueryParametersMatch(req, queryParameters), "request should match")

}
func TestDoesQueryParametersMatch_Present_ShouldFail(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://sub.domain.com/prefix/url?q1=2&q2=3", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	queryParameters := []*route.QueryParameterMatcher{
		{
			Name: "q1",
			QueryParameterMatchSpecifier: &route.QueryParameterMatcher_PresentMatch{
				PresentMatch: false,
			},
		},
		{
			Name: "q2",
			QueryParameterMatchSpecifier: &route.QueryParameterMatcher_StringMatch{
				StringMatch: &matcher.StringMatcher{
					MatchPattern: &matcher.StringMatcher_Exact{
						Exact: "3",
					},
				},
			},
		},
	}

	assert.False(t, doesQueryParametersMatch(req, queryParameters), "request should match")

}
