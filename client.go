// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Prometheus HTTP API client functionality.
//
// TODO(julius): This functionality should be moved to a separate
// library/repository once we have a good name for it (client_* is already used
// for the interface between metrics-exposing servers and Prometheus).

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	queryPath      = "/api/v1/query"
	queryRangePath = "/api/v1/query_range"
	metricsPath    = "/api/v1/metrics"
)

// Client is a client for executing queries against the Prometheus API.
type Client struct {
	Endpoint   string
	httpClient http.Client
}

// transport builds a new transport with the provided timeout.
func transport(netw, addr string, timeout time.Duration) (connection net.Conn, err error) {
	deadline := time.Now().Add(timeout)
	connection, err = net.DialTimeout(netw, addr, timeout)
	if err == nil {
		connection.SetDeadline(deadline)
	}

	return
}

// NewClient creates a new Client, given a server URL and timeout.
func NewClient(url string, timeout time.Duration) *Client {
	return &Client{
		Endpoint: url,
		httpClient: http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) { return transport(netw, addr, timeout) },
			},
		},
	}
}

// Query performs an instant expression query via the Prometheus API.
func (c *Client) Query(expr string) (QueryResponse, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}

	u.Path = strings.TrimRight(u.Path, "/") + queryPath
	q := u.Query()

	q.Set("query", expr)
	u.RawQuery = q.Encode()

	buf, err := SendRequestReadResponse(c, u.String())
	if err != nil {
		return nil, err
	}

	var r BaseResponse
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, err
	}

	if r.Status == STATUS_ERROR {
		return nil, fmt.Errorf(fmt.Sprintf("ErrorType: %s, Error: %s",
			r.ErrorType, r.Error))
	}

	var typedResp QueryResponse

	switch r.RespData.ResultType {

	case SCALAR_TYPE:
		typedResp = &ScalarQueryResponse{}
	case STRING_TYPE:
		typedResp = &ScalarQueryResponse{}
	case VECTOR_TYPE:
		typedResp = &VectorQueryResponse{}
	case MATRIX_TYPE:
		typedResp = &MatrixQueryResponse{}
	default:
		return nil, fmt.Errorf("invalid response type %s", r.RespData.ResultType)
	}

	if err := interfaceToStruct(r.RespData, typedResp); err != nil {
		return nil, err
	}

	// fill the data
	typedResp.Fill()
	return typedResp, err
}

// QueryRange performs an range expression query via the Prometheus API.
func (c *Client) QueryRange(expr string, end float64, rangeSec uint64, step uint64) (*MatrixQueryResponse, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}

	u.Path = strings.TrimRight(u.Path, "/") + queryRangePath
	q := u.Query()

	q.Set("query", expr)
	q.Set("end", fmt.Sprintf("%f", end))
	q.Set("range", fmt.Sprintf("%d", rangeSec))
	q.Set("step", fmt.Sprintf("%d", step))
	u.RawQuery = q.Encode()

	buf, err := SendRequestReadResponse(c, u.String())
	if err != nil {
		return nil, err
	}

	var r BaseResponse
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, err
	}

	if r.Status == STATUS_ERROR {
		return nil, fmt.Errorf(fmt.Sprintf("ErrorType: %s, Error: %s",
			r.ErrorType, r.Error))
	}

	switch r.RespData.ResultType {
	case MATRIX_TYPE:
		typedResp := &MatrixQueryResponse{}
		if err := interfaceToStruct(r.RespData, typedResp); err != nil {
			return nil, err
		}
		return typedResp, nil

	default:
		return nil, fmt.Errorf("invalid response type %s", r.RespData.ResultType)
	}
}

// Metrics retrieves the list of available metric names via the Prometheus API.
func (c *Client) Metrics() ([]string, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}

	u.Path = strings.TrimRight(u.Path, "/") + metricsPath
	buf, err := SendRequestReadResponse(c, u.String())
	if err != nil {
		return nil, err
	}

	var r []string
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, err
	}

	return r, nil
}
