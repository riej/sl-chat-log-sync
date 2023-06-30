package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/cheggaaa/pb/v3"
)

var (
	ArchiveOnly     = flag.Bool("archive-only", false, "don't replace existing chat log files, archive only")
	ArchiveFileName = flag.String("archive", "sl_chat_logs.zip", "Archive file name")
)

func main() {
	flag.Parse()

	var inputStorages []ChatLogsStorage

	// Check each SecondLife client.
	for _, clientApp := range SecondLifeClients {
		directory, err := clientApp.GetDirectory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s detection error, skipping it: %s\n", clientApp, err)
			continue
		}

		exists, err := IsDirectoryExists(directory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s detection error, skipping it: %s\n", clientApp, err)
			continue
		}

		if exists {
			fmt.Printf("%s found\n", clientApp)

			clientApp := clientApp
			inputStorages = append(inputStorages, &clientApp)
		}
	}

	if len(inputStorages) == 0 {
		fmt.Printf("No SecondLife clients found.\n")
		os.Exit(0)
		return
	}

	// Open archives.
	archive, err := ReadChatLogsArchive(*ArchiveFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s error, skipping it: %s\n", *ArchiveFileName, err)
		os.Exit(1)
		return
	}

	inputStorages = append(inputStorages, archive)
	defer func() {
		err := archive.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			os.Exit(1)
			return
		}
	}()

	var outputStorages []ChatLogsStorage
	if *ArchiveOnly {
		outputStorages = []ChatLogsStorage{archive}
	} else {
		outputStorages = inputStorages
	}

	// Retrieve all account names.
	var accountNames []string

	for _, storage := range inputStorages {
		storageAccountNames, err := storage.GetAccountNames()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			os.Exit(1)
			return
		}

		accountNames = append(accountNames, storageAccountNames...)
	}

	accountNames = Unique(accountNames)
	sort.Strings(accountNames)

	if len(accountNames) == 0 {
		fmt.Printf("No SecondLife accounts found.\n")
		os.Exit(0)
		return
	}

	fmt.Printf("Accounts found:\n")
	for _, accountName := range accountNames {
		fmt.Printf(" - %s\n", accountName)
	}

	// Read all chat logs and merge them.
	for _, accountName := range accountNames {
		fmt.Printf("Merging %s chat logs...\n", accountName)

		var chatLogsFileNames []string
		for _, storage := range inputStorages {
			_, fileNames, err := storage.ListChatLogFileNames(accountName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
				os.Exit(1)
				return
			}

			chatLogsFileNames = append(chatLogsFileNames, fileNames...)
		}

		chatLogsFileNames = Unique(chatLogsFileNames)
		sort.Strings(chatLogsFileNames)

		bar := pb.StartNew(len(chatLogsFileNames))

		for _, fileName := range chatLogsFileNames {
			var chatLogs []Messages

			for _, storage := range inputStorages {
				messages, err := storage.ReadChatLog(accountName, fileName)
				if err != nil {
					bar.Finish()
					fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
					os.Exit(1)
					return
				}

				chatLogs = append(chatLogs, messages)
			}

			merged := Merge(chatLogs...)

			for _, storage := range outputStorages {
				err := storage.WriteChatLog(accountName, fileName, merged)
				if err != nil {
					bar.Finish()
					fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
					os.Exit(1)
					return
				}
			}

			bar.Increment()
		}

		bar.Finish()
	}
}
