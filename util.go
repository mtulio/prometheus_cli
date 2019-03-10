package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func formatCSV(rows [][]string, delim rune) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = delim
	for _, row := range rows {
		w.Write(row)
		if err := w.Error(); err != nil {
			panic("error formatting CSV")
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic("error dumping CSV")
	}
	return buf.String()
}

// Convert an interface into a desired struct
func interfaceToStruct(i interface{}, target interface{}) error {
	data, _ := json.Marshal(i)
	err := json.Unmarshal(data, target)

	return err
}

func SendRequestReadResponse(c *Client, url string) ([]byte, error) {
	fmt.Println("URL: ", url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("Resp: ", string(buf))
	return buf, nil
}
