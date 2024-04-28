package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	BaseURL string
	Depth   int
	Visited map[string]bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	Data    map[string][]string
}

func NewCrawler(baseURL string, depth int) *Crawler {
	return &Crawler{
		BaseURL: baseURL,
		Depth:   depth,
		Visited: make(map[string]bool),
		Data:    make(map[string][]string),
	}
}

func (c *Crawler) Crawl() {
	c.crawlPage(c.BaseURL, 0)
	c.wg.Wait()
}

func (c *Crawler) crawlPage(pageURL string, depth int) {
	c.wg.Add(1)
	defer c.wg.Done()

	c.mu.Lock()
	if c.Visited[pageURL] || depth >= c.Depth {
		c.mu.Unlock()
		return
	}
	c.Visited[pageURL] = true
	c.mu.Unlock()

	resp, err := http.Get(pageURL)
	if err != nil {
		fmt.Printf("Error fetching URL %s: %v\n", pageURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching URL %s: %s\n", pageURL, resp.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing HTML from URL %s: %v\n", pageURL, err)
		return
	}

	c.processPage(pageURL, doc)

	links := c.extractLinks(doc)
	for _, link := range links {
		absoluteURL := c.resolveURL(pageURL, link)
		go c.crawlPage(absoluteURL, depth+1)
	}
}

func (c *Crawler) processPage(pageURL string, doc *goquery.Document) {
	// Perform data extraction and processing using goquery selectors
	title := doc.Find("title").Text()
	headings := doc.Find("h1, h2, h3").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})

	c.mu.Lock()
	c.Data[pageURL] = append(c.Data[pageURL], title)
	c.Data[pageURL] = append(c.Data[pageURL], headings...)
	c.mu.Unlock()

	fmt.Printf("Processing page: %s\n", pageURL)
}

func (c *Crawler) extractLinks(doc *goquery.Document) []string {
	var links []string
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && !strings.HasPrefix(href, "http") {
			links = append(links, href)
		}
	})
	return links
}

func (c *Crawler) resolveURL(baseURL, href string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return base.ResolveReference(uri).String()
}

func main() {
	baseURL := "https://coachtony.medium.com/" //example medium page you can scrape :)
	depth := 2

	crawler := NewCrawler(baseURL, depth)
	start := time.Now()
	crawler.Crawl()
	elapsed := time.Since(start)

	fmt.Printf("Crawling completed in %s\n", elapsed)

	// Print the extracted data
	for url, data := range crawler.Data {
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("Title: %s\n", data[0])
		fmt.Printf("Headings: %v\n", data[1:])
		fmt.Println("------------------------")
	}
}
