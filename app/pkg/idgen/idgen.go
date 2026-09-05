// Package idgen builds human-readable, URL-safe ids for forum topics and
// comments (e.g. "tihie-interfeysy", "c-4f9a2b") the way the frontend mock
// data does, instead of opaque UUIDs.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

var translit = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// Slugify transliterates Cyrillic to Latin and lowercases/dashes the rest,
// e.g. "Тихие интерфейсы?" -> "tihie-interfeysy".
func Slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if t, ok := translit[r]; ok {
			b.WriteString(t)
			continue
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	slug := b.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

// RandomHex returns a random lowercase hex string of the given byte length.
func RandomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}

// RandomCode returns a random 6-digit numeric one-time code.
func RandomCode() string {
	const digits = "0123456789"
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "000000"
	}
	out := make([]byte, 6)
	for i, b := range buf {
		out[i] = digits[int(b)%len(digits)]
	}
	return string(out)
}
