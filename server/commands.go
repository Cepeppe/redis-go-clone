package main

func getConstantCommands() map[string][]string {
	return map[string][]string{
		"GET":  []string{"GET", "Get", "get"},
		"SET":  []string{"SET", "Set", "set"},
		"DEL":  []string{"DEL", "Del", "del"},
		"ESC":  []string{"ESC", "Esc", "esc"},
		"PING": []string{"PING", "Ping", "ping"},
	}
}

func GET() string {
	return ""
}

func PUT() string {
	return ""
}

func DEL() string {
	return ""
}

func ESC() string {
	return ""
}

func PING() string {
	return "PONG"
}
