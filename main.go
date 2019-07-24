package main

import (
	"github.com/dgrijalva/jwt-go"
	//"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
)

const privateKey = "efjACGRY#WhxARaQ_Fhgm9Vp@zq=kn2Pn8$LNeqFcm#UZ3t7h?Bn@+Z?LsyWYatw"

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

	mongoProxy, err := NewMongoProxy()

	if err != nil {
		log.Fatalln(err)
	}

	go mongoProxy.StartProcessing()
	mongoProxy.queries <- "Hello!"
	mongoProxy.queries <- "World!"
	mongoProxy.queries <- "Cool stuff!"
	mongoProxy.Shutdown()
	//http.HandleFunc("/", handler)
	//log.Fatalln(http.ListenAndServe(":8080", nil))
}
