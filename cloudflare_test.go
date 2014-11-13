package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"

	"fmt"
	"log"
	"net/http"
)

var _ = Describe("CloudFlare", func() {
	var (
		server     *ghttp.Server
		query      *CloudFlareQuery
		logbuf     *gbytes.Buffer
		cloudFlare *CloudFlare
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		query = &CloudFlareQuery{RootURL: server.URL()}

		logbuf = gbytes.NewBuffer()
		logger := log.New(logbuf, "", 0)

		cloudFlare = NewCloudFlare(query, logger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("CloudFlareSettings", func() {
		Describe("ConfigItems()", func() {
			It("should return ConfigItems", func() {
				settings := CloudFlareSettings{
					CloudFlareSetting{
						ID:         "always_online",
						Value:      "off",
						ModifiedOn: "2014-07-09T11:50:56.595672Z",
						Editable:   true,
					},
					CloudFlareSetting{
						ID:         "browser_cache_ttl",
						Value:      14400,
						ModifiedOn: "2014-07-09T11:50:56.595672Z",
						Editable:   true,
					},
				}

				Expect(settings.ConfigItems()).To(Equal(ConfigItems{
					"always_online":     "off",
					"browser_cache_ttl": 14400,
				}))
			})
		})
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
				{
					ID:   "123",
					Name: "foo",
				},
				{
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

		Context("200, success: true, errors: []", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusOK, `{
							"errors": [],
							"messages": [],
							"result": [],
							"success": true
						}`),
					),
				)
			})

			It("should not return error", func() {
				resp, err := cloudFlare.MakeRequest(req)

				Expect(resp).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
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
				Expect(err).To(MatchError(`Response body indicated failure, response: main.CloudFlareResponse{Success:false, Errors:[]main.CloudFlareError{}, Messages:[]string{}, Result:json.RawMessage{0x5b, 0x5d}}`))
			})
		})

		Context("200, success: true, errors: [code: 1000, message: something bad]", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusOK, `{
							"errors": [{
								"code": 1000,
								"message": "something bad"
							}],
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
				Expect(err).To(MatchError(`Response body indicated failure, response: main.CloudFlareResponse{Success:true, Errors:[]main.CloudFlareError{main.CloudFlareError{Code:1000, Message:"something bad"}}, Messages:[]string{}, Result:json.RawMessage{0x5b, 0x5d}}`))
			})
		})

		Context("200, non-JSON body", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusOK, "something invalid"),
					),
				)
			})

			It("should return error", func() {
				resp, err := cloudFlare.MakeRequest(req)

				Expect(resp).To(BeNil())
				Expect(err).To(MatchError("invalid character 's' looking for beginning of value"))
			})
		})

		Context("500, non-JSON body", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/something"),
						ghttp.RespondWith(http.StatusServiceUnavailable, "something invalid"),
					),
				)
			})

			It("should return error", func() {
				resp, err := cloudFlare.MakeRequest(req)

				Expect(resp).To(BeNil())
				Expect(err).To(MatchError("Didn't get 200 response, body: something invalid"))
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

		It("should return two CloudFlareSettings", func() {
			settings, err := cloudFlare.Settings(zoneID)

			Expect(settings).To(Equal(CloudFlareSettings{
				CloudFlareSetting{
					ID:         "always_online",
					Value:      "off",
					ModifiedOn: "2014-07-09T11:50:56.595672Z",
					Editable:   true,
				},
				CloudFlareSetting{
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
						Errors: []CloudFlareError{{
							Code:    1006,
							Message: "Unrecognized zone setting name",
						}},
					}),
				),
			)
		})

		Context("logOnly false", func() {
			logOnly := false

			It("should set two config items and log progress", func() {
				config := ConfigItemsForUpdate{
					"always_online": ConfigItemForUpdate{
						Current:  "off",
						Expected: settingValAlwaysOnline,
					},
					"browser_cache_ttl": ConfigItemForUpdate{
						Current:  nil,
						Expected: settingValBrowserCache,
					},
				}

				err := cloudFlare.Update(zoneID, config, logOnly)

				Expect(server.ReceivedRequests()).To(HaveLen(2))
				Expect(err).To(BeNil())

				Expect(logbuf).To(gbytes.Say(fmt.Sprintf(
					`Changing setting "always_online" from "off" to "%s"`, settingValAlwaysOnline,
				)))
				Expect(logbuf).To(gbytes.Say(fmt.Sprintf(
					`Changing setting "browser_cache_ttl" from <nil> to %d`, settingValBrowserCache,
				)))
			})

			It("should return a public error when key is not supported by remote", func() {
				config := ConfigItemsForUpdate{
					"non_existent_devops_hero": ConfigItemForUpdate{
						Current:  nil,
						Expected: "always devopsing",
					},
					"browser_cache_ttl": ConfigItemForUpdate{
						Current:  nil,
						Expected: settingValBrowserCache,
					},
				}

				Expect(cloudFlare.Update(zoneID, config, false)).ToNot(BeNil())

				// TODO:
				// What happens when we get a non-200 error for a given key we're
				// trying to set? What should we do and how should it
				// be handled?
				//
				// When ranging over Go arrays order isn't guaranteed,
				// so the following assertion will flicker until we
				// can design a better test for it.
				// Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("logOnly true", func() {
			logOnly := true

			It("should log progress without setting config items", func() {
				config := ConfigItemsForUpdate{
					"always_online": ConfigItemForUpdate{
						Current:  "off",
						Expected: settingValAlwaysOnline,
					},
					"browser_cache_ttl": ConfigItemForUpdate{
						Current:  nil,
						Expected: settingValBrowserCache,
					},
				}

				err := cloudFlare.Update(zoneID, config, logOnly)

				Expect(server.ReceivedRequests()).To(HaveLen(0))
				Expect(err).To(BeNil())

				Eventually(logbuf).Should(gbytes.Say(fmt.Sprintf(
					`Would have changed setting "always_online" from "off" to "%s"`, settingValAlwaysOnline,
				)))
				Eventually(logbuf).Should(gbytes.Say(fmt.Sprintf(
					`Would have changed setting "browser_cache_ttl" from <nil> to %d`, settingValBrowserCache,
				)))
			})
		})
	})
})
