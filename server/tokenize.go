package main

import (
	"errors"
)

var (
	// ErrNoToken is returned when no token is found (only leading spaces/tabs).
	ErrNoToken = errors.New("missing token")
	// ErrMalformed is returned when the input looks like a token but is incomplete
	// (e.g., unclosed string or unbalanced braces/brackets).
	ErrMalformed = errors.New("malformed token")
)

// cutFirstTokenSpaceTab splits s into first token and remainder using space or tab as separators.
// Leading separators are skipped; trailing separators after the first token are consumed.
// Returns (token, rest, nil) on success; ("", s, ErrNoToken) if no token is found.
// Behavior is identical to the original version, except it returns an error instead of a boolean.
func cutFirstTokenSpaceTab(s string) (string, string, error) {
	i, n := 0, len(s)
	for i < n && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	if i == n {
		return "", s, ErrNoToken
	}
	j := i
	for j < n && s[j] != ' ' && s[j] != '\t' {
		j++
	}
	k := j
	for k < n && (s[k] == ' ' || s[k] == '\t') {
		k++
	}
	return s[i:j], s[k:], nil
}

// cutFirstTokenSmart extracts the first "data-aware" token from s and returns (token, rest, err).
// Token types supported:
//  1. Bare token: bytes until the next space or tab.
//  2. Quoted string: starts with '"' and ends at the matching non-escaped '"'.
//  3. JSON-like block: starts with '{' or '[' and ends when braces/brackets are balanced,
//     with strings inside handled correctly (quotes and escapes do not affect nesting).
//
// Separators are space ' ' and tab '\t'. Leading separators are skipped.
// After extracting a token, any trailing separators are consumed, and the remainder is returned.
// Complexity: O(n) over the examined bytes. The function does not validate JSON content; it only
// delimits tokens structurally.
func cutFirstTokenSmart(s string) (string, string, error) {
	n := len(s)
	i := 0

	// Skip leading separators (space or tab).
	for i < n && isSpaceTab(s[i]) {
		i++
	}
	if i >= n {
		return "", s, ErrNoToken
	}

	switch s[i] {
	case '"':
		// Quoted string: scan until a non-escaped closing '"'.
		j := i + 1
		escaped := false
		for j < n {
			c := s[j]
			if escaped {
				// Current byte is escaped; consume and reset escape state.
				escaped = false
				j++
				continue
			}
			if c == '\\' {
				escaped = true
				j++
				continue
			}
			if c == '"' {
				// Closing quote found; include it and finish.
				j++
				// Consume trailing separators for rest.
				k := j
				for k < n && isSpaceTab(s[k]) {
					k++
				}
				return s[i:j], s[k:], nil
			}
			j++
		}
		// Reached end without closing quote.
		return "", s, ErrMalformed

	case '{', '[':
		// JSON-like block with nesting and in-string tracking.
		j := i
		braceDepth, bracketDepth := 0, 0
		inString := false
		escaped := false

		for j < n {
			c := s[j]

			if inString {
				// Inside string: only quote and backslash matter.
				if escaped {
					escaped = false
				} else {
					if c == '\\' {
						escaped = true
					} else if c == '"' {
						inString = false
					}
				}
				j++
				continue
			}

			// Not inside string: update depths and possibly enter string mode.
			switch c {
			case '"':
				inString = true
			case '{':
				braceDepth++
			case '}':
				braceDepth--
				if braceDepth < 0 {
					return "", s, ErrMalformed
				}
			case '[':
				bracketDepth++
			case ']':
				bracketDepth--
				if bracketDepth < 0 {
					return "", s, ErrMalformed
				}
			}

			j++

			// Token ends when both depths return to zero.
			if braceDepth == 0 && bracketDepth == 0 {
				// Consume trailing separators for rest.
				k := j
				for k < n && isSpaceTab(s[k]) {
					k++
				}
				return s[i:j], s[k:], nil
			}
		}
		// Reached end with non-zero depths or inside string.
		return "", s, ErrMalformed

	default:
		// Bare token: read until next space/tab or end.
		j := i
		for j < n && !isSpaceTab(s[j]) {
			j++
		}
		// Consume trailing separators for rest.
		k := j
		for k < n && isSpaceTab(s[k]) {
			k++
		}
		return s[i:j], s[k:], nil
	}
}

func isSpaceTab(b byte) bool {
	return b == ' ' || b == '\t'
}
