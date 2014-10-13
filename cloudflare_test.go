package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func testCloudFlareServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprintf(w, body)
	}))
}

func TestGettingZones(t *testing.T) {
	testServer := testCloudFlareServer(200, `{
		"errors": [],
		"messages": [],
		"result": [{"id": "foo", "name": "bar"}],
		"success": true
	}`)
	defer testServer.Close()

	expectedZones := []CloudFlareZoneItem{
		CloudFlareZoneItem{
			ID:   "foo",
			Name: "bar",
		},
	}

	query := &CloudFlareQuery{RootURL: testServer.URL}
	cloudFlare := NewCloudFlare(query)

	zones, err := cloudFlare.Zones()
	if err != nil {
		t.Fatalf("Expected to get zones with no errors", err.Error())
	}

	if len(zones) != len(expectedZones) {
		t.Fatal("Didn't get the right number of zones back")
	}

	if zones[0] != expectedZones[0] {
		t.Fatal("Not the zones we were looking for", zones)
	}
}

func TestMakingARequestWithoutSuccess(t *testing.T) {
	testServer := testCloudFlareServer(200, `{
		"errors": [],
		"messages": [],
		"result": [],
		"success": false
	}`)
	defer testServer.Close()

	query := &CloudFlareQuery{RootURL: testServer.URL}
	cloudFlare := NewCloudFlare(query)

	req, _ := query.NewRequest("GET", "/foo")
	_, err := cloudFlare.makeRequest(req)
	if err == nil {
		t.Fatalf("Expected to be notified if the response wasn't successful")
	}
}

func TestMakingARequestWithErrors(t *testing.T) {
	testServer := testCloudFlareServer(200, `{
		"errors": ["something bad"],
		"messages": [],
		"result": [],
		"success": true
	}`)
	defer testServer.Close()

	query := &CloudFlareQuery{RootURL: testServer.URL}
	cloudFlare := NewCloudFlare(query)

	req, _ := query.NewRequest("GET", "/foo")
	_, err := cloudFlare.makeRequest(req)
	if err == nil {
		t.Fatalf("Expected to be notified if the response wasn't successful")
	}
}

func TestMakingARequestWithout200Code(t *testing.T) {
	testServer := testCloudFlareServer(500, ``)
	defer testServer.Close()

	query := &CloudFlareQuery{RootURL: testServer.URL}
	cloudFlare := NewCloudFlare(query)

	req, _ := query.NewRequest("GET", "/foo")
	_, err := cloudFlare.makeRequest(req)
	if err == nil {
		t.Fatalf("Expected to be notified if the response wasn't successful")
	}
}

func TestGettingSettings(t *testing.T) {
	const zoneID = "123"

	expectedSettings := []CloudFlareConfigItem{
		CloudFlareConfigItem{
			ID:         "always_online",
			Value:      "off",
			ModifiedOn: "2014-07-09T11:50:56.595672Z",
			Editable:   true,
		},
		CloudFlareConfigItem{
			ID:         "browser_cache_ttl",
			Value:      float64(14400),
			ModifiedOn: "2014-07-09T11:50:56.595672Z",
			Editable:   true,
		},
	}

	testServer := testCloudFlareServer(200, `{
		"errors": [], "messages": [], 
		"result": [
			{"id": "always_online", "value": "off", "modified_on": "2014-07-09T11:50:56.595672Z", "editable": true},
			{"id": "browser_cache_ttl", "value": 14400, "modified_on": "2014-07-09T11:50:56.595672Z", "editable": true}
		],
		"success": true
	}`)
	defer testServer.Close()

	query := &CloudFlareQuery{RootURL: testServer.URL}
	cloudFlare := NewCloudFlare(query)

	settings, err := cloudFlare.Settings(zoneID)
	if err != nil {
		t.Fatalf("Expected to get settings with no errors", err.Error())
	}
	if len(settings) != 2 {
		t.Fatalf("Expected 2 settings items, got %d", len(settings))
	}
	if !reflect.DeepEqual(settings, expectedSettings) {
		t.Fatal("Settings response doesn't match", settings)
	}
}
