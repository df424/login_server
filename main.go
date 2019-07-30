package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	//"golang.org/x/crypto/bcrypt"
	"./Proto"
	"io"
	"log"
	"net/http"
)

const privateKey = "efjACGRY#WhxARaQ_Fhgm9Vp@zq=kn2Pn8$LNeqFcm#UZ3t7h?Bn@+Z?LsyWYatw"

var mongoProxy MongoProxy

func handler(w http.ResponseWriter, r *http.Request) {
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

	// Try to get the user from the mongoproxy.
	user, err := mongoProxy.GetUser(info.Email)

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

func main() {
	log.Println("Starting login server...")

	var err error
	mongoProxy, err = NewMongoProxy()

	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", handler)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
