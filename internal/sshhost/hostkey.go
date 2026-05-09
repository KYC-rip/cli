package sshhost

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// loadOrCreateHostKey returns a Signer for an ed25519 host key, persisted at
// the supplied path. Creates the file (0600) on first run.
func loadOrCreateHostKey(path string) (ssh.Signer, string, error) {
	path = resolveHostKeyPath(path)
	if b, err := os.ReadFile(path); err == nil {
		signer, err := ssh.ParsePrivateKey(b)
		if err != nil {
			return nil, "", fmt.Errorf("parse host key %s: %w", path, err)
		}
		return signer, ssh.FingerprintSHA256(signer.PublicKey()), nil
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, "", err
	}
	pemBlock, err := ssh.MarshalPrivateKey(priv, "sshwap")
	if err != nil {
		return nil, "", err
	}
	out := pem.EncodeToMemory(pemBlock)
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return nil, "", err
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, "", err
	}
	_ = pub
	return signer, ssh.FingerprintSHA256(signer.PublicKey()), nil
}
