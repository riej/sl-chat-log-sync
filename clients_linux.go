package main

import (
	"fmt"
	"os"
	"strings"
)

// GetDirectory returns path for SL client settings directory for current OS.
// This function is based exactly on SL source code, with all the same possible caveats that it has.
// Directory detection took from indra/llvfs/lldir_linux.cpp
func (a ClientAppName) GetDirectory() (string, error) {
	envParam := os.Getenv(fmt.Sprintf("%s_USER_DIR", strings.ToUpper(string(a))))
	if envParam != "" {
		return envParam, nil
	}

	envParam = os.Getenv("HOME")
	return fmt.Sprintf("%s/.%s", envParam, strings.ToLower(string(a))), nil
}
