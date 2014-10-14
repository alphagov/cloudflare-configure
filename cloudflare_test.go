package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

var _ = Describe("CloudFlare", func() {
	var (
		server     *ghttp.Server
		query      *CloudFlareQuery
		cloudFlare *CloudFlare
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		query = &CloudFlareQuery{RootURL: server.URL()}
		cloudFlare = NewCloudFlare(query)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Zones()", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/zones"),
					ghttp.RespondWith(http.StatusOK, `{
						"errors": [],
						"messages": [],
						"result": [
							{"id": "123", "name": "foo"},
							{"id": "456", "name": "bar"}
						],
						"success": true
					}`),
				),
			)
		})

		It("should return two CloudFlareZoneItems", func() {
			zones, err := cloudFlare.Zones()

			Expect(err).To(BeNil())
			Expect(zones).To(Equal([]CloudFlareZoneItem{
				CloudFlareZoneItem{
					ID:   "123",
					Name: "foo",
				},
				CloudFlareZoneItem{
					ID:   "456",
					Name: "bar",
				},
			}))
		})
	})

	Describe("MakeRequest()", func() {
		var req *http.Request

		BeforeEach(func() {
			req, _ = query.NewRequest("GET", "/something")
		})

		Context("200, success: false, errors: []", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusOK, `{
							"errors": [],
							"messages": [],
							"result": [],
							"success": false
						}`),
					),
				)
			})

			It("should return error", func() {
				resp, err := cloudFlare.MakeRequest(req)

				Expect(resp).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})

		Context("200, success: true, errors: [something bad]", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusOK, `{
							"errors": ["something bad"],
							"messages": [],
							"result": [],
							"success": true
						}`),
					),
				)
			})

			It("should return error", func() {
				resp, err := cloudFlare.MakeRequest(req)

				Expect(resp).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})

		Context("500, empty body", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusServiceUnavailable, ""),
					),
				)
			})

			It("should return error", func() {
				resp, err := cloudFlare.MakeRequest(req)

				Expect(resp).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})
})

func testCloudFlareServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprintf(w, body)
	}))
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
		"errors": [],
		"messages": [], 
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

func TestChangeSetting(t *testing.T) {
	const zoneID = "123"
	const settingID = "always_online"
	const settingVal = "off"

	receivedRequest := false

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequest = true

		if method := r.Method; method != "PATCH" {
			t.Fatal("Incorrect request method", method)
		}

		expectedURL := fmt.Sprintf("/zones/%s/settings/%s", zoneID, settingID)
		if !strings.HasSuffix(r.URL.String(), expectedURL) {
			t.Fatal("Request URL was incorrect")
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal("Unable to read request body", err)
		}

		var setting CloudFlareRequestItem
		err = json.Unmarshal(body, &setting)
		if err != nil {
			t.Fatal("Unable to parse request body", err)
		}

		expectedSetting := &CloudFlareRequestItem{
			Value: settingVal,
		}
		if !reflect.DeepEqual(setting, *expectedSetting) {
			t.Fatal("Request was incorrect", setting, expectedSetting)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"errors": [],
			"messages": [], 
			"result": {
				"id": "always_online",
				"value": "off",
				"modified_on": "2014-07-09T11:50:56.595672Z",
				"editable": true
			},
			"success": true
		}`)
	}))
	defer testServer.Close()

	query := &CloudFlareQuery{
		RootURL:   testServer.URL,
		AuthEmail: "user@example.com",
		AuthKey:   "abc123",
	}
	cloudFlare := NewCloudFlare(query)

	err := cloudFlare.Set(zoneID, settingID, settingVal)
	if err != nil {
		t.Fatal("Unable to set setting")
	}

	if !receivedRequest {
		t.Fatal("Expected test server to receive request")
	}
}
