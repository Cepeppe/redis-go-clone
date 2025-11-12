package main

import (
	"errors"
	"strings"
)

type Handler func(args string) (string, error)

var cmdHandlers = map[string]Handler{
	"GET":  GET,
	"SET":  SET,
	"DEL":  DEL,
	"ESC":  ESC,
	"PING": PING,
	"HELP": HELP,
}

func getConstantCommandsArray() []string {
	dict := cmdHandlers
	ks := make([]string, 0, len(dict))
	for k := range dict {
		ks = append(ks, k)
	}
	return ks
}

// Try parse and execute command, returns: result_str, err
func tryParseExecuteCommand(command_raw string) (string, error) {

	cmd, args, ok := CutFirstToken(command_raw)
	if !ok {
		return "NOT_OK", errors.New("command parsing error")
	}

	return executeCommand(cmd, args)
}

// Returns execution result (string) and error (=nil if no error)
func executeCommand(cmd string, args string) (string, error) {
	handler, ok := cmdHandlers[cmd]
	if !ok || handler == nil {
		return "NOT_OK", errors.New("unknown command: " + cmd)
	}
	return handler(args)
}

func GET(args string) (string, error) {
	return "", nil
}

func SET(args string) (string, error) {
	return "", nil
}

func DEL(args string) (string, error) {
	return "", nil
}

func ESC(args string) (string, error) {
	return "", nil
}

func PING(args string) (string, error) {
	return "PONG", nil
}

func HELP(args string) (string, error) {
	return "", nil
}

// CutFirstToken returns (token, rest, ok).
// Separators are space or tab. Leading separators are skipped.
// On failure (empty or only separators) returns ("", s, false).
// separatori: spazio, tab, CR, LF
func CutFirstToken(s string) (string, string, bool) {
	i, n := 0, len(s)
	isSep := func(b byte) bool { return b == ' ' || b == '\t' || b == '\r' || b == '\n' }
	for i < n && isSep(s[i]) {
		i++
	}
	if i == n {
		return "", s, false
	}
	j := i
	for j < n && !isSep(s[j]) {
		j++
	}
	tok := s[i:j]
	k := j
	for k < n && isSep(s[k]) {
		k++
	}
	return tok, s[k:], true
}

func canonCmd(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.ToUpper(strings.TrimSpace(s))
}
