package main

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

const headerEmail = "X-Auth-Email"
const headerKey = "X-Auth-Key"

func TestBuildingAnHTTPRequest(t *testing.T) {
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

func TestSettingARequestBody(t *testing.T) {
	query := &CloudFlareQuery{
		RootURL:   "foo.com",
		AuthEmail: "user@example.com",
		AuthKey:   "abc123",
	}

	request, err := query.NewRequestBody("POST", "/foo", bytes.NewBuffer([]byte(`{"a": 1}`)))
	if err != nil {
		t.Fatal("Should've built a request with a body with no errors", err)
	}

	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		t.Fatal("Couldn't read the response body without errors", err)
	}

	if !reflect.DeepEqual(body, []byte(`{"a": 1}`)) {
		t.Fatal("The body should've been set to some JSON we provided but was:", body)
	}
}
