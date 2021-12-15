package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gocolly/colly/v2"
)

const url = "https://koronavirus.gov.hu/elhunytak"
const idClass = ".views-field-field-elhunytak-sorszam"
const sexClass = ".views-field-field-elhunytak-nem"
const ageClass = ".views-field-field-elhunytak-kor"
const umcClass = ".views-field-field-elhunytak-alapbetegsegek"

var lastPageRegexp = regexp.MustCompile(
	`^/elhunytak\?page=([0-9]+)$`,
)

func getPage(page int) string {
	return fmt.Sprintf("%s?page=%d", url, page)
}

func handleRequest(r *colly.Request) {
	fmt.Println("Visiting", r.URL)
}

type victim struct {
	ID  string
	Sex string
	Age string
	UMC string // Underlying Medical Conditions
}

// Parse victim table row
func getVictim(e *colly.HTMLElement) (victim, error) {
	var v victim
	v.ID = e.DOM.Find(idClass).Text()
	v.Sex = e.DOM.Find(sexClass).Text()
	v.Age = e.DOM.Find(ageClass).Text()
	v.UMC = e.DOM.Find(umcClass).Text()
	return v, nil
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

	var victims []victim
	c.OnHTML("tbody > tr", func(e *colly.HTMLElement) {
		v, err := getVictim(e)
		if err != nil {
			fmt.Println(err)
			return
		}
		victims = append(victims, v)
	})

	if err := c.Visit(url); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(lastPage, getPage(lastPage))
	fmt.Println(victims)
}
