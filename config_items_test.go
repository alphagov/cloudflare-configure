package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigItems", func() {
	Describe("Compare()", func() {
		settingValAlwaysOnline := "on"
		settingValBrowserCache := 123

		It("should return nothing when local and remote are identical", func() {
			config, err := CompareConfigItems(
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
			config, err := CompareConfigItems(
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
			config, err := CompareConfigItems(
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
			config, err := CompareConfigItems(
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
})
