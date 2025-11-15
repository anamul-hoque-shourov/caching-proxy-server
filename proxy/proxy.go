package proxy

import (
	"caching-proxy-server/cache"
	"fmt"
	"io"
	"log"
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

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func (p *Proxy) fetchFromUpstream(r *http.Request) ([]byte, int, http.Header, error) {
	log.Printf("UPSTREAM   - GET %s", r.URL.RequestURI())

	fullURL := p.UpstreamURL + r.URL.RequestURI()

	req, err := http.NewRequest(r.Method, fullURL, nil)
	if err != nil {
		return nil, 0, nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, err
	}

	if resp.StatusCode >= 400 {
		return body, resp.StatusCode, resp.Header, fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	return body, resp.StatusCode, resp.Header, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Please send GET request", http.StatusMethodNotAllowed)
		return
	}

	cacheKey := r.URL.String()

	if data, found := p.Cache.Get(cacheKey); found {
		log.Printf("CACHE HIT  - %s %s", r.Method, r.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	log.Printf("CACHE MISS - %s %s", r.Method, r.URL.String())

	body, status, headers, err := p.fetchFromUpstream(r)
	if err != nil {
		w.WriteHeader(status)
		w.Write(body)
		return
	}

	for key, values := range headers {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	p.Cache.Set(cacheKey, body, p.CacheDuration)

	w.WriteHeader(status)
	w.Write(body)
}
