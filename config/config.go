package config

import (
	"os"
	"strings"
)

var conf map[string]string
var _initialized bool = false

func applyDefaults() {
	conf["port"] = "9999"
	conf["address"] = "127.0.0.1"
	conf["checkin"] = "127.0.0.1:9999" //,127.0.0.1:8888
	conf["public"] = "true"
	conf["maxincoming"] = "5"
}

func initialize() {
	conf = make(map[string]string)
	applyDefaults()
	args := os.Args[1:]
	for _, arg := range args {
		pair := strings.Split(arg, "=")
		if len(pair) == 2 {
			conf[pair[0]] = pair[1]
		}
	}
	_initialized = true
}

//Attr -
func Attr(key string) string {
	if !_initialized {
		initialize()
	}
	return conf[key]
}
