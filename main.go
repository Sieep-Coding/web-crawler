package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	if len(os.Args) < 2 {
		fmt.Println("Please provide a URL as a command line argument.")
		os.Exit(1)
	}

	baseURL := os.Args[1]
	depth := 3
	maxConcurrency := 10
	delay := 500 * time.Millisecond
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36"

	crawler := NewCrawler(baseURL, depth, maxConcurrency, delay, userAgent)

	start := time.Now()
	crawler.Crawl()
	elapsed := time.Since(start)

	fmt.Printf("Crawling completed in %s\n", elapsed)

	csvFile, err := os.Create("crawl_results.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		os.Exit(1)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	headers := []string{"URL", "Title", "Meta Description", "Meta Keywords", "Headings", "Paragraphs", "Image URLs", "External Links", "Table Data"}
	writer.Write(headers)

	for url, data := range crawler.Data {
		row := []string{url}
		row = append(row, data...)
		writer.Write(row)
	}

	fmt.Println("Results saved to crawl_results.csv")
}
