package common

import (
	"os"
	"strconv"
)

func getInt(name string, defaultVal int) int {
	envName := ENV_PREFIX + name
	val := os.Getenv(envName)
	if val == "" {
		return defaultVal
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func getVar(name string, defaultVal string) string {
	envName := ENV_PREFIX + name
	val := os.Getenv(envName)
	if val == "" {
		return defaultVal
	}
	return val
}

var (
	DbHost = getVar("DB_HOST", "127.0.0.1")
	DbPort = getVar("DB_PORT", "3306")
	DbName = getVar("DB_NAME", "gluttoy")
	DbUser = getVar("DB_USER", "root")
	DbPass = getVar("DB_PASS", "root")

	GeneralLog = getVar("GENERAL_LOG", "general.log")
	InfoLog    = getVar("INFO_LOG", "info.log")
	ErrorLog   = getVar("ERROR_LOG", "error.log")

	ListenInterval = getInt("LISTEN_INTERVAL", 5)
	VideoDir       = getVar("VIDEO_DIR", "")
)
