package toml

import (
	"bytes"
	"fmt"
	"unicode"
	"unicode/utf8"
)

func Parse(toml []byte) (map[string]string, error) {
	lines := bytes.Split(toml, []byte{'\n'})

	res := make(map[string]string)

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		kv := bytes.Split(trimmed, []byte{'='})
		if len(kv) != 2 {
			return nil, fmt.Errorf(
				"error on line %d. expected key = value",
				i+1,
			)
		}

		k, v, err := parseKeyValue(kv[0], kv[1])
		if err != nil {
			return nil, fmt.Errorf("error on line %d: %w", i+1, err)
		}

		res[string(k)] = string(v)
	}

	return res, nil
}

func parseKeyValue(key, value []byte) ([]byte, []byte, error) {
	sanitizedKey, err := sanitizeKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to sanitize key %s: %w",
			string(bytes.TrimSpace(key)),
			err,
		)
	}

	sanitizedValue := bytes.TrimSpace(value)

	switch sanitizedValue[0] {
	case '\'':
		sanitizedValue, err = stripCharacter(sanitizedValue, '\'')
		if err != nil {
			return nil, nil, err
		}
	case '"':
		sanitizedValue, err = stripCharacter(sanitizedValue, '"')
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("expected value to start with ' or \"")
	}

	return sanitizedKey, sanitizedValue, nil
}

func sanitizeKey(key []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(key)

	b := trimmed
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError && size == 1 {
			return nil, fmt.Errorf("%c is not a valid utf8 character", r)
		}

		if !unicode.IsLetter(r) && r != '-' {
			return nil, fmt.Errorf(
				"%c is not a valid character. a key must only contain letters or a '-'",
				r,
			)
		}
		b = b[size:]
	}

	return trimmed, nil
}

func stripCharacter(b []byte, ch byte) ([]byte, error) {
	if b[len(b)-1] != ch {
		return nil, fmt.Errorf(
			"expected string to end with %c got=%c",
			ch, b[len(b)-1],
		)
	}
	return b[1 : len(b)-1], nil
}
