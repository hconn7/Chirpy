package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/hconn7/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	Platform       string
	JwtSecret      string
}
type httpServer struct {
	handler http.Handler
	address string
}

func main() {
	//Init

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	tokenSecret := os.Getenv("SECRET_TOKEN")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Print(err)
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		Platform:       platform,
		JwtSecret:      tokenSecret,
	}
	mux := http.NewServeMux()
	httpServ := httpServer{handler: mux, address: ":8080"}
	fileServer := http.FileServer(http.Dir("."))
	//Handlers
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerValidateRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefreshToken)
	mux.HandleFunc("POST /api/login", apiCfg.handlerValidateLogin)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetUsers)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpsID}", apiCfg.hanlerGetSingleChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.writeHits)
	mux.HandleFunc("GET /api/healthz", HandlerHealthz)
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.MiddlewareMetricsInc((fileServer))))
	//Serve
	http.ListenAndServe(httpServ.address, httpServ.handler)
}
