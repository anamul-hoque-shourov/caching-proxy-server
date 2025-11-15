package main

import (
	"caching-proxy-server/proxy"
	"log"
	"net/http"
	"time"
)

func main() {
	upstream := "http://localhost:3000"
	cacheTTL := 10 * time.Second

	p := proxy.NewProxy(upstream, cacheTTL)

	http.HandleFunc("/", p.ServeHTTP)

	log.Println("Caching proxy server running on :5000")
	log.Printf("Upstream: %s\n", upstream)

	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatal("server failed:", err)
	}
}
