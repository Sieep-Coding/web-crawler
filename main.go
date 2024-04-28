package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Crawler struct {
	BaseURL string
	Depth   int
	Visited map[string]bool
	mu      sync.Mutex
	wg      sync.WaitGroup
}

func NewCrawler(baseURL string, depth int) *Crawler {
	return &Crawler{
		BaseURL: baseURL,
		Depth:   depth,
		Visited: make(map[string]bool),
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

	doc, err := html.Parse(resp.Body)
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

func (c *Crawler) processPage(pageURL string, doc *html.Node) {
	// Perform analysis or processing of the parsed HTML document
	fmt.Printf("Processing page: %s\n", pageURL)
	// ...
}

func (c *Crawler) extractLinks(doc *html.Node) []string {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link := attr.Val
					if !strings.HasPrefix(link, "http") {
						links = append(links, link)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
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
	baseURL := "https://example.com"
	depth := 2

	crawler := NewCrawler(baseURL, depth)
	start := time.Now()
	crawler.Crawl()
	elapsed := time.Since(start)

	fmt.Printf("Crawling completed in %s\n", elapsed)
}
