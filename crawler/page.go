package crawler

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Crawler) processPage(pageURL string, doc *goquery.Document) {
	title := doc.Find("title").Text()
	headings := doc.Find("h1, h2, h3").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})
	paragraphs := doc.Find("p").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})
	metaDescription := doc.Find("meta[name='description']").AttrOr("content", "")
	metaKeywords := doc.Find("meta[name='keywords']").AttrOr("content", "")
	imageURLs := []string{}
	doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			absoluteURL := c.resolveURL(pageURL, src)
			imageURLs = append(imageURLs, absoluteURL)
		}
	})
	externalLinks := []string{}
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasPrefix(href, "http") && !strings.Contains(href, c.BaseURL) {
			externalLinks = append(externalLinks, href)
		}
	})
	tableData := [][]string{}
	doc.Find("table").Each(func(i int, tableSelection *goquery.Selection) {
		tableRows := [][]string{}
		tableSelection.Find("tr").Each(func(j int, rowSelection *goquery.Selection) {
			row := []string{}
			rowSelection.Find("th, td").Each(func(k int, cellSelection *goquery.Selection) {
				row = append(row, cellSelection.Text())
			})
			tableRows = append(tableRows, row)
		})
		tableData = append(tableData, tableRows...)
	})

	c.mu.Lock()
	c.Data[pageURL] = append(c.Data[pageURL], title)
	c.Data[pageURL] = append(c.Data[pageURL], metaDescription)
	c.Data[pageURL] = append(c.Data[pageURL], metaKeywords)
	c.Data[pageURL] = append(c.Data[pageURL], headings...)
	c.Data[pageURL] = append(c.Data[pageURL], paragraphs...)
	c.Data[pageURL] = append(c.Data[pageURL], strings.Join(imageURLs, ","))
	c.Data[pageURL] = append(c.Data[pageURL], strings.Join(externalLinks, ","))
	c.Data[pageURL] = append(c.Data[pageURL], fmt.Sprintf("%v", tableData))
	c.mu.Unlock()

	fmt.Printf("Processing page: %s\n", pageURL)
}
