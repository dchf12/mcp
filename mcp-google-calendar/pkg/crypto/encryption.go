package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

// TokenEncryption はトークンの暗号化・復号化を行います
type TokenEncryption struct {
	passphrase []byte
}

// NewTokenEncryption は新しいTokenEncryptionインスタンスを作成します
func NewTokenEncryption() (*TokenEncryption, error) {
	// システム固有の識別子を使用して暗号化キーを生成
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// ホスト名とプロセスIDを組み合わせてパスフレーズを生成
	passphrase := []byte("gcal_mcp_" + hostname)
	
	return &TokenEncryption{
		passphrase: passphrase,
	}, nil
}

// Encrypt はデータを暗号化します
func (te *TokenEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	// PBKDF2を使用してキーを派生
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key(te.passphrase, salt, 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// GCMモードを使用して認証付き暗号化
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	
	// salt + ciphertext の形で返す
	result := make([]byte, len(salt)+len(ciphertext))
	copy(result, salt)
	copy(result[len(salt):], ciphertext)
	
	return result, nil
}

// Decrypt はデータを復号化します
func (te *TokenEncryption) Decrypt(encrypted []byte) ([]byte, error) {
	if len(encrypted) < 16 {
		return nil, errors.New("encrypted data too short")
	}

	// saltとciphertextを分離
	salt := encrypted[:16]
	ciphertext := encrypted[16:]

	// PBKDF2を使用してキーを派生
	key := pbkdf2.Key(te.passphrase, salt, 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}