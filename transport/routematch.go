package transport

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/rs/zerolog/log"
)

var cacheRegex sync.Map

func getOrCreateRegexp(expr string) (*regexp.Regexp, error) {
	a, found := cacheRegex.Load(expr)
	if found {
		return a.(*regexp.Regexp), nil
	}

	r, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	cacheRegex.Store(expr, r)

	return r, nil
}

func doesMatchRoutes(req *http.Request, routeMatch *route.RouteMatch) bool {
	pathMatched := doesPathMatch(req, routeMatch)
	headersMatched := doesHeaderMatch(req, routeMatch.Headers)
	queryMatch := doesQueryParametersMatch(req, routeMatch.QueryParameters)
	return pathMatched && headersMatched && queryMatch
}

func doesQueryParametersMatch(req *http.Request, queryParameters []*route.QueryParameterMatcher) bool {
	for _, queryParameter := range queryParameters {
		switch match := queryParameter.QueryParameterMatchSpecifier.(type) {
		case *route.QueryParameterMatcher_StringMatch:
			if !doesStringMatch(req.URL.Query().Get(queryParameter.Name), match.StringMatch) {
				return false
			}
		case *route.QueryParameterMatcher_PresentMatch:
			if !(req.URL.Query().Has(queryParameter.Name) == match.PresentMatch) {
				return false
			}
		}
	}
	return true
}

func doesHeaderMatch(req *http.Request, headers []*route.HeaderMatcher) bool {
	for _, header := range headers {
		switch headerMatch := header.HeaderMatchSpecifier.(type) {
		case *route.HeaderMatcher_ExactMatch:
			if !doesExactMatch(req.Header.Get(header.Name), headerMatch.ExactMatch, true, header.InvertMatch) {
				return false
			}
		case *route.HeaderMatcher_SafeRegexMatch:
			if !doesRegexMatch(req.Header.Get(header.Name), headerMatch.SafeRegexMatch.Regex, header.InvertMatch) {
				return false
			}

		case *route.HeaderMatcher_RangeMatch:
			if !doesRangeMatch(req.Header.Get(header.Name), headerMatch.RangeMatch.Start, headerMatch.RangeMatch.End, header.InvertMatch) {
				return false
			}
		case *route.HeaderMatcher_PresentMatch:
			if !((req.Header.Get(header.Name) != "") != header.InvertMatch) {
				return false
			}
		case *route.HeaderMatcher_PrefixMatch:
			if !doesPrefixMatch(req.Header.Get(header.Name), headerMatch.PrefixMatch, true, header.InvertMatch) {
				return false
			}
		case *route.HeaderMatcher_SuffixMatch:
			if !doesSuffixMatch(req.Header.Get(header.Name), headerMatch.SuffixMatch, true, header.InvertMatch) {
				return false
			}
		case *route.HeaderMatcher_ContainsMatch:
			if !doesContainMatch(req.Header.Get(header.Name), headerMatch.ContainsMatch, true, header.InvertMatch) {
				return false
			}
		case *route.HeaderMatcher_StringMatch:
			if header.InvertMatch && doesStringMatch(req.Header.Get(header.Name), headerMatch.StringMatch) {
				return false
			}
			if !header.InvertMatch && !doesStringMatch(req.Header.Get(header.Name), headerMatch.StringMatch) {
				return false
			}

		}
	}
	return true
}

func doesPathMatch(req *http.Request, routeMatch *route.RouteMatch) bool {
	// null is True
	caseSensitive := routeMatch.CaseSensitive == nil || routeMatch.CaseSensitive.GetValue()
	switch pathMatch := routeMatch.PathSpecifier.(type) {
	case *route.RouteMatch_Prefix:
		return doesPrefixMatch(req.URL.Path, pathMatch.Prefix, caseSensitive, false)
	case *route.RouteMatch_Path:
		return doesExactMatch(req.URL.Path, pathMatch.Path, caseSensitive, false)
	case *route.RouteMatch_SafeRegex:
		return doesRegexMatch(req.URL.Path, pathMatch.SafeRegex.Regex, false)
	case *route.RouteMatch_ConnectMatcher_:
		panic("not implemented")
	default:
		panic("not implemented")
	}

}

func doesStringMatch(str string, sm *matcher.StringMatcher) bool {
	caseSensitive := !sm.IgnoreCase

	switch patternMatch := sm.MatchPattern.(type) {
	case *matcher.StringMatcher_Exact:
		return doesExactMatch(str, patternMatch.Exact, caseSensitive, false)
	case *matcher.StringMatcher_Prefix:
		return doesPrefixMatch(str, patternMatch.Prefix, caseSensitive, false)
	case *matcher.StringMatcher_Suffix:
		return doesSuffixMatch(str, patternMatch.Suffix, caseSensitive, false)
	case *matcher.StringMatcher_SafeRegex:
		return doesRegexMatch(str, patternMatch.SafeRegex.Regex, false)
	case *matcher.StringMatcher_Contains:
		return doesContainMatch(str, patternMatch.Contains, caseSensitive, false)
	default:
		panic("not implemented")
	}
}

func doesSuffixMatch(str, suffix string, casSensitive bool, invertResult bool) bool {
	if !casSensitive {
		str = strings.ToLower(str)
		suffix = strings.ToLower(suffix)
	}

	return strings.HasSuffix(str, suffix) != invertResult
}

func doesContainMatch(str, substr string, casSensitive bool, invertResult bool) bool {
	if !casSensitive {
		str = strings.ToLower(str)
		substr = strings.ToLower(substr)
	}

	return strings.Contains(str, substr) != invertResult
}

func doesPrefixMatch(str, prefix string, casSensitive bool, invertResult bool) bool {
	if !casSensitive {
		str = strings.ToLower(str)
		prefix = strings.ToLower(prefix)
	}

	return strings.HasPrefix(str, prefix) != invertResult
}

func doesExactMatch(str1, str2 string, casSensitive bool, invertResult bool) bool {
	if !casSensitive {
		str1 = strings.ToLower(str1)
		str2 = strings.ToLower(str2)
	}

	return (str1 == str2) != invertResult
}

func doesRegexMatch(str, expr string, invertResult bool) bool {
	r, err := getOrCreateRegexp(expr)
	if err != nil {
		log.Error().Err(err).Str("expr", expr).Msg("invalid regex expression")
		return false
	}

	return r.Match([]byte(str)) != invertResult
}

func doesRangeMatch(strNum string, start, end int64, invertResult bool) bool {
	value, err := strconv.ParseInt(strNum, 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("value", strNum).Msg("header value is not a number")
		return false
	}

	return (value >= start && value < end) != invertResult
}
