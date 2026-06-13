package services

import "errors"

var ErrEncryptionInvalid = errors.New("invalid encryption payload")

// The server stores ciphertext but never plaintext.
// For the scaffold we just validate required fields.
type EncryptionService struct{}

func NewEncryptionService(_ interface{}) *EncryptionService {
	return &EncryptionService{}
}

func (s *EncryptionService) ValidateVault(ciphertext, nonce, alg string) error {
	if ciphertext == "" || nonce == "" || alg == "" {
		return ErrEncryptionInvalid
	}
	return nil
}

