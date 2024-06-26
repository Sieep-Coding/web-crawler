package crawler

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
	BaseURL        string
	Depth          int
	Visited        map[string]bool
	mu             sync.Mutex
	wg             sync.WaitGroup
	Data           map[string][]string
	MaxConcurrency int
	Delay          time.Duration
	UserAgent      string
}

func NewCrawler(baseURL string, depth int, maxConcurrency int, delay time.Duration, userAgent string) *Crawler {
	return &Crawler{
		BaseURL:        baseURL,
		Depth:          depth,
		Visited:        make(map[string]bool),
		Data:           make(map[string][]string),
		MaxConcurrency: maxConcurrency,
		Delay:          delay,
		UserAgent:      userAgent,
	}
}

func (c *Crawler) Crawl() {
	semaphore := make(chan struct{}, c.MaxConcurrency)
	c.crawlPage(c.BaseURL, 0, semaphore)
	c.wg.Wait()
}

func (c *Crawler) crawlPage(pageURL string, depth int, semaphore chan struct{}) {
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	c.wg.Add(1)
	defer c.wg.Done()

	c.mu.Lock()
	if c.Visited[pageURL] || depth >= c.Depth {
		c.mu.Unlock()
		return
	}
	c.Visited[pageURL] = true
	c.mu.Unlock()

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		fmt.Printf("Error creating request for URL %s: %v\n", pageURL, err)
		return
	}
	req.Header.Set("User-Agent", c.UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
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
		go func(url string) {
			time.Sleep(c.Delay)
			c.crawlPage(url, depth+1, semaphore)
		}(absoluteURL)
	}
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
