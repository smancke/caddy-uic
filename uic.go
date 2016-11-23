package uic

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/lib-compose/cache"
	"github.com/tarent/lib-compose/composition"
	"net/http"
	"time"
)

//var globalCache = cache.NewCache("uic-cache", 1000, 1, time.Second*10)
var globalCache = cache.NewCache("uic-cache", 1, 1, time.Millisecond*1)

var SecondaryFetchRequestHeaders = []string{
	"Authorization",
	"Cache-Control",
	"Cookie",
	"Pragma",
	"Referer",
	"X-Forwarded-Host",
	"X-Correlation-Id",
	"X-Feature-Toggle",
}

type Uic struct {
	next               httpserver.Handler
	config             *Config
	compositionHandler *composition.CompositionHandler
}

func NewUic(next httpserver.Handler, config *Config) *Uic {
	h := &Uic{
		next:   next,
		config: config,
	}
	h.compositionHandler = composition.NewCompositionHandler(h.contentFetcherFactory)
	return h
}

func (h *Uic) contentFetcherFactory(r *http.Request) composition.FetchResultSupplier {
	replacer := httpserver.NewReplacer(r, nil, "")
	fetcher := composition.NewContentFetcher(nil)
	fetcher.Loader = composition.NewCachingContentLoader(globalCache)

	for i, f := range h.config.FetchRules {
		fd := composition.NewFetchDefinition(replacer.Replace(f.URL))
		fd.Priority = i

		fd.Header = copyHeaders(r.Header, fd.Header, SecondaryFetchRequestHeaders)
		fetcher.AddFetchJob(fd)
	}

	mainFD := composition.NewFetchDefinitionFromRequest(replacer.Replace(h.config.Upstream), r)
	mainFD.Priority = len(h.config.FetchRules)
	fetcher.AddFetchJob(mainFD)

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
	if !httpserver.Path(requestPath).Matches(h.config.Path) {
		return false
	}
	for _, p := range h.config.Except {
		if httpserver.Path(requestPath).Matches(p) {
			return false
		}
	}
	return true
}

// copyHeaders copies only the header contained in the the whitelist
// from src to dest. If dest is nil, it will be created.
// The dest will also be returned.
func copyHeaders(src, dest http.Header, whitelist []string) http.Header {
	if dest == nil {
		dest = http.Header{}
	}
	for _, k := range whitelist {
		headerValues := src[k]
		for _, v := range headerValues {
			dest.Add(k, v)
		}
	}
	return dest
}
