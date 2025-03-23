package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}
type httpServer struct {
	handler http.Handler
	address string
}

func main() {
	//Init

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	httpServ := httpServer{handler: mux, address: ":8080"}
	fileServer := http.FileServer(http.Dir("."))
	//Handlers
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)
	mux.HandleFunc("GET /admin/metrics", apiCfg.writeHits)
	mux.HandleFunc("GET /api/healthz", HandlerHealthz)
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.MiddlewareMetricsInc((fileServer))))
	//Serve
	http.ListenAndServe(httpServ.address, httpServ.handler)
}
