package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type KeypairService struct {
}

type Keypair struct {
	PriKey string `json:"pri_key"`
	PubKey string `json:"pub_key"`
}

var (
	ErrKeypairNotConfigured = errors.New("KeyFile is not configured")
	ErrInvalidPrivateKey    = errors.New("invalid private key")
)

// privateKeyPath derives the id_ed25519 (private) path from the configured
// id_ed25519.pub (public) KeyFile path - the same convention hbbs itself uses:
// the private key sits next to the public key, same name, without ".pub".
func (ks *KeypairService) privateKeyPath() (string, error) {
	if Config.Rustdesk.KeyFile == "" {
		return "", ErrKeypairNotConfigured
	}
	return strings.TrimSuffix(Config.Rustdesk.KeyFile, ".pub"), nil
}

// GetKeypair reads the current keypair from disk (the same files hbbs
// reads/writes), without generating anything.
func (ks *KeypairService) GetKeypair() (*Keypair, error) {
	priPath, err := ks.privateKeyPath()
	if err != nil {
		return nil, err
	}
	kp := &Keypair{}
	if b, err := os.ReadFile(priPath); err == nil {
		kp.PriKey = strings.TrimSpace(string(b))
	}
	if b, err := os.ReadFile(Config.Rustdesk.KeyFile); err == nil {
		kp.PubKey = strings.TrimSpace(string(b))
	}
	return kp, nil
}

// writeKeypair persists a keypair to the id_ed25519/id_ed25519.pub files and
// updates the in-memory Config.Rustdesk.Key so already-running endpoints
// (web client bootstrap, PC client API, admin config) serve the new public
// key immediately, without requiring a process restart.
func (ks *KeypairService) writeKeypair(priKey ed25519.PrivateKey) (*Keypair, error) {
	priPath, err := ks.privateKeyPath()
	if err != nil {
		return nil, err
	}
	pubKey := priKey.Public().(ed25519.PublicKey)

	priB64 := base64.StdEncoding.EncodeToString(priKey)
	pubB64 := base64.StdEncoding.EncodeToString(pubKey)

	if dir := filepath.Dir(priPath); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	if err := os.WriteFile(priPath, []byte(priB64), 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(Config.Rustdesk.KeyFile, []byte(pubB64), 0644); err != nil {
		return nil, err
	}

	Config.Rustdesk.Key = pubB64

	return &Keypair{PriKey: priB64, PubKey: pubB64}, nil
}

// ResetKeypair generates a brand new random Ed25519 keypair, in the same
// 64-byte-private/32-byte-public, base64-standard-encoded format hbbs itself
// writes (Go's crypto/ed25519 uses the same seed||pubkey layout as the
// reference/libsodium implementation hbbs is built on, so the files it
// produces are loadable by hbbs without any conversion).
//
// IMPORTANT: restart (or otherwise notify) hbbs/hbbr after calling this so
// they pick up the new key - this only updates the files and this process's
// own in-memory copy.
func (ks *KeypairService) ResetKeypair() (*Keypair, error) {
	_, priKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return ks.writeKeypair(priKey)
}

// SetKeypair installs a caller-provided private key (base64-encoded, either
// the 32-byte seed or the full 64-byte seed||pubkey private key) and derives
// the matching public key, rather than generating a random one.
func (ks *KeypairService) SetKeypair(priKeyB64 string) (*Keypair, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(priKeyB64))
	if err != nil {
		return nil, ErrInvalidPrivateKey
	}

	var priKey ed25519.PrivateKey
	switch len(raw) {
	case ed25519.SeedSize: // 32 bytes
		priKey = ed25519.NewKeyFromSeed(raw)
	case ed25519.PrivateKeySize: // 64 bytes
		priKey = ed25519.PrivateKey(raw)
	default:
		return nil, ErrInvalidPrivateKey
	}

	return ks.writeKeypair(priKey)
}
