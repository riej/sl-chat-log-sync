package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ChatLinesChunk struct {
	Timestamp int64
	ChatLines []string
}

func (chunk ChatLinesChunk) Text() string {
	return strings.Join(chunk.ChatLines, "")
}

type ChatLog []*ChatLinesChunk

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

	chunk := &ChatLinesChunk{
		Timestamp: 0,
		ChatLines: make([]string, 0),
	}

	reader := bufio.NewReader(file)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if s[0] == '[' && s[17] == ']' {
			t, err := time.Parse("2006/01/02 15:04", s[1:17])
			if err == nil && t.Unix() != chunk.Timestamp {
				log = append(log, chunk)
				chunk = &ChatLinesChunk{
					Timestamp: t.Unix(),
					ChatLines: make([]string, 0),
				}
			}
		}

		chunk.ChatLines = append(chunk.ChatLines, s)
	}

	log = append(log, chunk)

	return log, nil
}

func (log ChatLog) Unique() (result ChatLog) {
	result = make(ChatLog, 0)

	timestamps := make([]int64, 0)
	chunks := make(map[int64]*ChatLinesChunk)
	for _, chunk := range log {
		if existing, ok := chunks[chunk.Timestamp]; ok {
			if len(chunk.Text()) > len(existing.Text()) {
				chunks[chunk.Timestamp] = chunk
			}
		} else {
			timestamps = append(timestamps, chunk.Timestamp)
			chunks[chunk.Timestamp] = chunk
		}
	}

	for _, timestamp := range timestamps {
		if chunk, ok := chunks[timestamp]; ok {
			result = append(result, chunk)
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
	for _, chunk := range log {
		if _, err = w.WriteString(chunk.Text()); err != nil {
			return err
		}
	}
	w.Flush()

	return nil
}
