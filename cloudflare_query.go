package main

import (
	"fmt"
	"io"
	"net/http"
)

type CloudFlareQuery struct {
	RootURL   string
	AuthEmail string
	AuthKey   string
}

func (q *CloudFlareQuery) NewRequest(method, path string) (*http.Request, error) {
	return q.NewRequestBody(method, path, nil)
}

func (q *CloudFlareQuery) NewRequestBody(method, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", q.RootURL, path)

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Auth-Email", q.AuthEmail)
	request.Header.Set("X-Auth-Key", q.AuthKey)

	return request, nil
}
