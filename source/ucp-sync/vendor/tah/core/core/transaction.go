package core

import (
	"fmt"
	"log"
	"regexp"
)

var TRANSACTION_ID_HEADER = "aws-tah-tx-id"
var OBFUSCATION_STRING = "********************"

//type of object that is manageable by a micro service
type Transaction struct {
	TransactionID          string
	LogPrefix              string
	User                   string
	LogObfuscationPatterns []LogObfuscationPattern
}
type LogObfuscationPattern struct {
	Prefix  string
	Pattern string
	Suffix  string
}

func NewTransaction(logPrefix string, transactionID string) Transaction {
	if transactionID == "" {
		transactionID = GeneratUniqueId()
	}
	if logPrefix == "" {
		logPrefix = "NO_PREFIX_CONFIGURED"
	}
	return Transaction{
		TransactionID: transactionID,
		LogPrefix:     logPrefix,
	}
}

func NewTransactionWithUser(logPrefix string, transactionID string, user string) Transaction {
	tx := NewTransaction(logPrefix, transactionID)
	tx.User = user
	return tx
}

func (t *Transaction) SetPrefix(prefix string) {
	t.LogPrefix = prefix
}

func (t Transaction) addPrefix(text string) string {
	logline := "[tx-"
	logline += t.TransactionID
	logline += "] "
	if t.LogPrefix != "" {
		logline += "["
		logline += t.LogPrefix
		logline += "] "
	}
	logline += text
	return logline
}

func (t *Transaction) AddLogObfuscationPattern(prefix string, pattern string, suffix string) {
	t.LogObfuscationPatterns = append(t.LogObfuscationPatterns, LogObfuscationPattern{
		Prefix:  prefix,
		Pattern: pattern,
		Suffix:  suffix,
	})
}

func (t Transaction) Obfuscate(logLine string) string {
	obfuscated := logLine
	for _, pattern := range t.LogObfuscationPatterns {
		re := regexp.MustCompile(pattern.Prefix + pattern.Pattern + pattern.Suffix)
		obfuscated = re.ReplaceAllString(obfuscated, pattern.Prefix+OBFUSCATION_STRING+pattern.Suffix)
	}
	return obfuscated
}

func (t Transaction) BuildLog(format string, v ...interface{}) string {
	return fmt.Sprintf(t.addPrefix(format), v...)
}

func (t Transaction) Log(format string, v ...interface{}) {
	logLine := t.BuildLog(format, v...)
	obfuscated := t.Obfuscate(logLine)
	log.Printf(obfuscated)
}
