// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"log"
	"regexp"
)

var TRANSACTION_ID_HEADER = "aws-tah-tx-id"
var OBFUSCATION_STRING = "********************"

type LogLevel uint8

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// todo: include in documentation
const DefaultLogLevel = LogLevelInfo

// type of object that is manageable by a micro service
type Transaction struct {
	TransactionID          string
	LogPrefix              string
	User                   string
	LogObfuscationPatterns []LogObfuscationPattern
	LogLevel               LogLevel
}

type LogObfuscationPattern struct {
	Prefix  string
	Pattern string
	Suffix  string
}

func LogLevelFromString(logLevel string) LogLevel {
	switch logLevel {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	default:
		return DefaultLogLevel
	}
}

func NewTransaction(logPrefix string, transactionID string, logLevel LogLevel) Transaction {
	if transactionID == "" {
		transactionID = GenerateUniqueId()
	}
	if logPrefix == "" {
		logPrefix = "NO_PREFIX_CONFIGURED"
	}

	switch LogLevel(logLevel) {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
	default:
		logLevel = DefaultLogLevel
	}

	return Transaction{
		TransactionID: transactionID,
		LogPrefix:     logPrefix,
		LogLevel:      logLevel,
	}
}

func NewTransactionWithUser(logPrefix string, transactionID string, user string, logLevel LogLevel) Transaction {
	tx := NewTransaction(logPrefix, transactionID, logLevel)
	tx.User = user
	return tx
}

func (t *Transaction) SetPrefix(prefix string) {
	t.LogPrefix = prefix
}

func (t Transaction) isLevelEnabled(level LogLevel) bool {
	return t.LogLevel <= level
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

func (t Transaction) log(level LogLevel, format string, v ...interface{}) {
	if t.isLevelEnabled(level) {
		logLine := t.BuildLog(format, v...)
		obfuscated := t.Obfuscate(logLine)
		log.Println(obfuscated)
	}
}

var WARNING_PREFIX = "[tah_warning]"

func (t Transaction) Warn(format string, v ...interface{}) {
	t.log(LogLevelWarn, WARNING_PREFIX+" "+format, v...)
}

var ERROR_PREFIX = "[tah_error]"

func (t Transaction) Error(format string, v ...interface{}) {
	t.log(LogLevelError, ERROR_PREFIX+" "+format, v...)
}

func (t Transaction) Debug(format string, v ...interface{}) {
	t.log(LogLevelDebug, format, v...)
}

func (t Transaction) Info(format string, v ...interface{}) {
	t.log(LogLevelInfo, format, v...)
}
