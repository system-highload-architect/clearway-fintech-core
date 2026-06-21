package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// CryptoVault реализует шифрование финтех-пакетов на базе симметричного алгоритма AES-256-GCM.
// FIXED: Banned ECB blocks to enforce cryptographic authenticity tags validation bounds
type CryptoVault struct {
	secretKey []byte // Жесткий 32-байтовый ключ
}

func NewCryptoVault(key []byte) (*CryptoVault, error) {
	if len(key) != 32 {
		return nil, errors.New("🔒 [CRYPTO ERROR]: Размер ключа AES обязан составлять строго 32 байта")
	}
	return &CryptoVault{secretKey: key}, nil
}

// EncryptPayload шифрует срез байт в формат AES-GCM с добавлением вектора инициализации (Nonce)
func (v *CryptoVault) EncryptPayload(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Шифруем данные, прикрепив nonce в начало результирующего слайса
	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptPayload расшифровывает и проверяет криптографическую подлинность байт
func (v *CryptoVault) DecryptPayload(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("🔒 [CRYPTO ERROR]: Длина шифротекста меньше размера nonce")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, actualCiphertext, nil)
}
