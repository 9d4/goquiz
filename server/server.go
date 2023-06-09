package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/94d/goquiz/auth"
	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/util"
	"github.com/94d/goquiz/web"
	"github.com/asdine/storm"
	"github.com/golang-jwt/jwt/v5"
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

	staticFs, err := fs.Sub(web.Assets(), "dist")
	if err != nil {
		log.Fatal(err)
	}
	s.router.NotFoundHandler = http.FileServer(http.FS(staticFs))

	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { s.JSON(w, map[string]string{"message": "Hello World"}) })

	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/", s.handleAuth)
	auth.HandleFunc("/login", s.handleAuthLogin).Methods("POST")
	auth.HandleFunc("/logout", s.withAuth(s.handleAuthLogout)).Methods("POST")

	quiz := api.PathPrefix("/quiz").Subrouter()
	quiz.HandleFunc("/", s.withUser(s.handleQuiz))
	quiz.HandleFunc("/data", s.withAuth(s.handleQuizData))
	quiz.HandleFunc("/start", s.withUser(s.handleQuizStart)).Methods("POST")
}

func (s *server) Serve() {
	listener, err := net.Listen("tcp", config.V.GetString("address"))
	throwError(err)

	log.Printf("Quiz          : %#v", entity.GetQuizName())
	log.Printf("Question count: %#v", entity.CountQuestions())
	log.Printf("User count    : %#v", entity.CountUsers())
	log.Printf("Server listening on http://%s\n", listener.Addr().String())
	err = http.Serve(listener, s.router)
	throwError(err)
}

func (s *server) JSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func (s *server) handleAuth(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		return
	}

	t, err := auth.ParseToken(token.Value, auth.KeyFunc([]byte(config.V.GetString("secret"))))
	if err != nil || !t.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		return
	}

	s.JSON(w, map[string]interface{}{"user": t.Claims})
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
		Path:  "/",
		Value: tokenStr,
	}
	http.SetCookie(w, cookie)

	s.JSON(w, map[string]interface{}{"user": token.Claims})
}

func (s *server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:    "token",
		Value:   "",
		Path:    "/",
		Expires: time.Now().AddDate(0, 0, -1),
	}

	http.SetCookie(w, cookie)
	s.JSON(w, map[string]string{"message": "Logout success"})
}

func (s *server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		t, err := auth.ParseToken(token.Value, auth.KeyFunc([]byte(config.V.GetString("secret"))))
		if err != nil || !t.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey("token"), t)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

func (s *server) withUser(next http.HandlerFunc) http.HandlerFunc {
	return s.withAuth(func(w http.ResponseWriter, r *http.Request) {
		tk, err := getAuth(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := tk.Claims.(jwt.MapClaims)
		if !ok || claims["username"] == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var usr entity.User
		if s.db.One("Username", claims["username"], &usr) != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey("user"), usr)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func getAuth(reqCtx context.Context) (*jwt.Token, error) {
	tk, ok := reqCtx.Value(ctxKey("token")).(*jwt.Token)
	if !ok {
		return nil, errors.New("unable to get authorized token")
	}

	return tk, nil
}

func getUser(reqCtx context.Context) (*entity.User, error) {
	usr, ok := reqCtx.Value(ctxKey("user")).(entity.User)
	if !ok {
		return nil, errors.New("unable to get authorized user")
	}

	return &usr, nil
}

func throwError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
