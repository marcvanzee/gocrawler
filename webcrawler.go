package main

/* ======= Simple webcrawler
 * by Marc van Zee (marcvanzee@gmail.com)
 *
 *
 * Input syntax:
 *
 * $ ./webcrawler --url=<url> --depth=<depth> --max_urls=<max_urls>
 *
 * <url>      The url to start crawling from (default=http://www.marcvanzee.nl)
 * <depth>    Recursive depth of the crawling (default=3)
 * <max_urls> Maximum number of urls to crawl for (default=150)
 *
 * Extension of the last "A Tour of Go" exercise: https://tour.golang.org/concurrency/9
 * HTML parsing techniques from: http://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
 *
 * - Skips over filenames such as PDF, ZIP etc.
 * - Shows titles of URLs
 * - User can choose maximum depth and maximum number of websites to crawl
 */

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

var startURL = flag.String("url", "http://www.marcvanzee.nl", "The URL to start crawling from")
var depth = flag.Int("depth", 2, "Depth of the search")
var maxURLS = flag.Int("max_urls", 150, "Maximal number of URLs to crawl")

var countCrawled = 0

// A Fetcher visit the input url and returns the urls that occur on that website
// It returns an error when it cannot read the url
type Fetcher interface {
	Fetch(url string) (urls []string, err error)
}

// The crawhistory consists of an embed Fetcher (https://soniacodes.wordpress.com/2011/10/09/a-tour-of-go-69-exercise-web-crawler/)
// And an access token for the history map
type crawlHistory struct {
	Fetcher
	mapAccess chan map[string]bool
}

// The crawl function that is called by main
func Crawl(url string, depth int, fetcher Fetcher) {

	// define the crawlhistory
	c := &crawlHistory{
		fetcher,
		make(chan map[string]bool, 1),
	}

	// the first crawler has access to the map in the crawhistory
	c.mapAccess <- map[string]bool{url: true}

	// use the done token to ensure we do not exit the program early
	done := make(chan bool, 1)

	c.Crawl(url, depth, done)

	// exit when the crawler has finished
	<-done
}

// The crawl function that is called by the original crawler
func (c *crawlHistory) Crawl(url string, depth int, done chan bool) {
	// when we are done, add the token and return so we can quit
	if depth <= 0 {
		done <- true
		return
	}

	urls, err := c.Fetch(url)

	// we don't care about error messages
	// simply ignore website that we cannot visit
	if err != nil {
		done <- true
		return
	}

	// count the urls we visit so we can wait for them to finish crawling before exiting
	count := 0
	doneHere := make(chan bool)

	// request access to the history map
	m := <-c.mapAccess

	// iterate over all urls that were in the body of the input url
	for _, u := range urls {
		if !m[u] {
			m[u] = true
			count++
			go c.Crawl(u, depth-1, doneHere)
		}
	}

	// free the access token for the history map
	c.mapAccess <- m

	// wait for all crawlers to finish
	for ; count > 0; count-- {
		<-doneHere
	}

	done <- true
}

func main() {
	flag.Parse()
	fmt.Println("====== Starting crawling...")
	fmt.Println("=== Start URL: ", *startURL)
	fmt.Println("=== Depth:     ", *depth)
	fmt.Println("=== Max URLS:  ", *maxURLS)
	fmt.Println("=== Progress (1 dot is 1 URL found): ")

	f := fetcher(make(map[string]*result, 10))
	Crawl(*startURL, *depth, f)

	fmt.Println("\n==== Finished crawling!")

	fmt.Print("Start URL:", *startURL)
	if val, ok := f[*startURL]; ok {
		fmt.Printf("(%s)", val.title)
	}
	fmt.Println()

	i := 0
	for url, result := range f {
		fmt.Printf("%v (%v)\n", url, result.title)
		for _, url2 := range result.urls {
			i++
			fmt.Printf("|-- %v\n", url2)
		}
	}

	fmt.Printf("\nCrawled %d websites, found %d unique URLs\n", len(f), i)
}

// A fetcher is a mapping from a URL to the relevant content that we have crawled
type fetcher map[string]*result

// The result stores the relevant content of a URL, which is its title and all URLs that occur in the body
type result struct {
	title string
	urls  []string
}

// A fetcher implements fetching, which performs the crawling for an input URL and returns the URLs found in the body
func (f fetcher) Fetch(url string) ([]string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("error")
	}

	title := ""
	urls := []string{}

	b := resp.Body

	defer b.Close()

	// code for HTML parsing
	// from: http://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
	// only I added the parsing of the title of the URL
	z := html.NewTokenizer(b)

	done := false
	for !done {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			done = true
			break
		case html.StartTagToken:
			t := z.Token()

			switch t.Data {
			case "a":
				ok, u := getHref(t)
				if !ok {
					continue
				}

				if strings.Index(u, "http") == 0 {
					if countCrawled > *maxURLS {
						f[url] = &result{title, urls}
						return nil, fmt.Errorf("error")
					}
					fmt.Print(".")

					if !isFile(u) {
						urls = append(urls, u)
						countCrawled++
					}
				}
			case "title":
				if ttt := z.Next(); ttt == html.TextToken {
					title = z.Token().String()
				}
			}
		}
	}

	// store the result in the fetcher
	f[url] = &result{title, urls}

	return urls, nil
}

// retrieve the URL from a <a href="..."> token.
// from http://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	return
}

func isFile(url string) bool {
	files := []string{".pdf", ".zip", ".jpeg", ".jpg", ".gif", ".png", ".doc", ".docx", ".rar", ".gzip", ".tar", ".mp3",
		".wav", ".mpg", ".mpeg", ".swf", ".exe", ".bin"}

	for _, file := range files {
		if strings.HasSuffix(url, file) {
			return true
		}
	}

	return false
}
