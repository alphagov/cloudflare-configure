package main

import (
	"fmt"
	"net/http"
)

type CloudFlareQuery struct {
	RootURL   string
	AuthEmail string
	AuthKey   string
}

func (q *CloudFlareQuery) NewRequest(method, path string) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", q.RootURL, path)

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("X-Auth-Email", q.AuthEmail)
	request.Header.Set("X-Auth-Key", q.AuthKey)

	return request, nil
}
