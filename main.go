package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"project/crawler"
)

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

	crawler := crawler.NewCrawler(baseURL, depth, maxConcurrency, delay, userAgent)

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
