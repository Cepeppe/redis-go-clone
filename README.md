# Redis Go Clone

This repository contains a simple command-line server and client implementation simulating core Redis commands.

## Run

### Windows

```bash
go run .\server\
go run .\client\
```

### Linux / macOS

```bash
go run ./server/
go run ./client/
```

## Command Guide (HELP)

The following commands are available when connected to the server. The output is formatted like this:

```text
---------------------------------------------
        Redis Go Clone - Command Guide
---------------------------------------------

GET <key>
    Retrieves the value associated with <key>.
    Example: GET user:123

SET <key> <value> [expire_after]
    Sets the <value> for the <key>.
    [expire_after] (optional): Expiration time in seconds. If not set, the key has no expiration.
    Example 1 (No Expiration): SET username "Mario Rossi"
    Example 2 (With Expiration): SET session_token "abc" 3600

DEL <key>
    Deletes the specified <key>.
    Example: DEL temp_data

SETEXP <key> <expire_after>
    Sets or updates the expiration time of <key> to the new 
    <expire_after> value (in seconds, from the current instant).
    Example: SETEXP token 600

PING
    Checks the connection. Returns "PONG".

HELP
    Displays this help message.

ESC
    Closes the connection and exits the client.

---------------------------------------------
```
