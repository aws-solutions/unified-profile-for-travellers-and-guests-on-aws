// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"errors"
)

/*
Custom error types for LCS.

See more: https://go.dev/blog/go1.13-errors
*/

// Custom error type to indicate an error can be retried
type RetryableError struct {
	Err error
}

// Returns error message prefixed with "retryable: "
func (e *RetryableError) Error() string {
	return "retryable: " + e.Err.Error()
}

// Return the underlying error, allowing us to use RetryableError in a chain of errors
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// Wrap an existing error with its full error chain inside a RetryableError
func NewRetryable(err error) error {
	return &RetryableError{Err: err}
}

// Check if an error contains a RetryableError anywhere in the error chain
//
// Usage:
//
//	err := someFunc()
//	if IsRetryableError(err) {
//	  // handle retryable error
//	}
func IsRetryableError(err error) bool {
	var retryableError *RetryableError
	return errors.As(err, &retryableError)
}
