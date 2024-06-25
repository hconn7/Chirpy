package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hconn7/Chirpy/sources/dataBase"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	jwtSecret      string
	fileServerHits int
	DB             *database.DB
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) writeHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf=8")
	w.WriteHeader(http.StatusOK)
	hitsMessage := fmt.Sprintf(`
		<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>
		`, cfg.fileServerHits)
	w.Write([]byte(hitsMessage))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	const filepathRoot = "."
	const port = ":8080" // Corrected port format
	dbPath := "database.json"
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileServerHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUpgrade)

	mux.HandleFunc("GET /api/reset", apiCfg.resetHits)
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsRetrieve)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGet)
	mux.HandleFunc("GET /admin/metrics", apiCfg.writeHits)

	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)

	// Static file server
	fileserver := http.FileServer(http.Dir(filepathRoot))
	strippedFileServer := http.StripPrefix("/app", fileserver)
	wrapperHandler := apiCfg.middlewareMetricsInc(strippedFileServer)
	mux.Handle("/app/", wrapperHandler)

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
