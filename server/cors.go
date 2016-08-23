package main

import (
	"net/http"
)

func cors(inner http.HandlerFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		needCors := len(origin) > 0
		if needCors {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if r.Method != "OPTIONS" {
			inner.ServeHTTP(w, r)
			return
		}

		defer w.WriteHeader(http.StatusOK)

		AccessControlRequestHeaders := r.Header.Get("Access-Control-Request-Headers")
		if len(AccessControlRequestHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", AccessControlRequestHeaders)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Allow", "HEAD,GET,POST,PUT,DELETE,OPTIONS")
	}
	return http.HandlerFunc(fn)
}
