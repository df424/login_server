package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	//"golang.org/x/crypto/bcrypt"
	"./Proto"
	"io"
	"log"
	"net/http"
)

type server struct {
	db     *MongoProxy
	router *httprouter.Router
}

const privateKey = "efjACGRY#WhxARaQ_Fhgm9Vp@zq=kn2Pn8$LNeqFcm#UZ3t7h?Bn@+Z?LsyWYatw"

func (s *server) authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body := make([]byte, 256)
	n, err := r.Body.Read(body)

	// Not much we can do if we couldn't correctly read the data.
	if err != io.EOF {
		log.Fatalln(err)
	}

	// Decode the protobuffer message.
	info := &Proto.LoginInfo{}
	err = proto.Unmarshal(body[:n], info)

	if err != nil {
		panic(err)
	}

	log.Println("Login Attempt:", info)

	// Try to get the user from the database.
	user, err := s.db.GetUser(info.Email)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Found User:", user)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"status":    "OK",
		"ExpiresAt": 15000,
	})

	tokenStr, err := token.SignedString([]byte(privateKey))

	if err != nil {
		log.Println(err)
		return
	}

	w.Write([]byte(tokenStr))
}

func (s *server) createUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (s *server) defaultRoute(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("UNHANDLED ROUTE: ", r.URL)
}

func (s *server) setupRoutes() {
	s.router.POST("/auth", s.authenticate)
	s.router.POST("/createuser", s.createUser)
	s.router.POST("/", s.defaultRoute)
}

func main() {
	log.Println("Starting login server...")

	mp, err := NewMongoProxy()

	if err != nil {
		log.Fatalln(err)
	}

	s := server{
		&mp,
		httprouter.New(),
	}

	// Setup routes must be called before we start serving.
	s.setupRoutes()

	log.Fatalln(http.ListenAndServe(":8080", s.router))
}
