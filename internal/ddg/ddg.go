package ddg

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Item struct {
	Index   int
	Snippet string
	URL     string
}

type Result struct {
	Items []*Item
	Date  string
}

var (
	limitChan = make(chan struct{}, 10)
)

var tmpl = template.Must(template.New("ddg").Parse(`DDGSearch results:
{{range .Items}}
[{{.Index}}] "{{.Snippet}}"
URL: {{.URL}}
{{end}}

Current date: {{.Date}}`))

func (r Result) Text() (string, error) {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, r)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func Search(ctx context.Context, q string, count int) (*Result, error) {
	limitChan <- struct{}{}
	defer func() { <-limitChan }()

	req, _ := http.NewRequest("GET", "https://html.duckduckgo.com/html/?q="+q, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0")
	req.Header.Set("Host", "html.duckduckgo.com")

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	r, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Result{Date: time.Now().Format("2006-01-02")}
	r.Find("div.result").EachWithBreak(func(i int, s *goquery.Selection) bool {
		item := &Item{
			Index:   i + 1,
			Snippet: s.Find("a.result__snippet").Text(),
			URL:     s.Find("a.result__a").AttrOr("href", ""),
		}
		item.Snippet = strings.ReplaceAll(item.Snippet, `"`, "")
		result.Items = append(result.Items, item)
		return len(result.Items) < count
	})

	return result, nil
}
