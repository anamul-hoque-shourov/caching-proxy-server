package proxy

import (
	"caching-proxy-server/cache"
	"io"
	"net/http"
	"time"
)

type Proxy struct {
	Cache         *cache.Cache
	UpstreamURL   string
	CacheDuration time.Duration
}

func NewProxy(upstreamURL string, cacheDuration time.Duration) *Proxy {
	return &Proxy{
		Cache:         cache.NewCache(),
		UpstreamURL:   upstreamURL,
		CacheDuration: cacheDuration,
	}
}

func (p *Proxy) fetchFromUpstream(r *http.Request) ([]byte, int, error) {
	fullURL := p.UpstreamURL + r.URL.RequestURI()

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Please send GET request", http.StatusMethodNotAllowed)
		return
	}

	cacheKey := r.URL.String()

	if data, found := p.Cache.Get(cacheKey); found {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	body, status, err := p.fetchFromUpstream(r)
	if err != nil {
		http.Error(w, "Upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}

	p.Cache.Set(cacheKey, body, p.CacheDuration)

	w.WriteHeader(status)
	w.Write(body)
}
