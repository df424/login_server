package data

// AuthenticateRequest ... Data received via http during an authentication request.
type AuthenticationParams struct {
	Email    string
	Password string
}

// CreateUserRequest ... Data received via http during a create user request.
type CreateUserRequest struct {
	Auth  AuthenticationParams
	FName string
	LName string
}

// CreateUserResponse ... Data sent back via http from a succesful create user request.
type CreateUserResponse struct {
	Success bool
	Reason  string
	Token   string
}
