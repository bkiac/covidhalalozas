package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gocolly/colly/v2"
)

const url = "https://koronavirus.gov.hu/elhunytak"

var lastPageRegexp = regexp.MustCompile(
	`^/elhunytak\?page=([0-9]+)$`,
)

func getPage(page int) string {
	return fmt.Sprintf("%s?page=%d", url, page)
}

func handleRequest(r *colly.Request) {
	fmt.Println("Visiting", r.URL)
}

func main() {
	c := colly.NewCollector()
	c2 := c.Clone()

	c.OnRequest(handleRequest)
	c2.OnRequest(handleRequest)

	var lastPage int
	c.OnHTML(".pager-last > a", func(e *colly.HTMLElement) {
		i, err := strconv.Atoi(lastPageRegexp.FindStringSubmatch(e.Attr("href"))[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		lastPage = i
	})

	if err := c.Visit(url); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(lastPage, getPage(lastPage))
}
