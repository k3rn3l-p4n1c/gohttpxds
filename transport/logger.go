package transport

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func logRequest(req *http.Request) {
	log.Debug().Str("url", req.URL.String()).Msg("sending requests")
}
