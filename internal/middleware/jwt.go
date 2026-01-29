package middleware

import (
    "net/http"
    "log"

    "github.com/go-chi/jwtauth/v5"
)

// JWTAuth instance
var tokenAuth *jwtauth.JWTAuth

// Secret key untuk signing token (sebaiknya ambil dari ENV di production)
var jwtSecret []byte

// InitJWT initializes JWT with secret key
func InitJWT(secret string) {
    jwtSecret = []byte(secret)
    tokenAuth = jwtauth.New("HS256", jwtSecret, nil)
}

// JWTMiddleware verifies token and attaches user info to context
func JWTMiddleware(next http.Handler) http.Handler {
    return jwtauth.Verifier(tokenAuth)(jwtauth.Authenticator(tokenAuth)(next))
}

// ExtractClaims extracts JWT claims from request
func ExtractClaims(r *http.Request) (map[string]interface{}, error) {
    _, claims, err := jwtauth.FromContext(r.Context())
    if err != nil {
        log.Printf("JWT ExtractClaims error: %v", err)
        return nil, err
    }
    log.Printf("JWT claims: %+v", claims)
    return claims, err
}

// GetTokenAuth returns the jwtauth instance (useful for signing tokens in handlers)
func GetTokenAuth() *jwtauth.JWTAuth {
    return tokenAuth
}

// GetSecret returns the JWT secret (if needed for manual signing)
func GetSecret() []byte {
    return jwtSecret
}