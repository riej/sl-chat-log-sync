package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// SecondLifeClient is SecondLife client.
type SecondLifeClient string

// SecondLifeClients are application names of SecondLife clients.
// Usually can be taken from gDirUtilp->initAppDirs function call into indra/newview/llappviewer.cpp.
var SecondLifeClients = []SecondLifeClient{"SecondLife", "Kokua", "Firestorm", "Firestorm_x64"}

// GetAccountNames retrieves account names inside of the client settings directory.
func (a SecondLifeClient) GetAccountNames() ([]string, error) {
	directory, err := a.GetDirectory()
	if err != nil {
		return nil, err
	}

	exists, err := IsDirectoryExists(directory)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	matches, err := filepath.Glob(filepath.Join(directory, "*", "settings_per_account.xml"))
	if err != nil {
		return nil, fmt.Errorf("unable to read directory %s: %w", directory, err)
	}

	accountNames := make([]string, len(matches))
	for i, match := range matches {
		accountNames[i] = filepath.Base(filepath.Dir(match))
	}

	return accountNames, nil
}

// GetAccountChatLogsDirectory returns SecondLife chat logs directory for specified account.
// In SecondLife client you can specify own chat logs directory for each account.
// The function is written very dumb, because I don't want to parse whole LLSD file.
func (a SecondLifeClient) GetAccountChatLogsDirectory(accountName string) (string, error) {
	directory, err := a.GetDirectory()
	if err != nil {
		return "", err
	}

	settingsFilePath := filepath.Join(directory, accountName, "settings_per_account.xml")
	data, err := os.ReadFile(settingsFilePath)

	// Do not return error if file does not exists.
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("unable to read file %s: %w", settingsFilePath, err)
	}

	llsd := string(data)
	re := regexp.MustCompile(`<key>InstantMessageLogPath</key>\s*<map>([[:graph:]\s]*?)<key>Value</key>\s*<string>([^<]*)</string>`)
	matches := re.FindStringSubmatch(llsd)

	var logsDirectory string
	if len(matches) == 2 {
		logsDirectory = matches[1]
	} else {
		logsDirectory = filepath.Join(directory, accountName)
	}

	return logsDirectory, nil
}

// ListChatLogFileNames returns list of chat log file names for specified account.
// Returns absolute paths for the chat log files.
func (a SecondLifeClient) ListChatLogFileNames(accountName string) (absolutePaths []string, relativePaths []string, err error) {
	logsDirectory, err := a.GetAccountChatLogsDirectory(accountName)
	if err != nil {
		return nil, nil, err
	}

	if logsDirectory == "" {
		return nil, nil, nil
	}

	matches, err := filepath.Glob(filepath.Join(logsDirectory, "*.txt"))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read directory %s: %w", logsDirectory, err)
	}

	for _, match := range matches {
		fileName := filepath.Base(match)

		if !IsReservedFileName(fileName) {
			absolutePaths = append(absolutePaths, match)
			relativePaths = append(relativePaths, fileName)
		}
	}

	return
}

// ReadChatLog read chat log file for specified account.
// fileName must be relatiive to the chat logs directory.
func (a SecondLifeClient) ReadChatLog(accountName string, fileName string) (Messages, error) {
	logsDirectory, err := a.GetAccountChatLogsDirectory(accountName)
	if err != nil {
		return nil, err
	}

	if logsDirectory == "" {
		return nil, nil
	}

	logFilePath := filepath.Join(logsDirectory, fileName)
	f, err := os.Open(logFilePath)

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

// WriteChatLog writes chat log messages into temp file and replaces existing chat logs file with the new one.
func (a SecondLifeClient) WriteChatLog(accountName string, fileName string, messages Messages) error {
	logsDirectory, err := a.GetAccountChatLogsDirectory(accountName)
	if err != nil {
		return err
	}

	// If there's no account directory for current client, create it and save logs here.
	if logsDirectory == "" {
		directory, err := a.GetDirectory()
		if err != nil {
			return err
		}

		logsDirectory = filepath.Join(directory, accountName)

		exists, err := IsDirectoryExists(logsDirectory)
		if err != nil {
			return err
		}
		if !exists {
			err = os.Mkdir(logsDirectory, 0755)
			if err != nil {
				return fmt.Errorf("unable to create directory %s: %w", logsDirectory, err)
			}
		}
	}

	logFilePath := filepath.Join(logsDirectory, fileName)

	wf, err := os.Create(filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s", accountName, fileName)))
	if err != nil {
		return fmt.Errorf("error creating temp file for chat log %s/%s: %w", accountName, fileName, err)
	}

	writtenFileName := wf.Name()

	err = messages.Write(wf)
	if err != nil {
		_ = wf.Close()
		_ = os.Remove(writtenFileName)
		return fmt.Errorf("error writing temp file %s for chat log %s/%s: %w", writtenFileName, accountName, fileName, err)
	}

	err = wf.Close()
	if err != nil {
		_ = os.Remove(writtenFileName)
		return fmt.Errorf("error closing temp file %s for chat log %s/%s: %w", writtenFileName, accountName, fileName, err)
	}

	err = MoveFile(writtenFileName, logFilePath)
	if err != nil {
		_ = os.Remove(writtenFileName)
		return fmt.Errorf("error moving temp file %s into %s for chat log %s/%s: %w", writtenFileName, logFilePath, accountName, fileName, err)
	}

	return nil
}
