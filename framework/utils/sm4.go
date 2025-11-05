package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/tjfoc/gmsm/sm4"
)

// SM4Encrypt 使用SM4-ECB模式加密（PKCS#7填充）
// key必须为16字节
func SM4Encrypt(key []byte, plaintext string) (string, error) {
	if len(key) != sm4.BlockSize {
		return "", errors.New("sm4的key必须是16字节")
	}

	// PKCS#7填充：使明文长度为16的倍数
	padding := sm4.BlockSize - len(plaintext)%sm4.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	data := append([]byte(plaintext), padText...)

	// ECB模式加密（无需IV）
	cipher, err := sm4.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, len(data))
	for i := 0; i < len(data); i += sm4.BlockSize {
		cipher.Encrypt(cipherText[i:i+sm4.BlockSize], data[i:i+sm4.BlockSize])
	}

	// Base64编码输出
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// SM4Decrypt 使用SM4-ECB模式解密（自动去除PKCS#7填充）
// key必须为16字节
func SM4Decrypt(key []byte, ciphertext string) (string, error) {
	if len(key) != sm4.BlockSize {
		return "", errors.New("sm4的key必须是16字节")
	}

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data)%sm4.BlockSize != 0 {
		return "", errors.New("位数非法")
	}

	// ECB模式解密
	cipher, err := sm4.NewCipher(key)
	if err != nil {
		return "", err
	}

	plainText := make([]byte, len(data))
	for i := 0; i < len(data); i += sm4.BlockSize {
		cipher.Decrypt(plainText[i:i+sm4.BlockSize], data[i:i+sm4.BlockSize])
	}

	// 去除PKCS#7填充
	padding := int(plainText[len(plainText)-1])
	if padding < 1 || padding > sm4.BlockSize {
		return "", errors.New("位数非法")
	}
	return string(plainText[:len(plainText)-padding]), nil
}
