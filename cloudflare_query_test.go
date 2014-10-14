package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"net/http"
	"strings"
)

const headerEmail = "X-Auth-Email"
const headerKey = "X-Auth-Key"

var _ = Describe("CloudFlareQuery", func() {
	var (
		authEmail = "user@example.com"
		authKey   = "abc123"
		query     CloudFlareQuery
	)

	BeforeEach(func() {
		query = CloudFlareQuery{
			RootURL:   "https://example.com/api",
			AuthEmail: authEmail,
			AuthKey:   authKey,
		}
	})

	Describe("MakeRequest", func() {
		var (
			req *http.Request
			err error
		)

		BeforeEach(func() {
			req, err = query.NewRequest("GET", "/zones")
		})

		It("should return request and no errors", func() {
			Expect(req).ToNot(BeNil())
			Expect(err).To(BeNil())
		})

		It("should set method and path", func() {
			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.String()).To(Equal("https://example.com/api/zones"))
		})

		It("should not set request body", func() {
			Expect(req.Body).To(BeNil())
		})

		It("should set Content-Type to JSON", func() {
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("should set authentication email and key header", func() {
			Expect(req.Header.Get(headerEmail)).To(Equal(authEmail))
			Expect(req.Header.Get(headerKey)).To(Equal(authKey))
		})
	})

	Describe("MakeRequestBody", func() {
		var (
			body = `{"foo": "bar"}`
			req  *http.Request
			err  error
		)

		BeforeEach(func() {
			req, err = query.NewRequestBody(
				"PATCH", "/settings", strings.NewReader(body),
			)
		})

		It("should return request and no errors", func() {
			Expect(req).ToNot(BeNil())
			Expect(err).To(BeNil())
		})

		It("should set method and URL", func() {
			Expect(req.Method).To(Equal("PATCH"))
			Expect(req.URL.String()).To(Equal("https://example.com/api/settings"))
		})

		It("should set request body", func() {
			buf, err := ioutil.ReadAll(req.Body)
			defer req.Body.Close()

			Expect(err).To(BeNil())
			Expect(buf).To(MatchJSON(body))
		})
	})
})
