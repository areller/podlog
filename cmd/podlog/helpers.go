package main

import (
	"strings"
	"runtime"
	"os"
)

func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
        if home == "" {
            home = os.Getenv("USERPROFILE")
        }
        return home
	}
	return os.Getenv("HOME")
}

func GetEnvVarPath(key string) string {
	return strings.Replace(os.Getenv(key), "~", GetHomeDir(), -1)
}