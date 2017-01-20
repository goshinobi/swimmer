package crawl

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var (
	targetHost      = map[string]*regexp.Regexp{}
	ignoreHost      = map[string]*regexp.Regexp{}
	crawlUrlPattern = regexp.MustCompile(".")
)

func SetTargetHost(host string) {
	targetHost[host] = regexp.MustCompile(host)
}

func SetIgnoreHost(host string) {
	ignoreHost[host] = regexp.MustCompile(host)
}

func SetCrawlUrlPattern(p string) {
	crawlUrlPattern = regexp.MustCompile(p)
}

func isTargetHost(u string) bool {
	urlObj, err := url.Parse(u)
	if err != nil {
		return false
	}
	if _, ok := ignoreHost[urlObj.Host]; ok {
		return false
	}
	for _, v := range ignoreHost {
		if v.MatchString(urlObj.Host) {
			return false
		}
	}
	if _, ok := targetHost[urlObj.Host]; ok {
		return true
	}
	for _, v := range targetHost {
		if v.MatchString(urlObj.Host) {
			return true
		}
	}
	return false
}

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}
var isCrawled = map[string]bool{}

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	return
}

func crawl(u string, ch chan string, chFinished chan bool) {

	time.Sleep(30 * time.Millisecond)
	defer func() {
		chFinished <- true
	}()

	if _, ok := isCrawled[u]; ok {
		return
	}
	if !isTargetHost(u) {
		return
	}

	isCrawled[u] = true
	resp, err := client.Get(u)

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + u + "\"")
		return
	}
	fmt.Println(u)

	b := resp.Body
	defer b.Close()

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			ok, u := getHref(t)
			if !ok {
				continue
			}
			re := crawlUrlPattern.Copy()
			if !re.MatchString(u) {
				continue
			}

			hasProto := strings.Index(u, "http") == 0
			if hasProto {
				ch <- u
			}
		}
	}
}

func Crawl(depth int, seedURLs ...string) map[string]bool {
	foundUrls := make(map[string]bool)
	chUrls := make(chan string)
	chFinished := make(chan bool)
	if depth == -1 {
		return foundUrls
	}

	for _, url := range seedURLs {
		go crawl(url, chUrls, chFinished)
	}

	for c := 0; c < len(seedURLs); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	close(chUrls)
	buffer := map[string]bool{}
	for url, _ := range foundUrls {
		buffer[url] = true
		for k, _ := range Crawl(depth-1, url) {
			buffer[k] = true
		}
	}
	return buffer
}
