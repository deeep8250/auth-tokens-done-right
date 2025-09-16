package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sync"
)

// Manager holds one RSA key pair for the service.
type Manager struct {
	mu   sync.RWMutex
	kid  string          // Key ID (a short label for this key)
	priv *rsa.PrivateKey // PRIVATE key (server-only, used to sign JWTs)
	pub  rsa.PublicKey   // PUBLIC key (shared, used to verify JWTs)
}

// New creates a fresh RSA-2048 key pair at startup.
// (Good for learning. In prod you'll load from files/KMS so keys persist across restarts.)
func New() (*Manager, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &Manager{
		kid:  kidOf(&priv.PublicKey),
		priv: priv,
		pub:  priv.PublicKey,
	}, nil
}

// Private returns the PRIVATE key (for signing).
func (m *Manager) Private() *rsa.PrivateKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.priv
}

// Public returns the PUBLIC key (for verifying).
func (m *Manager) Public() rsa.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pub
}

// JWKS returns the public key in JWKS (JSON Web Key Set) format.
// Clients/middlewares will fetch this to verify your JWTs.
func (m *Manager) JWKS() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// RSA public key components:
	// - N (modulus) big integer
	// - E (exponent) small integer
	n := base64.RawURLEncoding.EncodeToString(m.pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(m.pub.E)).Bytes())

	// Minimal JWKS with a single RSA key entry
	doc := map[string]any{
		"keys": []map[string]string{{
			"kty": "RSA",   // key type
			"kid": m.kid,   // key id (lets verifiers pick the right key)
			"use": "sig",   // this key is for signatures
			"alg": "RS256", // intended signing algorithm
			"n":   n,       // modulus (base64url)
			"e":   e,       // exponent (base64url)
		}},
	}

	b, _ := json.Marshal(doc)
	return b
}

// kidOf creates a short, stable Key ID from the public key bytes.
func kidOf(pub *rsa.PublicKey) string {
	spki, _ := x509.MarshalPKIXPublicKey(pub) // standard-encoded public key bytes
	sum := sha1.Sum(spki)                     // short fingerprint (fine for KID)
	return base64.RawURLEncoding.EncodeToString(sum[:8])
}
