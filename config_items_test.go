package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigItems", func() {
	Describe("Union", func() {
		It("merges two ConfigItems objects overwriting values of the latter with the former", func() {
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
		It("returns the difference of two ConfigItems objects", func() {
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
