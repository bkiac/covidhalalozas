package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
)

const url = "https://koronavirus.gov.hu/elhunytak"
const idSelector = ".views-field-field-elhunytak-sorszam"
const sexSelector = ".views-field-field-elhunytak-nem"
const ageSelector = ".views-field-field-elhunytak-kor"
const umcSelector = ".views-field-field-elhunytak-alapbetegsegek"
const victimsCSVPath = "victims.csv"
const victimRowSelector = "tbody > tr"
const lastPageSelector = ".pager-last > a"

var pageRegexp = regexp.MustCompile(`^.*\?page=([0-9]+)$`)

func getPageURL(page int) string {
	return fmt.Sprintf("%s?page=%d", url, page)
}

func getPage(s string) (int, error) {
	match := pageRegexp.FindStringSubmatch(s)
	if len(match) == 0 {
		return 0, errors.New("page not found")
	}
	return strconv.Atoi(match[1])
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

func getVictim(e *colly.HTMLElement) victim {
	return victim{
		ID:  getVictimField(e, idSelector),
		Sex: getVictimField(e, sexSelector),
		Age: getVictimField(e, ageSelector),
		UMC: getVictimField(e, umcSelector),
	}
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
	var lastPage int
	var victims []victim

	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("[%s]: Visiting first page (%s)\n", time.Now().String(), r.URL)
	})
	c.OnHTML(lastPageSelector, func(e *colly.HTMLElement) {
		i, err := getPage(e.Attr("href"))
		if err != nil {
			fmt.Println(err)
			return
		}
		lastPage = i
	})
	handleVictimRowHTML := func(e *colly.HTMLElement) {
		victims = append(victims, getVictim(e))
	}
	c.OnHTML(victimRowSelector, handleVictimRowHTML)
	if err := c.Visit(url); err != nil {
		fmt.Println(err)
		return
	}

	c2 := c.Clone()
	q, _ := queue.New(2, &queue.InMemoryQueueStorage{MaxSize: lastPage})
	for i := 1; i <= lastPage; i++ {
		if err := q.AddURL(getPageURL(i)); err != nil {
			fmt.Println(err)
		}
	}
	c2.OnRequest(func(r *colly.Request) {
		p, err := getPage(r.URL.String())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(
			"[%s]: Visiting page %d/%d (%s)\n",
			time.Now().String(),
			p+1,
			lastPage+1,
			r.URL,
		)
	})
	c2.OnHTML(victimRowSelector, handleVictimRowHTML)
	if err := q.Run(c2); err != nil {
		fmt.Println(err)
		return
	}

	sort.Slice(
		victims,
		func(i, j int) bool {
			in, err := strconv.Atoi(victims[i].ID)
			ij, err2 := strconv.Atoi(victims[j].ID)
			if err != nil || err2 != nil {
				return false
			}
			return in > ij
		},
	)
	if err := writeCSV(victims); err != nil {
		fmt.Println(err)
	}
}
