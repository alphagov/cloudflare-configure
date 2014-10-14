package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"fmt"
	"net/http"
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
			Expect(err).To(BeNil())
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

	Describe("Settings()", func() {
		zoneID := "123"

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET",
						fmt.Sprintf("/zones/%s/settings", zoneID),
					),
					ghttp.RespondWith(http.StatusOK, `{
						"errors": [],
						"messages": [],
						"result": [
							{
								"id": "always_online",
								"value": "off",
								"modified_on": "2014-07-09T11:50:56.595672Z",
								"editable": true
							},
							{
								"id": "browser_cache_ttl",
								"value": 14400,
								"modified_on": "2014-07-09T11:50:56.595672Z",
								"editable": true
							}
						],
						"success": true
					}`),
				),
			)
		})

		It("should return two CloudFlareConfigItems", func() {
			settings, err := cloudFlare.Settings(zoneID)

			Expect(settings).To(Equal([]CloudFlareConfigItem{
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
			}))
			Expect(err).To(BeNil())
		})
	})

	Describe("Set()", func() {
		zoneID := "123"
		settingKey := "always_online"
		settingVal := "off"

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH",
						fmt.Sprintf("/zones/%s/settings/%s", zoneID, settingKey),
					),
					ghttp.VerifyJSON(fmt.Sprintf(`{"value": "%s"}`, settingVal)),

					ghttp.RespondWith(http.StatusOK, `{
						"errors": [],
						"messages": [],
						"result": {
							"id": "always_online",
							"value": "off",
							"modified_on": "2014-07-09T11:50:56.595672Z",
							"editable": true
						},
						"success": true
					}`),
				),
			)
		})

		It("should return two CloudFlareConfigItems", func() {
			err := cloudFlare.Set(zoneID, settingKey, settingVal)

			Expect(err).To(BeNil())
		})
	})
})
