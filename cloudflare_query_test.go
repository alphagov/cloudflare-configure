package main

import (
	"testing"
)

func TestBuildingAnHTTPRequest(t *testing.T) {
	const headerEmail = "X-Auth-Email"
	const headerKey = "X-Auth-Key"

	query := &CloudFlareQuery{
		RootURL:   "foo.com",
		AuthEmail: "user@example.com",
		AuthKey:   "abc123",
	}

	request, err := query.NewRequest("GET", "/zones")
	if err != nil {
		t.Fatalf("Should've built request without errors", err.Error())
	}

	if request.Method != "GET" {
		t.Fatal("Expected a GET request")
	}

	if request.URL.String() != "foo.com/zones" {
		t.Fatal("Not the zones path for CF")
	}

	if val := request.Header.Get(headerEmail); val != query.AuthEmail {
		t.Error("AuthEmail incorrect:", val)
	}

	if val := request.Header.Get(headerKey); val != query.AuthKey {
		t.Error("AuthKey incorrect:", val)
	}
}
