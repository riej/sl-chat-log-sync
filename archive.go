package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ChatLogsArchive is .zip archive containing chat logs.
// Structure: <account_name>/<chat_logs.txt>
type ChatLogsArchive struct {
	fileName string
	r        *zip.ReadCloser
	wf       *os.File
	w        *zip.Writer
}

// ReadChatLogsArchive opens chat logs archive.
func ReadChatLogsArchive(fileName string) (*ChatLogsArchive, error) {
	r, err := zip.OpenReader(fileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	wf, err := os.Create(filepath.Join(os.TempDir(), fileName))
	if err != nil {
		if r != nil {
			r.Close()
		}
		return nil, err
	}

	fmt.Println(wf.Name())

	return &ChatLogsArchive{
		fileName: fileName,
		r:        r,
		wf:       wf,
		w:        zip.NewWriter(wf),
	}, nil
}

// Close closes internal .zip reader, writer and replaces old .zip file with the new one.
func (a *ChatLogsArchive) Close() error {
	if a.r != nil {
		_ = a.r.Close()
	}

	writtenFileName := a.wf.Name()

	err := a.w.Close()
	if err != nil {
		a.wf.Close()
		_ = os.Remove(writtenFileName)
		return fmt.Errorf("error closing .zip file %s: %w", a.wf.Name(), err)
	}

	err = a.wf.Close()
	if err != nil {
		_ = os.Remove(writtenFileName)
		return fmt.Errorf("error closing file %s: %w", writtenFileName, err)
	}

	err = MoveFile(writtenFileName, a.fileName)
	if err != nil {
		return fmt.Errorf("error overwriting file %s with %s: %w", a.fileName, writtenFileName, err)
	}

	return nil
}

// GetAccountNames extracts account names from the archive.
func (a *ChatLogsArchive) GetAccountNames() ([]string, error) {
	if a.r == nil {
		return nil, nil
	}

	var accountNamesMap = make(map[string]interface{})
	for _, f := range a.r.File {
		// Take only .txt files.
		if filepath.Ext(f.Name) != ".txt" {
			continue
		}

		path := filepath.Dir(f.Name)

		// Take files from 1st level directories only.
		if path != "" && !strings.ContainsAny(path, "/\\") {
			accountNamesMap[path] = nil
		}
	}

	var accountNames []string
	for accountName := range accountNamesMap {
		accountNames = append(accountNames, accountName)
	}

	return accountNames, nil
}

// ListChatLogFileNames returns list of chat log files for the specified account name inside of archive.
func (a *ChatLogsArchive) ListChatLogFileNames(accountName string) (absolutePaths []string, relativePaths []string, err error) {
	if a.r == nil {
		return nil, nil, nil
	}

	var logFileNames []string
	for _, f := range a.r.File {
		// Take only .txt files.
		if filepath.Ext(f.Name) != ".txt" {
			continue
		}

		path := filepath.Dir(f.Name)
		if path != accountName {
			continue
		}

		fileName := filepath.Base(f.Name)
		if !IsReservedFileName(fileName) {
			absolutePaths = append(absolutePaths, f.Name)
			relativePaths = append(logFileNames, fileName)
		}
	}

	return
}

// ReadChatLog read chat log file for specified account.
// fileName must be relatiive to the chat logs directory.
func (a *ChatLogsArchive) ReadChatLog(accountName string, fileName string) (Messages, error) {
	if a.r == nil {
		return nil, nil
	}

	logFilePath := strings.Join([]string{accountName, fileName}, "/")
	f, err := a.r.Open(logFilePath)

	// Do not return error if file doesn't exists.
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to open chat log %s: %w", logFilePath, err)
	}
	defer f.Close()

	messages, err := ReadMessages(f)
	if err != nil {
		return messages, fmt.Errorf("unable to read chat log %s: %w", logFilePath, err)
	}

	return messages, nil
}

// WriteChatLog writes chat log messages into new archive.
func (a *ChatLogsArchive) WriteChatLog(accountName string, fileName string, messages Messages) error {
	logFilePath := strings.Join([]string{accountName, fileName}, "/")

	f, err := a.w.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", logFilePath, err)
	}

	err = messages.Write(f)
	if err != nil {
		return fmt.Errorf("error writing file %s: %w", logFilePath, err)
	}

	return nil
}
