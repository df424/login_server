package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User ... Struct for dealing with user data.
type User struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
}

// String ... Get string representation of the User.
func (u User) String() string {
	return fmt.Sprintf("{ID:%s, email:%s, password:%s}", u.ID.String(), u.Email, u.Password)
}
