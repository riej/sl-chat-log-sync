package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// IsDirectoryExists returns true if the client settings directory exists and it's really directory.
// Returns false and nil, if directory is not exists.
// Returns false and error, if it's not a directory or there's another error.
func IsDirectoryExists(directory string) (bool, error) {
	info, err := os.Stat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("unable to open directory %s: %w", directory, err)
	}
	if info.IsDir() {
		return true, nil
	}

	return false, fmt.Errorf("unable to open directory %s: not a directory", directory)
}

// IsReservedFileName returns true if file name is reserved and contains other information than chat logs.
func IsReservedFileName(fileName string) bool {
	return fileName == "search_history.txt" ||
		fileName == "teleport_history.txt" ||
		fileName == "typed_locations.txt" ||
		fileName == "plugin_cookies.txt"
}

// Unique returns unique items from the slice.
// It doesn't preserve items order.
func Unique[K comparable](items []K) (result []K) {
	m := make(map[K]interface{})
	for _, item := range items {
		m[item] = nil
	}

	for item := range m {
		result = append(result, item)
	}

	return
}

// UniqueBasenames returns unique base names for file paths.
func UniqueBasenames(fileNames []string) []string {
	basenames := make([]string, len(fileNames))
	for i, fileName := range fileNames {
		basenames[i] = filepath.Base(fileName)
	}

	return Unique(basenames)
}

// Contains returns true if items contains into items slice.
func Contains(items []string, item string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}

	return false
}

// MoveFile moves file from sourcePath to destPath.
// Took from https://stackoverflow.com/a/50741908
func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
