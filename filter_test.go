package main

import (
	"github.com/beevik/etree"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RSS filtering", func() {
	It("removes items by blocked creator", func() {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
    <item>
      <title>Keep me</title>
      <dc:creator>Jason Weisberger</dc:creator>
    </item>
    <item>
      <title>Remove me</title>
      <dc:creator>Boing Boing's Shop</dc:creator>
    </item>
  </channel>
</rss>`

		doc := etree.NewDocument()
		err := doc.ReadFromString(xml)
		Expect(err).NotTo(HaveOccurred())

		removed := filter(doc, FilterConfig{BlockedCreators: []string{"Boing Boing's Shop"}})
		Expect(removed).To(Equal(1))

		items := doc.FindElements("//item")
		Expect(items).To(HaveLen(1))
		Expect(items[0].FindElement("title").Text()).To(Equal("Keep me"))
	})

	It("removes items by blocked category", func() {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
    <item>
      <title>Keep me</title>
      <category>Post</category>
    </item>
    <item>
      <title>Remove me</title>
      <category>BoingBoing Shop</category>
    </item>
  </channel>
</rss>`

		doc := etree.NewDocument()
		err := doc.ReadFromString(xml)
		Expect(err).NotTo(HaveOccurred())

		removed := filter(doc, FilterConfig{BlockedCategories: []string{"BoingBoing Shop"}})
		Expect(removed).To(Equal(1))

		items := doc.FindElements("//item")
		Expect(items).To(HaveLen(1))
		Expect(items[0].FindElement("title").Text()).To(Equal("Keep me"))
	})

	It("does not remove anything when config has no blocked values", func() {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
    <item>
      <title>A</title>
      <dc:creator>X</dc:creator>
      <category>Post</category>
    </item>
  </channel>
</rss>`

		doc := etree.NewDocument()
		err := doc.ReadFromString(xml)
		Expect(err).NotTo(HaveOccurred())

		removed := filter(doc, FilterConfig{})
		Expect(removed).To(Equal(0))

		items := doc.FindElements("//item")
		Expect(items).To(HaveLen(1))
	})
})
