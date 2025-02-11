// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestObject struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// object must implement JSONObject interface to be able to be passed to HTTP client methods
func (p TestObject) Decode(dec json.Decoder) (error, JSONObject) {
	return dec.Decode(&p), p
}

func TestHttpPut(t *testing.T) {
	// Set up a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodPut {
			t.Errorf("expected HTTP method to be PUT, got %s", r.Method)
		}
		if r.URL.Path != "/test/123" {
			t.Errorf("expected URL path to be /test/123, got %s", r.URL.Path)
		}

		// Write a mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "123", "name": "Test Object"}`))
	}))
	defer server.Close()

	// Create an HttpService instance and test the HttpPut function
	svc := HttpInit(server.URL+"/test", LogLevelDebug)
	obj := struct{ Name string }{Name: "Test Object"}
	res, err := svc.HttpPut("123", obj, &TestObject{}, RestOptions{})
	if err != nil {
		t.Errorf("HttpPut returned an error: %v", err)
	}
	if res == nil {
		t.Error("HttpPut returned a nil result")
	}
}

func TestHttpPost(t *testing.T) {
	// Set up a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodPost {
			t.Errorf("expected HTTP method to be POST, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("expected URL path to be /test, got %s", r.URL.Path)
		}

		// Write a mock response
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "123", "name": "Test Object"}`))
	}))
	defer server.Close()

	// Create an HttpService instance and test the HttpPost function
	svc := HttpInit(server.URL+"/test", LogLevelDebug)
	obj := struct{ Name string }{Name: "Test Object"}
	res, err := svc.HttpPost(obj, &TestObject{}, RestOptions{})
	if err != nil {
		t.Errorf("HttpPost returned an error: %v", err)
	}
	if res == nil {
		t.Error("HttpPost returned a nil result")
	}
}

func TestHttpGet(t *testing.T) {
	// Set up a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodGet {
			t.Errorf("expected HTTP method to be GET, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("expected URL path to be /test, got %s", r.URL.Path)
		}

		// Write a mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "123", "name": "Test Object 1"}, {"id": "456", "name": "Test Object 2"}]`))
	}))
	defer server.Close()

	// Create an HttpService instance and test the HttpGet function
	svc := HttpInit(server.URL+"/test", LogLevelDebug)
	res, err := svc.HttpGet(map[string]string{}, &TestObject{}, RestOptions{})
	if err != nil {
		t.Errorf("HttpGet returned an error: %v", err)
	}
	if res == nil {
		t.Error("HttpGet returned a nil result")
	}
}

func TestHttpDelete(t *testing.T) {
	// Set up a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodDelete {
			t.Errorf("expected HTTP method to be DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/test/123" {
			t.Errorf("expected URL path to be /test/123, got %s", r.URL.Path)
		}

		// Write a mock response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create an HttpService instance and test the HttpDelete function
	svc := HttpInit(server.URL+"/test", LogLevelDebug)
	res, err := svc.HttpDelete("123", map[string]string{"param": "p1"}, &TestObject{}, RestOptions{})
	if err != nil {
		t.Errorf("HttpDelete returned an error: %v", err)
	}
	if res == nil {
		t.Error("HttpDelete returned a nil result")
	}
}
