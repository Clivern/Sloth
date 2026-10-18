// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package locale

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/leonelquinteros/gotext"
	"github.com/stretchr/testify/assert"
)

func TestUnitLocale(t *testing.T) {
	t.Run("Load and T", func(t *testing.T) {
		locales = make(map[string]*gotext.Po)
		assert.NoError(t, Load(os.DirFS("locales")))

		assert.Equal(t, "Account is not active.", T("en", "account_not_active"))
		assert.Equal(t, "Account is not active.", T("", "account_not_active"))
		assert.Equal(t, "Le compte n'est pas actif.", T("fr", "account_not_active"))
		assert.Equal(t, "Account is not active.", T("de", "account_not_active"))
		assert.Equal(t, "missing_key", T("en", "missing_key"))
	})

	t.Run("GetLangFromRequest", func(t *testing.T) {
		locales = make(map[string]*gotext.Po)
		assert.NoError(t, Load(os.DirFS("locales")))

		assert.Equal(t, "en", GetLangFromRequest(nil))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "en", GetLangFromRequest(r))

		r.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en;q=0.8")
		assert.Equal(t, "fr", GetLangFromRequest(r))

		r.Header.Set("Accept-Language", "de-DE")
		assert.Equal(t, "de", GetLangFromRequest(r))
	})

	t.Run("TR", func(t *testing.T) {
		locales = make(map[string]*gotext.Po)
		assert.NoError(t, Load(os.DirFS("locales")))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept-Language", "en")
		assert.Equal(t, "Account is not active.", TR(r, "account_not_active"))
	})

	t.Run("Load skips non-po files", func(t *testing.T) {
		locales = make(map[string]*gotext.Po)
		fsys := fstest.MapFS{
			"en.po":     &fstest.MapFile{Data: []byte("msgid \"hi\"\nmsgstr \"Hello\"\n")},
			"notes.txt": &fstest.MapFile{Data: []byte("ignore")},
			"subdir":    &fstest.MapFile{Mode: os.ModeDir},
		}
		assert.NoError(t, Load(fsys))
		assert.Equal(t, "Hello", T("en", "hi"))
	})
}
