package main

import (
	"github.com/mkideal/cli"
	"fmt"
	"net/http"
	"io"
	"golang.org/x/net/html"
	"net/url"
)

func main() {
	type argT struct {
		Cities []string `cli:"C" usage:"Fetch Central Offices in this city."`
	}


	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		fetchCities(argv.Cities)
		return nil
	})

}

func fetchCities(cities []string) {
	var Url *url.URL
	Url, _ = url.Parse("http://www.sandman.com")
	Url.Path += "cosearch.asp"

	for _, city := range cities {
		parameters := url.Values{}
		parameters.Add("formType", "City")
		parameters.Add("txtCity", city)
		Url.RawQuery = parameters.Encode()
		resp, err := http.Get(Url.String())
		if err != nil {
			fmt.Println(err)
			continue
		}
		parseHtml(resp.Body)
		resp.Body.Close()
	}
}

func parseHtml(body io.ReadCloser) {
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
			return
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "tr" {
				counter++
				if counter > 2 {
					displayCsv(parseTableReturn(z))
				}

			}
		case tt == html.EndTagToken:
			t := z.Token()
			if t.Data == "table" {
				return
			}
		}

	}
}

func displayCsv(co []string) {
	fmt.Println(co[0], co[1], co[2], co[3])
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
