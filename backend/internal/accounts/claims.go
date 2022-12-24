package accounts

import jwt "github.com/golang-jwt/jwt/v4"

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TODO: Read from a kubernetes secret
var jwtKey = []byte("secret")
