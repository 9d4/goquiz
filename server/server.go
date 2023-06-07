package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/94d/goquiz/auth"
	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/util"
	"github.com/asdine/storm"
	"github.com/gorilla/mux"
)

type server struct {
	router *mux.Router
	db     *storm.DB
}

func Start() {
	err := entity.Open()
	if err != nil {
		log.Fatal(err)
	}

	New(entity.DB()).Serve()
}

func New(db *storm.DB) *server {
	srv := &server{
		router: mux.NewRouter(),
		db:     db,
	}
	srv.SetupRoutes()
	return srv
}

func (s *server) SetupRoutes() {
	s.router.StrictSlash(true)

	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { s.JSON(w, map[string]string{"message": "Hello World"}) })

	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", s.handleAuthLogin)
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

func (s *server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var usr entity.User
	if err := s.db.One("Username", req.Username, &usr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if !util.CheckPasswordHash(req.Password, usr.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := auth.GenerateTokenRaw(&auth.ClaimsData{Fullname: usr.Fullname, Username: usr.Username})

	tokenStr, err := token.SignedString([]byte(config.V.GetString("secret")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenStr,
	}
	http.SetCookie(w, cookie)

	s.JSON(w, map[string]interface{}{"user": token.Claims})
	return
}

func throwError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
