package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// GetDirectory returns path for SL client settings directory for current OS.
// This function is based exactly on SL source code, with all the same possible caveats that it has.
func (a SecondLifeClient) GetDirectory() (string, error) {
	// Directory detection took from indra/llvfs/lldir_win32.cpp
	envParam := os.Getenv("APPDATA")
	if envParam != "" {
		return fmt.Sprintf("%s\\%s", envParam, a), nil
	}

	knownPath, err := windows.KnownFolderPath(windows.FOLDERID_RoamingAppData, 0)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve application data path: %w", err)
	}

	return fmt.Sprintf("%s\\%s", knownPath, a), nil
}
