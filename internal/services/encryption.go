package services

import "errors"

var ErrEncryptionInvalid = errors.New("invalid encryption payload")

// EncryptionService validates client-supplied vault metadata.
// The server stores only ciphertext and never sees plaintext will content.
type EncryptionService struct{}

func NewEncryptionService(_ interface{}) *EncryptionService {
	return &EncryptionService{}
}

// supportedAlgorithms is the allowlist of accepted vault encryption algorithms.
// All entries are symmetric AEAD schemes; unauthenticated encryption is rejected.
var supportedAlgorithms = map[string]bool{
	"AES-GCM-256": true,
}

// ValidateVault checks that a vault ciphertext, nonce, and algorithm are
// structurally valid. It does not decrypt or inspect the plaintext.
func (s *EncryptionService) ValidateVault(ciphertext, nonce, alg string) error {
	if ciphertext == "" || nonce == "" || alg == "" {
		return ErrEncryptionInvalid
	}
	if !supportedAlgorithms[alg] {
		return ErrEncryptionInvalid
	}
	return nil
}
