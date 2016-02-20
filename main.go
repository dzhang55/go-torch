package main

import (
	"net/http"
	_ "net/http/pprof" // import for side effects

	"github.com/hack4impact/audio-transcription-service/web"
)

func main() {
	router := web.NewRouter()
	middlewareRouter := web.ApplyMiddleware(router)

	// serve http
	http.Handle("/", middlewareRouter)
	http.ListenAndServe(":8080", nil)
}
