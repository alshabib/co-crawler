package main

import (
	"github.com/mkideal/cli"
	"fmt"
	"net/http"
	"io"
	"golang.org/x/net/html"
	"net/url"
	"strings"
	"googlemaps.github.io/maps"
	"log"
	"golang.org/x/net/context"
	"os"
	"encoding/json"
)

type zipcodes struct {
	zips []string
}

func (d *zipcodes) Decode(s string) error {
	d.zips = strings.Split(s, ",")
	return nil
}

func main() {
	type argT struct {
		cli.Helper
		Codes zipcodes `cli:"Z" usage:"Fetch Central Offices in this zipcode. (Comma seperated)"`
		Apikey string `cli:"a,apikey" usage:"APIKEY for google Geocode API"`
		Csv bool `cli:"c,csv" usage:"Output CSV for importing to Google Maps (default: false)"`
	}


	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		fetchCities(argv.Codes.zips, argv.Apikey, argv.Csv)
		return nil
	})

}

func fetchCities(zips []string, apikey string, csv bool) {
	var Url *url.URL
	Url, _ = url.Parse("http://www.sandman.com")
	Url.Path += "cosearch.asp"
	var addresses []string

	for _, city := range zips {
		parameters := url.Values{}
		parameters.Add("formType", "City")
		parameters.Add("txtZip", city)
		Url.RawQuery = parameters.Encode()
		resp, err := http.Get(Url.String())
		if err != nil {
			fmt.Println(err)
			continue
		}
		for key, _ := range parseHtml(resp.Body, csv) {
			addresses = append(addresses, key)
		}
		resp.Body.Close()
	}
	if !csv {
		fetchGeocode(addresses, apikey)
	}
}

func parseHtml(body io.ReadCloser, csv bool) map[string]int {
	var locations = make(map[string]int)
	counter := 0
	z := html.NewTokenizer(body)

	z = findTable(z)

	if z == nil {
		fmt.Println("Nothing to see here")
	}

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return nil
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "tr" {
				counter++
				if counter > 2 {
					data := parseTableReturn(z)
					if len(data) > 1 && locations[data[2]] == 0 {
						locations[data[2]]++
						if csv {
							displayCsv(data)
						}
					}
				}

			}
		case tt == html.EndTagToken:
			t := z.Token()
			if t.Data == "table" {
				return locations
			}
		}

	}

}

func check(err error) {
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
}


func fetchGeocode(addresses []string, apikey string) {
	var latlongs []string
	var client *maps.Client
	var err error

	if apikey != "" {
		client, err = maps.NewClient(maps.WithAPIKey(apikey))
		check(err)
	} else {
		println("Must pass an APIKEY!")
		os.Exit(2)
	}

	for _, address := range addresses {

		r := &maps.GeocodingRequest{
			Address: address,
		}

		resp, err := client.Geocode(context.Background(), r)
		latlongs = append(latlongs, fmt.Sprintf("%v,%v",resp[0].Geometry.Location.Lat, resp[0].Geometry.Location.Lng))
		check(err)
	}

	slcB, _ := json.Marshal(latlongs)
	fmt.Println(string(slcB))

}

func displayCsv(co []string) {
	fmt.Printf("\"%s\",\"%s\"\n", co[2], co[3])
}

func findTable(z *html.Tokenizer) *html.Tokenizer {
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return nil
		case tt == html.StartTagToken:
			t := z.Token()
			for _, a := range t.Attr {
				if a.Key == "id" && a.Val == "Table4" {
					return z
				}
			}
		}
	}

}

func parseTableReturn(z *html.Tokenizer) []string {
	var data []string
	for {
		tt := z.Next()
		switch tt {
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "td" {
				data = append(data, parseTableEntry(z))
			}
		case html.EndTagToken:
			t := z.Token()
			if t.Data == "tr" {
				return data
			}

		}
	}
}

func parseTableEntry(z *html.Tokenizer) string {
	var text string
	for {
		tt := z.Next()
		switch tt {
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "font" {
				text = parseFont(z)
			}
		case html.EndTagToken:
			t := z.Token()
			if t.Data == "td" {
				return text
			}
		}
	}
}

func parseFont(z *html.Tokenizer) string  {

	var text string
	for {
		tt := z.Next()
		switch tt {
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				return text
			}
		case html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "br" {
				continue
			} else {
				return text
			}
		case html.EndTagToken:
			t := z.Token()
			if t.Data == "font" {
				return text
			}
		case html.TextToken:
			text = text + " " + string(z.Text()[0:])
		}
	}

}
