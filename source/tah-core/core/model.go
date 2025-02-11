// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var DYNAMO_STREAM_EVENT_NAME_CREATE string = "INSERT"
var DYNAMO_STREAM_EVENT_NAME_UPDATE string = "MODIFY"
var DYNAMO_STREAM_EVENT_NAME_DELETE string = "REMOVE"

var VALIDATION_ERROR_PREFIX = "VALIDATION_ERROR"
var FUNCTIONAL_ERROR_PREFIX = "FUNCTIONAL_ERROR"
var TECHNICAL_ERROR_PREFIX = "TECHNICAL_ERROR"
var prefixes = []string{VALIDATION_ERROR_PREFIX, FUNCTIONAL_ERROR_PREFIX, TECHNICAL_ERROR_PREFIX}

type User struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Groups   string `json:"groups"`
	Setting  string `json:"setting"`
	//auth token
	Token string `json:"token"`
}

// Web Socket notification
type Notification struct {
	MessageType string            `json:"messageType"`
	MessageData map[string]string `json:"messageData"`
}

type ResError struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

func (r ResError) Error() string {
	b, err := json.Marshal(r)
	if err != nil {
		log.Println("cannot marshal error")
		panic(err)
	}
	return string(b[:])
}

func (r ResError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Type string `json:"type"`
	}{
		Code: r.Code,
		Msg:  r.Msg,
		Type: r.Type,
	})
}

// Build a ResError struct from a Go error based on Error text.
func BuildResError(err error) ResError {
	if err != nil {
		str := err.Error()
		return ResError{Code: MapCode(str), Msg: str}
	}
	return ResError{}
}

func BuildResErrors(errors []error) ResError {
	errorString := "Multiple Errors:\n"
	for _, err := range errors {
		errorString += err.Error() + "\n"
		return ResError{Code: MapCode(errorString), Msg: errorString}
	}
	return ResError{}
}

// get code form stringFormated error
func MapCode(str string) string {
	myRegexp := "((?:" + strings.Join(prefixes, "|") + `)_[0-9]*)\-.*`
	reg := regexp.MustCompile(myRegexp)
	found := reg.FindAllStringSubmatch(str, -1)
	log.Printf("[MapCode]  regexp: %+v | str: %v | found: %+v", myRegexp, str, found)
	if len(found) > 0 && len(found[0]) > 1 {
		return found[0][1]
	}
	return ""
}

// returns properly formated erropr test from code + msg input
func ValidationError(code int, msg string) ResError {
	return ResError{Type: VALIDATION_ERROR_PREFIX, Code: VALIDATION_ERROR_PREFIX + "_" + strconv.Itoa(code), Msg: msg}
}

func FunctionalError(code int, msg string) ResError {
	return ResError{Type: FUNCTIONAL_ERROR_PREFIX, Code: FUNCTIONAL_ERROR_PREFIX + "_" + strconv.Itoa(code), Msg: msg}
}

func TechnicalError(code int, msg string) ResError {
	return ResError{Type: TECHNICAL_ERROR_PREFIX, Code: TECHNICAL_ERROR_PREFIX + "_" + strconv.Itoa(code), Msg: msg}
}
