package uic

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/lib-compose/cache"
	"github.com/tarent/lib-compose/composition"
	"net/http"
	"strings"
	"time"
)

var globalCache = cache.NewCache("uic-cache", 1000, 1, time.Second*10)

type Fetch struct {
	*composition.FetchDefinition
}

type Uic struct {
	next               httpserver.Handler
	path               string
	upstream           string
	fetchRules         []Fetch
	compositionHandler *composition.CompositionHandler
}

func NewUic(next httpserver.Handler, path string, upstream string, fetchRules []Fetch) *Uic {
	h := &Uic{
		next:       next,
		path:       path,
		upstream:   upstream,
		fetchRules: fetchRules,
	}
	h.compositionHandler = composition.NewCompositionHandler(h.contentFetcherFactory)
	return h
}

func (h *Uic) contentFetcherFactory(r *http.Request) composition.FetchResultSupplier {
	fetcher := composition.NewContentFetcher(nil)
	fetcher.Loader = composition.NewCachingContentLoader(globalCache)

	for _, f := range h.fetchRules {
		fetcher.AddFetchJob(f.FetchDefinition)
	}

	fetcher.AddFetchJob(composition.NewFetchDefinitionFromRequest(h.upstream, r))

	return fetcher
}

func (h *Uic) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if strings.HasPrefix(r.URL.Path, h.path) {
		h.compositionHandler.ServeHTTP(w, r)
		return 0, nil
	} else {
		return h.next.ServeHTTP(w, r)
	}
}
