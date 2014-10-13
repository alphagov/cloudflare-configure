package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigItems", func() {
	Describe("Union", func() {
		It("two ConfigItems", func() {
			Expect(UnionConfigItems(
				ConfigItems{
					"always_online": "off",
					"browser_check": "off",
				},
				ConfigItems{
					"always_online":     "on",
					"browser_cache_ttl": 14400,
				},
			)).To(Equal(
				ConfigItems{
					"always_online":     "on",
					"browser_check":     "off",
					"browser_cache_ttl": 14400,
				},
			))
		})
	})

	Describe("Difference", func() {
		It("two ConfigItems", func() {
			Expect(DifferenceConfigItems(
				ConfigItems{
					"always_online": "off",
					"browser_check": "off",
				},
				ConfigItems{
					"always_online":     "on",
					"browser_check":     "off",
					"browser_cache_ttl": 14400,
				},
			)).To(Equal(
				ConfigItems{
					"always_online":     "on",
					"browser_cache_ttl": 14400,
				},
			))
		})
	})
})
