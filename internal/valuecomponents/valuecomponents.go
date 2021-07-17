// Package valuecomponents contains a parser for
// field value components as specified in RFC 7230 section 3.2.6
// (https://datatracker.ietf.org/doc/html/rfc7230#section-3.2.6)
package valuecomponents

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Value implements a single component value. A value may have a
// primary element as well as key value pairs. Either may be empty.
type Value struct {
	Primary string
	Pairs   map[string]string
}

// ValueList implements an ordered list of values.
type ValueList []Value

// ParseValueList parses the given string into a value list and returns
// it or an error.
func ParseValueList(s string) (ValueList, error) {
	i := 0

	var vals ValueList

outer:
	for {
		if i >= len(s) {
			break
		}

		val := Value{
			Pairs: make(map[string]string),
		}

		for {
			i += consumeWhitespace(s[i:])
			if i >= len(s) {
				break outer
			}

			v, l, err := tokenOrQuotedString(s[i:])
			if err != nil {
				return nil, err
			}

			i += l
			i += consumeWhitespace(s[i:])

			c, l := utf8.DecodeRuneInString(s[i:])
			if c != '=' {
				val.Primary = v
				i += l
			} else if c == '=' {
				key := v

				i += l
				i += consumeWhitespace(s[i:])

				v, l, err = tokenOrQuotedString(s[i:])
				if err != nil {
					return nil, err
				}

				val.Pairs[key] = v
				i += l
				i += consumeWhitespace(s[i:])

				c, l = utf8.DecodeRuneInString(s[i:])
			}

			if c == ',' || i >= len(s) {
				// Parse next value
				i += l
				break
			}

			if c == ';' {
				i += l
				// Parse trailing pairs
				continue
			}

			return nil, fmt.Errorf("unexpected delimiter: '%c' at '%s'", c, s[i:])
		}

		vals = append(vals, val)
	}

	return vals, nil
}

func consumeWhitespace(s string) int {
	i := 0
	for {
		if i >= len(s) {
			return i
		}

		c, s := utf8.DecodeRuneInString(s[i:])
		if !unicode.IsSpace(c) {
			break
		}

		i += s
	}
	return i
}

func tokenOrQuotedString(s string) (string, int, error) {
	c, _ := utf8.DecodeRuneInString(s)
	if c == '"' {
		r, err := ParseQuotedString(s)
		return r, len(r) + 2, err
	}

	t := ParseToken(s)
	return t, len(t), nil
}

const (
	Delimiters = `"(),/:;<=>?@[\]{}`
	TokenChars = "!#$%&'*+-.^_`|~0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// ParseToken parses a single token and returns it.
func ParseToken(s string) string {
	i := 0
	for {
		if i >= len(s) {
			return s
		}

		c, l := utf8.DecodeRuneInString(s[i:])
		if !isTokenChar(c) {
			return s[:i]
		}

		i += l
	}
}

func isTokenChar(r rune) bool {
	return strings.ContainsRune(TokenChars, r)
}

func ParseQuotedString(v string) (string, error) {
	c, s := utf8.DecodeRuneInString(v)
	if c != '"' {
		return "", nil
	}

	i := s

	for {
		if i >= len(v) {
			return "", fmt.Errorf("not a quoted string: '%s'", v)
		}

		c, s := utf8.DecodeRuneInString(v[i:])
		if c == '"' {
			return v[1:i], nil
		}

		// TODO: Handle backslash

		i += s
	}
}
