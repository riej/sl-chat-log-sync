package main

import (
	"fmt"
)

// GetDirectory returns path for SL client settings directory for current OS.
// This function is based exactly on SL source code, with all the same possible caveats that it has.
// I tried to take directory detection from indra/llvfs/lldir_mac.cpp...
// But unfortunately, I don't know how to call NSSearchPathForDirectoriesInDomains from Go,
// so I made it in very dumb way instead.
func (a ClientAppName) GetDirectory() (string, error) {
	return fmt.Sprintf("%s/Library/Application Support/%s", os.Getenv("HOME"), string(a)), nil
}
