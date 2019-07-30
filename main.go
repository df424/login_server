package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"login_server/Proto"
	"net/http"
)

type server struct {
	db     *MongoProxy
	router *httprouter.Router
}

const privateKey = "efjACGRY#WhxARaQ_Fhgm9Vp@zq=kn2Pn8$LNeqFcm#UZ3t7h?Bn@+Z?LsyWYatw"

func getSignedKey(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"status":    "OK",
		"user":      userID,
		"ExpiresAt": 15000,
	})

	return token.SignedString([]byte(privateKey))
}

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

	// Now that we have the user we can check if the password is correct.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(info.Password))

	// If the password was not correct...
	if err != nil {
		log.Println(err)
		w.Write([]byte("CREDENTIALS REJECTED"))
	} else { // Password must have been correct.
		log.Println("Found User:", user)

		tokenStr, err := getSignedKey(user.ID.String())

		if err != nil {
			log.Println(err)
			return
		}

		w.Write([]byte(tokenStr))
	}
}

func (s *server) createUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body := make([]byte, 1024)
	n, err := r.Body.Read(body)

	// Not much we can do if we couldn't correctly read the data.
	if err != io.EOF {
		panic(err)
	}

	// Decode the protobuffer message.
	info := &Proto.NewUserInfo{}
	err = proto.Unmarshal(body[:n], info)

	// Couldn't decode the message, nothing to do here...
	if err != nil {
		panic(err)
	}

	log.Println("New User Attempt", info)

	// Check the db to see if the user already exists.
	_, err = s.db.GetUser(info.Email)
	// If we found a user we can't actually make a new one so return already exists.
	if err == nil {
		w.Write([]byte("User already exists."))
		log.Println("User already exists.")
		return
	}

	// Okay now we can create the user in teh database.
	userID, err := s.db.CreateUser(info)

	if err != nil {
		panic(err)
	}

	tokenStr, err := getSignedKey(userID.String())

	if err != nil {
		panic(err)
	}

	w.Write([]byte(tokenStr))
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
