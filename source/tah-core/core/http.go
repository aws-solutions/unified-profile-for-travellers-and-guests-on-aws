// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

////////////////////////////////////////////////////////
// This code creates and HTTP client to be used across all solution to interaction with REST APIS
// it implements standard behaviors for transaction traceability, SigV4 (if needed), query param encoding...
// TODO: Sanitized inputs and RestOptions (example trailing or leading 'slashes')
////////////////////////////////////////////////////////

// type of object that is manageable by a micro service
type JSONObject interface {
	//ParseBody(body io.ReadCloser) (interface{}, error)
	Decode(dec json.Decoder) (error, JSONObject)
}

/****************
* HTTP WRAPPERS
******************/

type HttpService struct {
	Endpoint string
	Tx       Transaction
	Headers  map[string]string
	Region   string
}

type RestOptions struct {
	SubEndpoint  string
	Headers      map[string]string
	SigV4Service string
	Debug        bool
}

type IHttpServiceConfig interface {
	HttpPut(id string, object interface{}, tpl JSONObject, options RestOptions) (interface{}, error)
	HttpPost(object interface{}, tpl JSONObject, options RestOptions) (interface{}, error)
	HttpGet(params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error)
	HttpDelete(id string, params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error)
}

func HttpInit(endpoint string, logLevel LogLevel) HttpService {
	tx := NewTransaction("CORE", "", logLevel)
	tx.Info("[HTTP] initializing HTTP client with endpoint %v\n", endpoint)
	return HttpService{Endpoint: endpoint, Tx: tx, Headers: map[string]string{}}
}

func HttpInitWithRegion(endpoint string, region string, logLevel LogLevel) HttpService {
	tx := NewTransaction("CORE", "", logLevel)
	tx.Info("[HTTP] initializing HTTP client with endpoint %v\n", endpoint)
	return HttpService{Endpoint: endpoint, Tx: tx, Headers: map[string]string{}, Region: region}
}

func (s HttpService) SetHeader(key string, val string) {
	s.Headers[key] = val
}

func (s *HttpService) SetTx(tx Transaction) {
	tx.LogPrefix = "CORE"
	s.Tx = tx
}

func processResponse(resp *http.Response) (*http.Response, error) {
	if resp.StatusCode >= 400 {
		return resp, errors.New(resp.Status)
	}
	return resp, nil
}

func (s HttpService) HttpPut(id string, object interface{}, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if id != "" {
		endpoint = endpoint + "/" + id
	}
	s.Tx.Info("[HTTP] PUT %v", endpoint)
	// s.Tx.Log("[HTTP] Body: %+v", object)

	var resp *http.Response
	bytesRepresentation, err := json.Marshal(object)
	if err != nil {
		s.Tx.Error("[HTTP] error marshalling object: %+v", err)
		return tpl, err
	}
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		s.Tx.Error("[HTTP] error marshalling object: %+v", err)
		return tpl, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Add(headerName, headerValue)
	}
	resp, err = client.Do(req)
	s.Tx.Info("[HTTP] PUT RESPONSE %v\n", resp.Status)
	if err == nil && resp != nil {
		resp, err = processResponse(resp)
		res, _ := s.ParseBody(resp.Body, tpl)
		// s.Tx.Log("[HTTP] PUT RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) HttpPost(object interface{}, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if options.SubEndpoint != "" {
		endpoint = endpoint + "/" + options.SubEndpoint
	}
	s.Tx.Info("[HTTP] POST %v\n", endpoint)
	// s.Tx.Log("[HTTP] Body: %+v\n", object)

	var resp *http.Response
	bytesRepresentation, err := json.Marshal(object)
	// s.Tx.Log("[HTTP][DEBUG] POST REQUEST Body Json %+v\n", string(bytesRepresentation))
	if err != nil {
		s.Tx.Error("[HTTP] error marshalling object: %+v", err)
		return tpl, err
	}
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		s.Tx.Error("[HTTP] error marshalling object: %+v", err)
		return tpl, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Set(headerName, headerValue)
	}
	// s.Tx.Log("[HTTP][DEBUG] POST REQUEST Headers %+v\n", req.Header)

	if options.SigV4Service != "" {
		s.Tx.Debug("[HTTP][SIGV4] Adding SigV4")
		creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), os.Getenv("AWS_SESSION_TOKEN"))
		s.Tx.Debug("[HTTP][SIGV4] creds:  %+v\n", creds)
		signer := v4.NewSigner(creds)
		s.Tx.Debug("[HTTP][SIGV4] signer:  %+v\n", signer)
		bodyPayload := strings.NewReader(string(bytesRepresentation))
		signer.Sign(req, bodyPayload, options.SigV4Service, s.Region, time.Now())
	}

	resp, err = client.Do(req)
	s.Tx.Info("[HTTP] POST RESPONSE %v\n", resp.Status)
	if err == nil && resp != nil {
		if options.Debug {
			b, _ := io.ReadAll(resp.Body)
			s.Tx.Debug("[HTTP][DEBUG] POST RESPONSE Body %v\n", string(b))
			return tpl, err
		}
		resp, err = processResponse(resp)
		res, _ := s.ParseBody(resp.Body, tpl)
		// s.Tx.Log("[HTTP] POST RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) HttpGet(params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if options.SubEndpoint != "" {
		endpoint = endpoint + "/" + options.SubEndpoint
	}
	s.Tx.Info("[HTTP] GET %v\n", endpoint)
	var resp *http.Response
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Add(headerName, headerValue)
	}
	if err != nil {
		s.Tx.Error("[HTTP] error marshalling object: %+v", err)
		return tpl, err
	}
	//Adding params
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()
	//Workaround: encode comma to support comma separated params
	req.URL.RawQuery = strings.ReplaceAll(req.URL.RawQuery, "%2C", ",")
	// s.Tx.Log("[HTTP] GET REQUEST %v\n", req)
	resp, err = client.Do(req)
	s.Tx.Info("[HTTP] GET RESPONSE %v\n", resp.Status)
	if err == nil {
		resp, err = processResponse(resp)

		res, _ := s.ParseBody(resp.Body, tpl)
		// s.Tx.Log("[HTTP] GET RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) HttpDelete(id string, params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if options.SubEndpoint != "" {
		endpoint = endpoint + "/" + options.SubEndpoint
	}
	if id != "" {
		endpoint = endpoint + "/" + id
	}
	s.Tx.Info("[HTTP] DELETE %v\n", s.Endpoint)
	var resp *http.Response
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Add(headerName, headerValue)
	}
	if err != nil {
		s.Tx.Error("[HTTP] error marshalling object: %+v", err)
		return tpl, err
	}
	//Adding params
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()
	//Workaround: encode comma to support comma separated params
	req.URL.RawQuery = strings.ReplaceAll(req.URL.RawQuery, "%2C", ",")
	// s.Tx.Log("[HTTP] DELETE REQUEST %v\n", req)
	resp, err = client.Do(req)
	s.Tx.Info("[HTTP] DELETE RESPONSE %v\n", resp.Status)
	if err == nil {
		resp, err = processResponse(resp)

		res, _ := s.ParseBody(resp.Body, tpl)
		// s.Tx.Log("[HTTP] DELETE RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) ParseBody(body io.ReadCloser, p JSONObject) (interface{}, error) {
	s.Tx.Debug("[HTTP] parsing json body into %v", reflect.TypeOf(p))
	dec := json.NewDecoder(body)
	var err error = nil
	for {
		//if err = dec.Decode(&p); err == io.EOF {
		if err, p = p.Decode(*dec); err == io.EOF {
			break
		} else if err != nil {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return p, err
}
