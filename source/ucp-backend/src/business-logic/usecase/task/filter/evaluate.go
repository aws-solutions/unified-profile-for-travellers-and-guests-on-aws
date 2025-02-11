// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"log"
	"strconv"
	"strings"
)

func shouldIncludeObject(algorithm string, context map[string]interface{}) bool {

	// Split the algorithm into individual tokens
	tokens := strings.Fields(algorithm)

	// If there is no condition in the filter, the record should be included
	if len(tokens) == 0 {
		return true
	}

	log.Printf("Tokens: %v", tokens)

	// Evaluate expression inside parenthesis first,
	// by recursively calling function on subexpression
	tokens, ok := recursivelyEvaluate(algorithm, tokens, context)
	if !ok {
		return false
	}

	tokens = handleOperators(tokens, context)

	// The final token should be the result of the algorithm
	result, err := strconv.ParseBool(tokens[0])
	return (err == nil) && result
}

func recursivelyEvaluate(algorithm string, tokens []string, context map[string]interface{}) ([]string, bool) {
	// Evaluate expression inside parenthesis first,
	// by recursively calling function on subexpression
	done := false
	for !done {
		isOpen := false
		openingIndex := -1
		for i := 0; i < len(tokens); i++ {
			isOpen, openingIndex = openingParen(tokens, i, openingIndex, isOpen)
			if tokens[i][len(tokens[i])-1] == ')' {
				if !isOpen {
					log.Printf("Error: unmatched parenthesis in algorithm: %v", algorithm)
					return []string{}, false
				}
				// Evaluate and return innermost subexpression
				parts := tokens[openingIndex : i+1]
				subexpression := strings.Trim(strings.Join(parts, " "), "()")

				// Recursively call this program
				subResult := shouldIncludeObject(subexpression, context)

				// replace subexpression with result
				tokens = append(tokens[:openingIndex], append([]string{strconv.FormatBool(subResult)}, tokens[i+1:]...)...)
				isOpen = false
				break
			}
		}
		if isOpen {
			log.Printf("Error: unmatched parenthesis in algorithm: %v", algorithm)
			return []string{}, false
		}
		if openingIndex == -1 {
			done = true
		}
	}
	return tokens, true
}

func openingParen(tokens []string, i int, openingIndex int, isOpen bool) (bool, int) {
	if tokens[i][0] == '(' {
		isOpen = true
		openingIndex = i
	}
	return isOpen, openingIndex
}

func handleOperators(tokens []string, context map[string]interface{}) []string {
	operators := []string{"and", "or", "=", "!=", "<", "<=", ">", ">=", "in"}
	// Loop through operators
	for _, op := range operators {
		for i := 0; i < len(tokens); i++ {
			// Check if the current token is the current operation
			tokens, i = handleTokenFromOperation(tokens, op, i, context)
		}
	}
	return tokens
}

func handleTokenFromOperation(tokens []string, op string, i int, context map[string]interface{}) ([]string, int) {
	if tokens[i] == op {
		// Special consideration for and / or
		if op == "and" || op == "or" {
			tokens, i = handleAndOr(tokens, op, i, context)
		} else {
			// Evaluate the operation based on the surrounding tokens
			left, right := tokens[i-1], tokens[i+1]
			result := evaluateOperation(left, right, op, context)
			// Replace the tokens with the result of the operation
			tokens = append(tokens[:i-1], append([]string{strconv.FormatBool(result)}, tokens[i+2:]...)...)
			i -= 2 // Move the index back 2 places to account for the removal of two tokens
		}
	}
	return tokens, i
}

func handleAndOr(tokens []string, op string, i int, context map[string]interface{}) ([]string, int) {
	var result1 bool
	var result2 bool
	var tokensLeft int
	var tokensRight int
	// Evaluate left side
	if tokens[i-1] == "true" || tokens[i-1] == "false" {
		result1, _ = strconv.ParseBool(tokens[i-1])
		tokensLeft = 1
	} else {
		left1, op1, right1 := tokens[i-3], tokens[i-2], tokens[i-1]
		result1 = evaluateOperation(left1, right1, op1, context)
		tokensLeft = 3
	}
	// Evaluate right side
	// Side already evaluated
	if tokens[i+1] == "true" || tokens[i+1] == "false" {
		result2, _ = strconv.ParseBool(tokens[i+1])
		tokensRight = 1
		// Side not evaluated yet
	} else {
		left2, op2, right2 := tokens[i+1], tokens[i+2], tokens[i+3]
		result2 = evaluateOperation(left2, right2, op2, context)
		tokensRight = 3
	}
	// Update tokens
	if op == "and" {
		tokens = append(tokens[:i-tokensLeft], append([]string{strconv.FormatBool(result1 && result2)}, tokens[i+tokensRight+1:]...)...)
	} else {
		tokens = append(tokens[:i-tokensLeft], append([]string{strconv.FormatBool(result1 || result2)}, tokens[i+tokensRight+1:]...)...)
	}
	i -= tokensLeft - tokensRight + 2 // Move the index back 2 places to account for the removal of two tokens
	return tokens, i
}
