package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/db"
	"github.com/gorilla/mux"
)

type server struct {
	router *mux.Router
}

func Start() {
	err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	db.Migrate()
	New(db.DB()).Serve()
}

func New(db *sql.DB) *server {
	srv := &server{
		router: mux.NewRouter(),
	}
	srv.SetupRoutes()
	return srv
}

func (s *server) SetupRoutes() {
	s.router.StrictSlash(true)

	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { s.JSON(w, map[string]string{"message": "Hello World"}) })
}

func (s *server) Serve() {
	listener, err := net.Listen("tcp", config.V.GetString("address"))
	throwError(err)

	log.Printf("Server listening on %s\n", listener.Addr().String())
	err = http.Serve(listener, s.router)
	throwError(err)
}

func (s *server) JSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func throwError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
