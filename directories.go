package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Copy of os.UserHomeDir()
// Just old versions of Go hasn't this function
func UserHomeDir() (string, error) {
	env, enverr := "HOME", "$HOME"
	switch runtime.GOOS {
	case "windows":
		env, enverr = "USERPROFILE", "%userprofile%"
	case "plan9":
		env, enverr = "home", "$home"
	}
	if v := os.Getenv(env); v != "" {
		return v, nil
	}
	// On some geese the home directory is not always defined.
	switch runtime.GOOS {
	case "nacl":
		return "/", nil
	case "android":
		return "/sdcard", nil
	case "darwin":
		if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
			return "/", nil
		}
	}
	return "", errors.New(enverr + " is not defined")
}

func GetSLDataPath() (string, error) {
	homeDir, err := UserHomeDir()
	if err != nil {
		return "", err
	}

	var slDir string

	switch runtime.GOOS {
	case "windows":
		slDir = filepath.Join(homeDir, "AppData", "Roaming", "SecondLife")
		break

	case "linux":
		slDir = filepath.Join(homeDir, ".secondlife")
		break

	case "darwin":
		slDir = filepath.Join(homeDir, "Library", "Application Support", "SecondLife")
		break

	default:
		return "", errors.New("Unknown OS")
	}

	slDir = filepath.ToSlash(slDir)

	err = os.MkdirAll(slDir, 0664)
	if err != nil {
		return "", err
	}

	return slDir, nil
}

func GetBackupPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	wd = filepath.Join(wd, "SecondLife_backup")

	wd = filepath.ToSlash(wd)

	err = os.MkdirAll(wd, 0664)
	if err != nil {
		return "", err
	}

	return wd, nil
}

func RemovePathPrefix(path string, filenames []string) (result []string) {
	result = make([]string, len(filenames))
	for i, fname := range filenames {
		result[i] = strings.TrimPrefix(filepath.ToSlash(fname), path)
	}
	return result
}
