package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

const url = "https://koronavirus.gov.hu/elhunytak"
const idClass = ".views-field-field-elhunytak-sorszam"
const sexClass = ".views-field-field-elhunytak-nem"
const ageClass = ".views-field-field-elhunytak-kor"
const umcClass = ".views-field-field-elhunytak-alapbetegsegek"
const victimsCSVPath = "victims.csv"

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

func getVictimField(e *colly.HTMLElement, selector string) string {
	return strings.TrimSpace(e.DOM.Find(selector).Text())
}

// Parse victim table row
func getVictim(e *colly.HTMLElement) (victim, error) {
	var v victim
	v.ID = getVictimField(e, idClass)
	v.Sex = getVictimField(e, sexClass)
	v.Age = getVictimField(e, ageClass)
	v.UMC = getVictimField(e, umcClass)
	return v, nil
}

func victimsToData(victims []victim) [][]string {
	var data [][]string
	for _, v := range victims {
		var row []string
		row = append(row, v.ID)
		row = append(row, v.Sex)
		row = append(row, v.Age)
		row = append(row, v.UMC)
		data = append(data, row)
	}
	fmt.Println(data)
	return data
}

func writeCSV(victims []victim) error {
	f, err := os.Create(victimsCSVPath)
	if err != nil {
		return err
	}
	var data = victimsToData(victims)
	w := csv.NewWriter(f)
	err = w.WriteAll(data)
	if err != nil {
		return err
	}
	return nil
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

	if err := writeCSV(victims); err != nil {
		fmt.Println(err)
	}
}
