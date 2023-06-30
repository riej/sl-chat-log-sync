package main

type ChatLogsStorage interface {
	GetAccountNames() ([]string, error)
	ListChatLogFileNames(accountName string) (absolutePaths []string, relativePaths []string, err error)
	ReadChatLog(accountName string, fileName string) (Messages, error)
	WriteChatLog(accountName string, fileName string, messages Messages) error
}
