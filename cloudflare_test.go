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

		It("should set the value with no errors", func() {
			Expect(cloudFlare.Set(zoneID, settingKey, settingVal)).To(BeNil())
		})
	})

	Describe("Compare()", func() {
		settingValAlwaysOnline := "on"
		settingValBrowserCache := 123

		It("should return nothing when local and remote are identical", func() {
			config, err := cloudFlare.Compare(
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(Equal(ConfigItems{}))
			Expect(err).To(BeNil())
		})

		It("should return all items in local when remote is empty", func() {
			config, err := cloudFlare.Compare(
				ConfigItems{},
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(Equal(ConfigItems{
				"always_online":     settingValAlwaysOnline,
				"browser_cache_ttl": settingValBrowserCache,
			}))
			Expect(err).To(BeNil())
		})

		It("should return one item in local overwriting always_online", func() {
			config, err := cloudFlare.Compare(
				ConfigItems{
					"always_online":     "off",
					"browser_cache_ttl": settingValBrowserCache,
				},
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(Equal(ConfigItems{
				"always_online": settingValAlwaysOnline,
			}))
			Expect(err).To(BeNil())
		})

		It("should return a public error when item is missing in local", func() {
			config, err := cloudFlare.Compare(
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
				ConfigItems{
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(
				ConfigMismatch{Missing: ConfigItems{"always_online": settingValAlwaysOnline}}))
		})
	})

	Describe("Update()", func() {
		zoneID := "123"
		settingValAlwaysOnline := "on"
		settingValBrowserCache := 123

		BeforeEach(func() {
			server.RouteToHandler("PATCH", fmt.Sprintf("/zones/%s/settings/always_online", zoneID),
				ghttp.CombineHandlers(
					ghttp.VerifyJSON(fmt.Sprintf(`{"value": "%s"}`, settingValAlwaysOnline)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, CloudFlareResponse{Success: true}),
				),
			)
			server.RouteToHandler("PATCH", fmt.Sprintf("/zones/%s/settings/browser_cache_ttl", zoneID),
				ghttp.CombineHandlers(
					ghttp.VerifyJSON(fmt.Sprintf(`{"value": %d}`, settingValBrowserCache)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, CloudFlareResponse{Success: true}),
				),
			)
			server.RouteToHandler("PATCH", fmt.Sprintf("/zones/%s/settings/non_existent_devops_hero", zoneID),
				ghttp.CombineHandlers(
					ghttp.RespondWithJSONEncoded(http.StatusBadRequest, CloudFlareResponse{
						Success: false,
						Errors: []CloudFlareError{CloudFlareError{
							Code:    1006,
							Message: "Unrecognized zone setting name",
						}},
					}),
				),
			)
		})

		It("should set two config items", func() {
			err := cloudFlare.Update(zoneID, ConfigItems{
				"always_online":     settingValAlwaysOnline,
				"browser_cache_ttl": settingValBrowserCache,
			})

			Expect(server.ReceivedRequests()).To(HaveLen(2))
			Expect(err).To(BeNil())
		})

		It("should return a public error when key is not supported by remote", func() {
			err := cloudFlare.Update(zoneID, ConfigItems{
				"non_existent_devops_hero": "always devopsing",
				"browser_cache_ttl":        settingValBrowserCache,
			})

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(err).ToNot(BeNil())

			// TODO:
			// What happens when we get a non-200 error for a given key we're
			// trying to set? What should we do and how should it be handled?
		})
	})

})
