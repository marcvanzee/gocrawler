# gocrawler

### Simple Webcrawler written in Go

- Crawls a given URLs for new URLS

- Recursively crawls URLs that are found until a certain depth or a maximum number of URLs visited

- Skips over filenames such as PDF, ZIP etc.

- Shows titles of URLs

- Extension of the last ["A Tour of Go" exercise](https://tour.golang.org/concurrency/9)

- HTML parsing techniques from: http://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html

### Input syntax:

```$ ./webcrawler --url=<url> --depth=<depth> --max_urls=<max_urls>```

```<url>``` The url to start crawling from (default=http://www.marcvanzee.nl)

```<depth>``` Recursive depth of the crawling (default=2)

```<max_urls>``` Maximum number of urls to crawl for (default=150)

### Installing

Requires the ```golang.org/x/net/html``` package from the [golang subrepositories](https://github.com/golang/go/wiki/SubRepositories). 
Install as follows to get all ```net``` packages (including ```html```):

```go get golang.org/x/net/...```

Get this package as follows:

```go get github.com/marcvanzee/gocrawler```

### Examples

```
$./webcrawler --url=http://www.golang.org --depth=2 --max-urls=200
$./webcrawler --url=http://www.golang.org --max-urls=1000
$./webcrawler --url=http://www.golang.org --depth=1
```
