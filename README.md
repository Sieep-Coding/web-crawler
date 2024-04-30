## See it in action

![](https://github.com/Sieep-Coding/web-crawler/blob/main/gif.gif)


<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
<h1>Go Web Crawler</h1>
<p>This is a concurrent web crawler implemented in Go.
It allows you to crawl websites, extract links, and scrape specific data from the visited pages.</p>
<h2>Features</h2>
<ul>
<li>Crawls web pages concurrently using goroutines</li>
<li>Extracts links from the visited pages</li>
<li>Scrapes data such as page title, meta description, meta keywords, headings, paragraphs, image URLs, external links, and table data from the visited pages</li>
<li>Supports configurable crawling depth</li>
<li>Handles relative and absolute URLs</li>
<li>Tracks visited URLs to avoid duplicate crawling</li>
<li>Provides timing information for the crawling process</li>
<li>Saves the extracted data in a well-formatted CSV file</li>
</ul>
<h2>Installation</h2>
<ol>
<li>Make sure you have Go installed on your system. You can download and install Go from the official website: <a href="https://golang.org">https://golang.org</a></li>
<li>Clone this repository to your local machine:
<pre><code>git clone https://github.com/sieep-coding/web-crawler.git</code></pre>
</li>
<li>Navigate to the project directory:
<pre><code>cd web-crawler</code></pre>
</li>
<li>Install the required dependencies:
<pre><code>go mod download</code></pre>
</li>
</ol>
<h2>Usage</h2>
<ol>
<li>Open a terminal and navigate to the project directory.</li>
<li>Run the following command to start the web crawler:
<pre><code>go run main.go &lt;url&gt;</code></pre>
Replace <code>&lt;url&gt;</code> with the URL you want to crawl.
</li>
<li>Wait for the crawling process to complete. The crawler will display the progress and timing information in the terminal.</li>
<li>Once the crawling is finished, the extracted data will be saved in a CSV file named <code>crawl_results.csv</code> in the project directory.</li>
</ol>
<h2>Customization</h2>
<p>You can customize the web crawler according to your needs:</p>
<ul>
<li>Modify the <code>processPage</code> function in <code>crawler/page.go</code> to extract additional data from the visited pages using the <code>goquery</code> package.</li>
<li>Extend the <code>Crawler</code> struct in <code>crawler/crawler.go</code> to include more fields for storing extracted data.</li>
<li>Customize the CSV file generation in <code>main.go</code> to match your desired format.</li>
<li>Implement rate limiting to avoid overloading the target website.</li>
<li>Add support for handling robots.txt and respecting crawling restrictions.</li>
<li>Integrate the crawler with a database or file storage to persist the extracted data.</li>
</ul>
<h2>License</h2>
<p>This project is licensed under the <a href="LICENSE">UNLICENSE</a>.</p>
</body>
</html>
