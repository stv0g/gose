package utils

import (
	"os"
	"strings"
)

// EnvToMap returns the current set of environment variables as a map
func EnvToMap() (map[string]string, error) {
	envMap := make(map[string]string)
	var err error

	for _, v := range os.Environ() {
		splitV := strings.SplitN(v, "=", 2)
		envMap[splitV[0]] = splitV[1]
	}

	return envMap, err
}
