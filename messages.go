package main

import (
	"bufio"
	"io"
	"time"
)

// Message is SecondLife chat log message.
type Message struct {
	// Timestamp is unixtime of the message.
	Timestamp int64
	// Message is complete chat log message, including timestamp and newlines.
	Message string
}

// Messages is slice of clat log messages.
// It implements sort.Interface.
type Messages []*Message

// Len returns count of chat log messages.
func (m Messages) Len() int {
	return len(m)
}

// Less compares two chat log messages' timestamps.
func (m Messages) Less(a, b int) bool {
	return m[a].Timestamp < m[b].Timestamp
}

// Contains returns true if message is already presents.
func (m Messages) Contains(message Message) bool {
	for _, msg := range m {
		if msg.Timestamp == message.Timestamp && msg.Message == message.Message {
			return true
		}
	}

	return false
}

// Swap swaps two messages.
func (m Messages) Swap(a, b int) {
	m[a], m[b] = m[b], m[a]
}

// Write writes chat log messages into the writer.
func (m Messages) Write(w io.Writer) error {
	for _, message := range m {
		_, err := w.Write([]byte(message.Message))
		if err != nil {
			return err
		}
	}

	return nil
}

// Merge merges several chat logs into single.
func Merge(messages ...Messages) (result Messages) {
	var stream = TimedMessagesStream{sources: messages}

	for {
		nextMessages := stream.NextMessages()
		if len(nextMessages) == 0 {
			return
		}

		result = append(result, nextMessages...)
	}
}

// ReadMessages reads chat log messages from the reader.
func ReadMessages(r io.Reader) (messages Messages, err error) {
	var message *Message

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		s := scanner.Text() + "\n"

		if s[0] == '[' && s[17] == ']' {
			t, err := time.Parse("2006/01/02 15:04", s[1:17])
			if err == nil {
				if message != nil {
					messages = append(messages, message)
				}
				message = &Message{
					Timestamp: t.Unix(),
					Message:   s,
				}
				continue
			}
		}

		message.Message += s
	}

	if message != nil {
		messages = append(messages, message)
	}

	return
}
