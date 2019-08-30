package main

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const (
	entrypoint    = "https://www.berlin.de/special/immobilien-und-wohnen/adressen/wohnungsbaugenossenschaft/"
	nextPageQuery = `a[aria-label="Weiter blättern"]`
	wbosQuery     = `article[class="block teaser place basis"]`
)

type wohnungsbaugenossenschaft struct {
	Name    string
	Address string
	Website string
}

var results = map[string]wohnungsbaugenossenschaft{}

func main() {
	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML(nextPageQuery, func(e *colly.HTMLElement) {
		e.Request.Visit(entrypoint + e.Attr("href"))
	})

	// Populate info
	c.OnHTML(wbosQuery, func(e *colly.HTMLElement) {
		name := e.ChildText("h3>a")
		addr := strings.ReplaceAll(
			strings.ReplaceAll(
				e.ChildText("p>span[class=\"address\"]"),
				"\n", ""),
			"            ", " ")
		website, err := WGwebsite(name)
		if err != nil {
			fmt.Printf("[ERROR] Website search error: %v\n", err)
		}
		results[name] = wohnungsbaugenossenschaft{
			Name:    name,
			Address: addr,
			Website: website,
		}
	})

	c.Visit(entrypoint + "?trpg=1")
	wrtr := csv.NewWriter(os.Stdout)
	wrtr.Write([]string{"nr", "name", "website", "addres"})
	var cntr int
	for _, v := range results {
		cntr++
		wrtr.Write([]string{strconv.Itoa(cntr), v.Name, v.Website, v.Address})
	}
	wrtr.Flush()
}

func WGwebsite(name string) (string, error) {
	const google = "https://www.google.com/search?q="
	escapedName := strings.ReplaceAll(name, " ", "+")
	doc, err := goquery.NewDocument(google + escapedName)
	if err != nil {
		return "", err
	}
	var rawURL string
	doc.Find("a").EachWithBreak(
		func(n int, sel *goquery.Selection) bool {
			link, ok := sel.Attr("href")
			if !ok {
				return true
			}
			if !strings.HasPrefix(link, "/url?q=") {
				return true
			}
			rawURL = strings.ReplaceAll(link, "/url?q=", "")
			return false
		},
	)
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}
