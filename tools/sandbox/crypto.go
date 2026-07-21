package sandbox

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io"
)

// Seal magic marker (8 bytes) used to locate sealed payload inside a binary.
const sealMagic = "KOOLSAND"

const sealVersion uint32 = 1

// seal encrypts packJSON with a one-time RSA keypair + AES-256-GCM.
// Layout:
//
//	Magic(8) | Version(u32 BE) |
//	PrivKeyLen(u32 BE) | PrivKey(PKCS#8 DER) |
//	WrappedDEKLen(u32 BE) | WrappedDEK |
//	NonceLen(u32 BE) | Nonce |
//	CipherLen(u32 BE) | Ciphertext
func seal(packJSON []byte) ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, packJSON, nil)

	wrappedDEK, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &priv.PublicKey, dek, nil)
	if err != nil {
		return nil, fmt.Errorf("wrap DEK: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(sealMagic)
	_ = binary.Write(&buf, binary.BigEndian, sealVersion)
	writeLenBytes(&buf, privDER)
	writeLenBytes(&buf, wrappedDEK)
	writeLenBytes(&buf, nonce)
	writeLenBytes(&buf, ciphertext)
	return buf.Bytes(), nil
}

func writeLenBytes(w *bytes.Buffer, b []byte) {
	_ = binary.Write(w, binary.BigEndian, uint32(len(b)))
	w.Write(b)
}

// unseal decrypts a sealed payload produced by seal.
func unseal(data []byte) (*PackBlob, error) {
	if len(data) < 8+4 {
		return nil, fmt.Errorf("sealed payload too short")
	}
	if string(data[:8]) != sealMagic {
		return nil, fmt.Errorf("invalid seal magic")
	}
	off := 8
	ver := binary.BigEndian.Uint32(data[off : off+4])
	off += 4
	if ver != sealVersion {
		return nil, fmt.Errorf("unsupported seal version %d", ver)
	}

	privDER, off, err := readLenBytes(data, off)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	wrappedDEK, off, err := readLenBytes(data, off)
	if err != nil {
		return nil, fmt.Errorf("read wrapped DEK: %w", err)
	}
	nonce, off, err := readLenBytes(data, off)
	if err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}
	ciphertext, _, err := readLenBytes(data, off)
	if err != nil {
		return nil, fmt.Errorf("read ciphertext: %w", err)
	}

	keyAny, err := x509.ParsePKCS8PrivateKey(privDER)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	priv, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	dek, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, wrappedDEK, nil)
	if err != nil {
		return nil, fmt.Errorf("unwrap DEK: %w", err)
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt pack: %w", err)
	}
	return unmarshalPackBlob(plain)
}

func readLenBytes(data []byte, off int) ([]byte, int, error) {
	if off+4 > len(data) {
		return nil, 0, fmt.Errorf("truncated length")
	}
	n := int(binary.BigEndian.Uint32(data[off : off+4]))
	off += 4
	if n < 0 || off+n > len(data) {
		return nil, 0, fmt.Errorf("truncated data (want %d)", n)
	}
	return data[off : off+n], off + n, nil
}

// findSealedPayload locates the first KOOLSAND magic in binary and returns
// the slice starting at that magic through end-of-parsed-structure (or EOF).
func findSealedPayload(bin []byte) ([]byte, error) {
	magic := []byte(sealMagic)
	idx := bytes.Index(bin, magic)
	if idx < 0 {
		return nil, fmt.Errorf("no sealed sandbox payload found in binary")
	}
	// Return from magic to end; unseal will parse length-prefixed fields.
	return bin[idx:], nil
}
