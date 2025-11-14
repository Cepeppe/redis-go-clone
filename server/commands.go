package main

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

type Handler func(args string) (string, error)

var cmdHandlers = map[string]Handler{
	"GET":    GET,
	"SET":    SET,
	"DEL":    DEL,
	"SETEXP": SETEXP,
	"ESC":    ESC,
	"PING":   PING,
	"HELP":   HELP,
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

	cmd, args, err := cutFirstTokenSpaceTab(command_raw)
	if err != nil {
		return "NOT_OK", errors.New("command parsing error: " + err.Error())
	}

	return executeCommand(cmd, args)
}

// Returns execution result (string) and error (=nil if no error)
func executeCommand(cmd string, args string) (string, error) {

	handler, ok := cmdHandlers[strings.ToUpper(cmd)]
	if !ok || handler == nil {
		return "NOT_OK", errors.New("unknown command: " + cmd)
	}
	return handler(args)
}

func GET(args string) (string, error) {

	key, _, err := cutFirstTokenSpaceTab(args)
	if err != nil {
		return "NOT_OK", errors.New("command parsing error: " + err.Error())
	}

	value, exists := keyDataSpace.Get(key)
	if !exists {
		return "NOT_OK", errors.New("No such KEY is present: " + key)
	}

	return value, nil
}

func SET(args string) (string, error) {

	var err error
	key, args, err := cutFirstTokenSpaceTab(args)
	if err != nil {
		return "NOT_OK", errors.New("command parsing error: " + err.Error())
	}

	data, args, err := cutFirstTokenSmart(args)
	if err != nil {
		return "NOT_OK", errors.New("command parsing error: " + err.Error())
	}

	var expiration_sec int64 = -1
	if args != "" {
		exp, _, err := cutFirstTokenSpaceTab(args)
		if err != nil {
			return "NOT_OK", errors.New("command parsing error: " + err.Error())
		}

		expiration_sec, err = strconv.ParseInt(exp, 10, 64)
		if err != nil {
			return "NOT_OK", errors.New("command parsing error: " + err.Error())
		}
	}

	var expire_at_ts int64 = math.MaxInt64
	if expiration_sec == -1 {
		expire_at_ts = math.MaxInt64
	} else {
		expire_at_ts = time.Now().UnixMilli() + expiration_sec*1000
	}

	keyDataSpace.Add(key, data)
	keyExpirations.PushItem(KeyExpiration{key: key, expire_timestamp: expire_at_ts})
	//TODO: WRITE ON AOF

	return "", nil
}

func DEL(args string) (string, error) {
	key, _, err := cutFirstTokenSpaceTab(args)
	if err != nil {
		return "NOT_OK", errors.New("command parsing error: " + err.Error())
	}
	keyDataSpace.Remove(key)
	keyExpirations.Remove(key)
	//TODO: WRITE ON AOF

	return "OK", nil

}

func SETEXP(args string) (string, error) {
	var expiration_sec int64

	key, remaining, err := cutFirstTokenSpaceTab(args)
	if err != nil {
		return "NOT_OK", errors.New("command parsing error: " + err.Error())
	}

	if remaining != "" {
		exp, _, err := cutFirstTokenSpaceTab(remaining)
		if err != nil {
			return "NOT_OK", errors.New("command parsing error: " + err.Error())
		}

		expiration_sec, err = strconv.ParseInt(exp, 10, 64)
		if err != nil {
			return "NOT_OK", errors.New("command parsing error: " + err.Error())
		}
	}

	expire_at_ts := time.Now().UnixMilli() + expiration_sec*1000
	exists := keyExpirations.UpdateExpiration(key, expire_at_ts)

	if !exists {
		return "NOT_OK", errors.New("you tried to update expiration for a non existing key")
	}

	return "OK", nil
}

func ESC(args string) (string, error) {
	return "", nil
}

func PING(args string) (string, error) {
	return "PONG", nil
}

func HELP(args string) (string, error) {
	return "cant help ya rn", nil
}

func canonCmd(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.ToUpper(strings.TrimSpace(s))
}
