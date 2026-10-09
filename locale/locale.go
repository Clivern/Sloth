// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package locale

import (
	"io/fs"
	"net/http"
	"reflect"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/samber/lo"
)

var locales = make(map[string]*gotext.Po)

const defaultLang = "en"

// Load loads .po files from the given filesystem. Panics on any failure.
func Load(f fs.FS) error {
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		panic("locale: failed to read dir: " + err.Error())
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.HasSuffix(name, ".po") {
			continue
		}

		data, err := fs.ReadFile(f, name)
		if err != nil {
			panic("locale: failed to read " + name + ": " + err.Error())
		}

		po := gotext.NewPo()
		po.Parse(data)
		locales[strings.TrimSuffix(name, ".po")] = po
	}

	return nil
}

// T returns the translation for the given language and key.
func T(lang, key string) string {
	if lo.IsEmpty(lang) {
		lang = defaultLang
	}

	po, ok := locales[lang]
	if ok && po != nil {
		if msg := getMsg(po, key); lo.IsNotEmpty(msg) {
			return msg
		}
	}

	if lang != defaultLang {
		en := locales[defaultLang]
		if en != nil {
			if msg := getMsg(en, key); lo.IsNotEmpty(msg) {
				return msg
			}
		}
	}

	return key
}

// getMsg looks up the translation for key in po. Key is a msgid, not a format string.
func getMsg(po *gotext.Po, key string) string {
	return reflect.ValueOf(po).
		MethodByName("Get").
		Call([]reflect.Value{reflect.ValueOf(key)})[0].
		String()
}

// GetLangFromRequest returns the preferred language from the request.
func GetLangFromRequest(r *http.Request) string {
	if r == nil {
		return defaultLang
	}

	al := r.Header.Get("Accept-Language")
	if lo.IsEmpty(al) {
		return defaultLang
	}

	for p := range strings.SplitSeq(strings.TrimSpace(al), ",") {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, ";"); idx > 0 {
			p = p[:idx]
		}

		p = strings.TrimSpace(p)
		if len(p) >= 2 {
			lang := strings.ToLower(p[:2])
			if _, ok := locales[lang]; ok {
				return lang
			}

			return lang
		}
	}

	return defaultLang
}

// TR is a shorthand: T(GetLangFromRequest(r), key).
func TR(r *http.Request, key string) string {
	return T(GetLangFromRequest(r), key)
}
