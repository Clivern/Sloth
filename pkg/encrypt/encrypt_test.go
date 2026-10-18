// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package encrypt

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitEncryptDecrypt(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("app.encryption.key", "test-key")

	t.Run("round trip", func(t *testing.T) {
		cipher, err := Encrypt("secret-value")
		require.NoError(t, err)
		assert.NotEqual(t, "secret-value", cipher)

		plain, err := Decrypt(cipher)
		require.NoError(t, err)
		assert.Equal(t, "secret-value", plain)
	})

	t.Run("rejects wrong key", func(t *testing.T) {
		cipher, err := Encrypt("secret-value")
		require.NoError(t, err)

		viper.Set("app.encryption.key", "other-key")
		_, err = Decrypt(cipher)
		require.Error(t, err)
	})
}

func TestUnitEncryptionKey(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("app.encryption.key", "test-key")
	assert.Equal(t, "test-key", EncryptionKey())
}
