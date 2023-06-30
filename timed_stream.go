package main

type TimedMessagesStream struct {
	sources       []Messages
	lastTimestamp int64
}

func (t *TimedMessagesStream) NextMessages() Messages {
	var messages []string
	var timestamp int64 = -1

	for _, source := range t.sources {
		if len(source) == 0 {
			continue
		}

		if timestamp < 0 || timestamp > source[0].Timestamp {
			timestamp = source[0].Timestamp
		}
	}

	for i, source := range t.sources {
		for {
			if len(source) == 0 {
				break
			}

			if source[0].Timestamp == timestamp {
				if !Contains(messages, source[0].Message) {
					messages = append(messages, source[0].Message)
				}
				source = source[1:]
				continue
			}

			break
		}

		t.sources[i] = source
	}

	result := make(Messages, len(messages))
	for i, message := range messages {
		result[i] = &Message{
			Timestamp: timestamp,
			Message:   message,
		}
	}

	t.lastTimestamp = timestamp

	return result
}
