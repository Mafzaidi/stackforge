package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/golang-jwt/jwt/v5"
)

// generateTestKeyPair generates an RSA key pair for testing
func generateTestKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// createTestToken creates a JWT token for testing
func createTestToken(privateKey *rsa.PrivateKey, kid string, claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(claims))
	token.Header["kid"] = kid
	return token.SignedString(privateKey)
}

// encodePublicKeyN encodes the RSA public key modulus to base64url format
func encodePublicKeyN(publicKey *rsa.PublicKey) string {
	return base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
}

// encodePublicKeyE encodes the RSA public key exponent to base64url format
func encodePublicKeyE(publicKey *rsa.PublicKey) string {
	return base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())
}
