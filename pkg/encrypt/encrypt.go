// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package encrypt

import (
	"github.com/clivern/sloth/pkg/util"
	"github.com/spf13/viper"
)

// EncryptionKey returns the app encryption key.
func EncryptionKey() string {
	return viper.GetString("app.encryption.key")
}

// Encrypt seals plaintext with the app encryption key.
func Encrypt(plaintext string) (string, error) {
	return util.Encrypt(EncryptionKey(), plaintext)
}

// Decrypt opens ciphertext sealed by Encrypt.
func Decrypt(ciphertext string) (string, error) {
	return util.Decrypt(EncryptionKey(), ciphertext)
}
