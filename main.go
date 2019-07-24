package main

import (
	"github.com/dgrijalva/jwt-go"
	//"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
)

const privateKey = "efjACGRY#WhxARaQ_Fhgm9Vp@zq=kn2Pn8$LNeqFcm#UZ3t7h?Bn@+Z?LsyWYatw"

var mongoProxy MongoProxy

func handler(w http.ResponseWriter, r *http.Request) {
	x := make([]byte, 128)
	log.Println(r.URL.Path)
	n, err := r.Body.Read(x)

	if err != io.EOF {
		log.Fatalln(err)
	}

	log.Println(string(x[:n]))

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

	_, err := NewMongoProxy()

	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", handler)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
