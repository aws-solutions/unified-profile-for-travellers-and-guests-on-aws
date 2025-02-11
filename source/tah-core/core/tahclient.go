// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import "net/http"

const (
	TEST_SOLUTION_ID      = "SO0244"
	TEST_SOLUTION_VERSION = "0.0.0"
)

// Custom implementation of the http.RoundTripper interface to add the solution,
// allowing us access to request headers.
type TahTransport struct {
	Base            http.RoundTripper
	SolutionId      string
	SolutionVersion string
}

// RoundTrip implementation that adds AWS Solution info to the request
// https://pkg.go.dev/net/http#RoundTripper
func (t *TahTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Leave original headers if id and version are not provided
	if t.SolutionId == "" || t.SolutionVersion == "" {
		return t.Base.RoundTrip(req)
	}

	// match format AWSSOLUTION/SO0244/v1.0.0
	original := req.Header.Get("User-Agent")
	solutionData := "AWSSOLUTION/" + t.SolutionId + "/v" + t.SolutionVersion
	req.Header.Set("User-Agent", original+" "+solutionData)
	return t.Base.RoundTrip(req)
}

func CreateClient(solutionId string, solutionVersion string) *http.Client {
	return &http.Client{
		Transport: &TahTransport{
			Base:            http.DefaultTransport,
			SolutionId:      solutionId,
			SolutionVersion: solutionVersion,
		},
	}
}

func CreateCustomClient(solutionId, solutionVersion string, transport *http.Transport) *http.Client {
	return &http.Client{
		Transport: &TahTransport{
			Base:            transport,
			SolutionId:      solutionId,
			SolutionVersion: solutionVersion,
		},
	}
}
