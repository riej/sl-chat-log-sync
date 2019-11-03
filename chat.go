package main

import (
	"bufio"
	"os"
	"path/filepath"
	"time"
)

type ChatLine struct {
	Timestamp int64
	Text      string
}

type ChatLog []ChatLine

func (log ChatLog) Len() int {
	return len(log)
}

func (log ChatLog) Less(a, b int) bool {
	return log[a].Timestamp < log[b].Timestamp
}

func (log ChatLog) Swap(a, b int) {
	log[a], log[b] = log[b], log[a]
}

func ReadChatLog(Filename string) (log ChatLog, err error) {
	file, err := os.Open(Filename)
	if err != nil {
		if os.IsNotExist(err) {
			return ChatLog{}, nil
		}

		return nil, err
	}

	defer file.Close()

	var lastTimestamp int64 = 0

	reader := bufio.NewReader(file)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if s[0] == '[' && s[17] == ']' {
			t, err := time.Parse("2006/01/02 15:04", s[1:17])
			if err == nil {
				lastTimestamp = t.Unix()
			}
		}

		log = append(log, ChatLine{
			Timestamp: lastTimestamp,
			Text:      s,
		})
	}

	return log, nil
}

func (log ChatLog) Unique() (result ChatLog) {
	result = make(ChatLog, 0)

	var lastTimestamp int64 = 0
	var lines map[string]bool = make(map[string]bool)

	for _, line := range log {
		if line.Timestamp != lastTimestamp {
			lines = make(map[string]bool)
			lastTimestamp = line.Timestamp
		}

		if _, ok := lines[line.Text]; !ok {
			lines[line.Text] = true
			result = append(result, line)
		}
	}

	return result
}

func (log ChatLog) WriteFile(fname string) (err error) {
	if err = os.MkdirAll(filepath.Dir(fname), 0664); err != nil {
		return err
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range log {
		if _, err = w.WriteString(line.Text); err != nil {
			return err
		}
	}
	w.Flush()

	return nil
}
