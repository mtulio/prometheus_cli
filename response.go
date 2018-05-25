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

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/common/model"
)

const (
	STATUS_ERROR   = "error"
	STATUS_SUCCESS = "success"

	STRING_TYPE = "string"
	SCALAR_TYPE = "scalar"
	VECTOR_TYPE = "vector"
	MATRIX_TYPE = "matrix"
	ERROR_TYPE  = "error"
)

// Interface for query results of various result types.
type QueryResponse interface {
	ToText() string
	ToCSV(delim rune) string
	Fill()
}

// Type for unserializing a generic Prometheus query response.
type StubQueryResponse struct {
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
	Version int         `json:"version"`
}

type Data struct {
	ResultType string      `json:"resultType"`
	Result     interface{} `json:"result"`
}

type BaseResponse struct {
	RespData  Data   `json:"data,omitempty"`
	Status    string `json:"status"`
	ErrorType string `json:"errorType,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Type for unserializing a scalar-typed and string typed
// Prometheus query response.
type ScalarQueryResponse struct {
	ActualValue []interface{} `json:"result"`
	Value       string
	Timestamp   float64
}

func (r *ScalarQueryResponse) ToCSV(delim rune) string {
	return formatCSV([][]string{{r.Value}}, delim)
}

func (r *ScalarQueryResponse) ToText() string {
	return fmt.Sprint(r.Value)
}

func (r *ScalarQueryResponse) Fill() {
	r.Timestamp = r.ActualValue[0].(float64)
	r.Value = r.ActualValue[1].(string)
}

// Type for unserializing a vector-typed Prometheus query response.
type VectorQueryResponse struct {
	Value []struct {
		Metric model.Metric `json:"metric"`
		// Supposed to have 2 entries, 0: float64, 1: string
		ActualValue []interface{} `json:"value"`
		Value       string
		Timestamp   float64
	} `json:"result"`
}

func (r *VectorQueryResponse) ToText() string {
	lines := make([]string, 0, len(r.Value))
	for _, v := range r.Value {
		lines = append(lines, fmt.Sprintf("%s %s@%.3f\n", v.Metric, v.Value, v.Timestamp))
	}
	return strings.Join(lines, "")
}

func (r *VectorQueryResponse) ToCSV(delim rune) string {
	rows := make([][]string, 0, len(r.Value))
	for _, v := range r.Value {
		rows = append(rows, []string{
			v.Metric.String(),
			v.Value,
			strconv.FormatFloat(v.Timestamp, 'f', -1, 64),
		})
	}
	return formatCSV(rows, delim)
}

func (r *VectorQueryResponse) Fill() {
	for i := 0; i < len(r.Value); i++ {
		r.Value[i].Timestamp = r.Value[i].ActualValue[0].(float64)
		r.Value[i].Value = r.Value[i].ActualValue[1].(string)
	}
}

// Type for unserializing a matrix-typed Prometheus query response.
type MatrixQueryResponse struct {
	Value []struct {
		Metric model.Metric `json:"metric"`
		Values [][]interface{}
	} `json:"result"`
}

func (r *MatrixQueryResponse) ToText() string {
	lines := make([]string, 0, len(r.Value))
	for _, v := range r.Value {
		vals := make([]string, 0, len(v.Values))
		for _, s := range v.Values {
			vals = append(vals, fmt.Sprintf("%s@%.3f ", s[1], s[0]))
		}
		lines = append(lines, fmt.Sprintf("%s %s\n", v.Metric, strings.Join(vals, " ")))
	}
	return strings.Join(lines, "")
}

func (r *MatrixQueryResponse) ToCSV(delim rune) string {
	rows := make([][]string, 0, len(r.Value))
	for _, v := range r.Value {
		vals := make([]string, 0, len(v.Values))
		for _, s := range v.Values {
			vals = append(vals, fmt.Sprintf("%s@%.3f", s[1], s[0]))
		}
		rows = append(rows, []string{
			v.Metric.String(),
			strings.Join(vals, " "),
		})
	}
	return formatCSV(rows, delim)
}

func (r *MatrixQueryResponse) Fill() {
	// do nothing
}
