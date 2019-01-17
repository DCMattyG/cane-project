package auth

import "crypto/rsa"

// MySigningKey Variable
var MySigningKey = []byte("secret")

// AuthTypes Variable
var AuthTypes = map[string]string{
	"basic":   "Basic",
	"session": "Session",
	"apikey":  "APIKey",
	"rfc3447": "Rfc3447",
}

// Basic Auth Type
type Basic struct {
	userName string
	password string
}

// Session Auth Type
type Session struct {
	userName       string
	password       string
	cookieLifetime int32
}

// APIKey Auth Type
type APIKey struct {
	key string
}

// Rfc3447 Auth Type
type Rfc3447 struct {
	publicKey  string
	privateKey *rsa.PrivateKey
}
