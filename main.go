package main

import (
	"fmt"

	"github.com/goshinobi/swimmer/crawl"
)

func main() {
	crawl.SetTargetHost("example.co.jp")
	ret := crawl.Crawl(1, "https://example.co.jp")
	fmt.Println(ret)
}
