// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"github.com/stretchr/testify/mock"
)

// HttpServiceMock is a mock implementation of the HttpService interface
type HttpServiceMock struct {
	mock.Mock
}

var _ IHttpServiceConfig = &HttpServiceMock{}

func HttpInitMock() *HttpServiceMock {
	return &HttpServiceMock{}
}

func (m *HttpServiceMock) HttpPut(id string, object interface{}, tpl JSONObject, options RestOptions) (interface{}, error) {
	args := m.Called(id, object, tpl, options)
	return args.Get(0), args.Error(1)
}

func (m *HttpServiceMock) HttpPost(object interface{}, tpl JSONObject, options RestOptions) (interface{}, error) {
	args := m.Called(object, tpl, options)
	return args.Get(0), args.Error(1)
}

func (m *HttpServiceMock) HttpGet(params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error) {
	args := m.Called(params, tpl, options)
	return args.Get(0), args.Error(1)
}

func (m *HttpServiceMock) HttpDelete(id string, params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error) {
	args := m.Called(id, params, tpl, options)
	return args.Get(0), args.Error(1)
}
