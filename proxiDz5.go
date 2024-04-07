package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

var number int = 0

func main() {
	server1URL, _ := url.Parse("http://localhost:8080")
	server2URL, _ := url.Parse("http://localhost:8082")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if number == 0 {
			proxy := httputil.NewSingleHostReverseProxy(server1URL)
			proxy.ServeHTTP(w, r)
			number++
		} else {
			proxy2 := httputil.NewSingleHostReverseProxy(server2URL)
			proxy2.ServeHTTP(w, r)
			number--
		}
	})

	http.ListenAndServe(":9000", nil)
}
