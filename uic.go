package uic

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/lib-compose/cache"
	"github.com/tarent/lib-compose/composition"
	"net/http"
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
	except             []string
	compositionHandler *composition.CompositionHandler
}

func NewUic(next httpserver.Handler, path string, upstream string, fetchRules []Fetch, except []string) *Uic {
	h := &Uic{
		next:       next,
		path:       path,
		upstream:   upstream,
		fetchRules: fetchRules,
		except:     except,
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
	if h.matches(r.URL.Path) {
		h.compositionHandler.ServeHTTP(w, r)
		return 0, nil
	} else {
		return h.next.ServeHTTP(w, r)
	}
}

func (h *Uic) matches(requestPath string) bool {
	if !httpserver.Path(requestPath).Matches(h.path) {
		return false
	}
	for _, p := range h.except {
		if httpserver.Path(requestPath).Matches(p) {
			return false
		}
	}
	return true
}
