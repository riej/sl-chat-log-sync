package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/cheggaaa/pb/v3"
	"github.com/logrusorgru/aurora"
)

func DisplayErrorAndExit(err error) {
	fmt.Println(aurora.Red("Error:"), err)

	switch runtime.GOOS {
	case "windows":
		exec.Command("pause").Run()
		break
	default:
	}
	os.Exit(1)
}

func UniqueLines(lines []string) []string {
	mLines := make(map[string]bool)
	result := make([]string, 0)

	for _, line := range lines {
		if _, ok := mLines[line]; !ok {
			mLines[line] = true
			result = append(result, line)
		}
	}

	return result
}

func main() {
	slDataPath, err := GetSLDataPath()
	if err != nil {
		DisplayErrorAndExit(err)
	}

	backupPath, err := GetBackupPath()
	if err != nil {
		DisplayErrorAndExit(err)
	}

	fileNames1, err := filepath.Glob(filepath.Join(slDataPath, "*", "*.txt"))
	if err != nil {
		DisplayErrorAndExit(err)
	}

	fileNames2, err := filepath.Glob(filepath.Join(backupPath, "*", "*.txt"))
	if err != nil {
		DisplayErrorAndExit(err)
	}

	fileNames1 = RemovePathPrefix(slDataPath, fileNames1)
	fileNames2 = RemovePathPrefix(backupPath, fileNames2)

	fileNames := UniqueLines(append(fileNames1, fileNames2...))
	bar := pb.StartNew(len(fileNames))

	worker := func(filenames <-chan string, wg *sync.WaitGroup) {
		for fileName := range filenames {
			file1, err := ReadChatLog(filepath.Join(slDataPath, fileName))
			if err != nil {
				DisplayErrorAndExit(err)
			}

			file2, err := ReadChatLog(filepath.Join(backupPath, fileName))
			if err != nil {
				DisplayErrorAndExit(err)
			}

			merged := append(file1, file2...)
			sort.Stable(merged)
			merged = merged.Unique()

			if err := merged.WriteFile(filepath.Join(slDataPath, fileName)); err != nil {
				DisplayErrorAndExit(err)
			}

			if err := merged.WriteFile(filepath.Join(backupPath, fileName)); err != nil {
				DisplayErrorAndExit(err)
			}

			bar.Increment()
		}

		wg.Done()
	}

	jobs := make(chan string, len(fileNames))
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	for _, fileName := range fileNames {
		jobs <- fileName
	}
	close(jobs)

	wg.Wait()
	bar.Finish()
}
