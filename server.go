package main

import (
	"context"
	"encoding/json"
	"login_server/data"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

const privateKey = "efjACGRY#WhxARaQ_Fhgm9Vp@zq=kn2Pn8$LNeqFcm#UZ3t7h?Bn@+Z?LsyWYatw"

type LoginServer struct {
	db     *data.MongoUserDB
	router *httprouter.Router
	port   int
}

// NewLoginServer ... Construct a new login server with the given parameters.
func NewLoginServer(port int, mongoURI string) LoginServer {

	db, err := data.NewMongoUserDB(context.Background(), mongoURI)
	if err != nil {
		zap.L().Fatal("Could not create mongo databse.", zap.Error(err))
	}

	ls := LoginServer{
		&db,
		httprouter.New(),
		port,
	}

	ls.setupRoutes()

	return ls
}

// Start ... Start the login server.
func (ls *LoginServer) Start() error {
	return http.ListenAndServe(":"+strconv.Itoa(ls.port), ls.router)
}

func (ls *LoginServer) setupRoutes() {
	ls.router.POST("/auth", ls.authenticate)
	ls.router.POST("/createuser", ls.createUser)
	ls.router.POST("/", ls.defaultRoute)
}

func getSignedKey(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"status":    "OK",
		"user":      userID,
		"ExpiresAt": 15000,
	})

	return token.SignedString([]byte(privateKey))
}

func (ls *LoginServer) defaultRoute(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	zap.L().Warn("Unhandled url", zap.String("url", r.URL.String()))
}

func (ls *LoginServer) authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// body := make([]byte, 256)
	// n, err := r.Body.Read(body)

	// // Not much we can do if we couldn't correctly read the data.
	// if err != io.EOF {
	// 	log.Fatalln(err)
	// }

	// // Decode the protobuffer message.
	// info := &Proto.LoginInfo{}
	// err = proto.Unmarshal(body[:n], info)

	// if err != nil {
	// 	panic(err)
	// }

	// log.Println("Login Attempt:", info)

	// // Try to get the user from the database.
	// user, err := s.db.GetUser(info.Email)

	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// // Now that we have the user we can check if the password is correct.
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(info.Password))

	// // If the password was not correct...
	// if err != nil {
	// 	log.Println(err)
	// 	w.Write([]byte("CREDENTIALS REJECTED"))
	// } else { // Password must have been correct.
	// 	log.Println("Found User:", user)

	// 	tokenStr, err := getSignedKey(user.ID.String())

	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}

	// 	w.Write([]byte(tokenStr))
	// }
}

func (ls *LoginServer) createUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	select {
	case <-r.Context().Done():
		zap.L().Info("Request handler was cancelled by the timeout handler.")
	default:
		// Metrics.
		start := time.Now()

		// Don't let the client send more than a kilobyte.
		r.Body = http.MaxBytesReader(w, r.Body, 1024)

		// Create a new json decoder.
		dec := json.NewDecoder(r.Body)

		// Don't let the client send fields we aren't expecting.
		dec.DisallowUnknownFields()

		var createUserReq data.CreateUserRequest
		err := dec.Decode(&createUserReq)

		// Couldn't decode the message, nothing to do here...
		if err != nil {
			zap.L().Error("Bad request from client", zap.Error(err))
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Construct our response.
		createUserResp := data.CreateUserResponse{Success: true, Reason: "OK"}

		// Check the db to see if the user already exists.
		_, err = ls.db.GetUser(r.Context(), createUserReq.Auth.Email)

		// If we found a user we can't actually make a new one so return already exists.
		if err == nil {
			createUserResp.Success = false
			createUserResp.Reason = "User already exists."
			zap.L().Info("Client requested to create a new user that already existed.", zap.Any("request", createUserReq))
		} else {
			// // Okay now we can create the user in teh database.
			userID, err := ls.db.CreateUser(r.Context(), &createUserReq)

			if err != nil {
				panic(err)
			}

			tokenStr, err := getSignedKey(userID.String())

			if err != nil {
				panic(err)
			}

			createUserResp.Token = tokenStr
		}

		zap.L().Info("Create user requests completed", zap.Duration("executionTime", time.Since(start)), zap.Any("response", createUserResp))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createUserResp)
	}
}
