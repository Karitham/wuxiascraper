package scrape

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"github.com/PuerkitoBio/goquery"
)

func Run(wuxiaURL string, a fyne.App, pg binding.Float, file fyne.URIWriteCloser) error {
	wg := &sync.WaitGroup{}
	ch := make(chan Chapter)

	b := Book{}

	urls, err := b.FindAllChapters(wuxiaURL)
	if err != nil {
		return err
	}

	go b.Consume(ch, pg)
	for i, url := range urls {
		time.Sleep(20 * time.Millisecond)
		wg.Add(1)

		go func(url string, i int) {
			defer wg.Done()
			c, err := scrape(url, i)
			if err != nil {
				log.Println(err)
			}
			ch <- c
		}(url, i)
	}
	wg.Wait()
	close(ch)

	return b.Write(a, file)
}

func (b *Book) FindAllChapters(u string) ([]string, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	URL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	urls := []string{}
	doc.Find("li.chapter-item > a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		nu := fmt.Sprintf("%s://%s%s", URL.Scheme, URL.Host, href)
		urls = append(urls, nu)
	})
	b.Author = doc.Find("div.novel-index > div.novel-container > div > div:nth-child(5) dd").Text()
	b.Title = doc.Find("div.novel-index > div.novel-container h2").Text()
	b.ChapterCount = len(urls)
	b.Cover = doc.Find("img.img-thumbnail").AttrOr("src", "")

	log.Printf("Found %d chapters", b.ChapterCount)

	return urls, nil
}

func (b *Book) Consume(ch chan Chapter, pg binding.Float) {
	ts := []Chapter{}
	for c := range ch {
		ts = append(ts, c)
		err := pg.Set(float64(len(ts)) / float64(b.ChapterCount))
		if err != nil {
			log.Println(err)
		}
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Number < ts[j].Number
	})

	b.Chapters = ts
}

func scrape(url string, count int) (Chapter, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Chapter{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Chapter{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Chapter{}, err
	}

	body := &bytes.Buffer{}
	doc.Find("#chapter-content > p").Each(func(i int, s *goquery.Selection) {
		h, _ := s.Html()
		body.WriteString("<p>")
		body.WriteString(h)
		body.WriteString("</p>")
	})

	ch := Chapter{
		Title:   doc.Find(".caption > div > h4").Text(),
		Content: body,
		Number:  count,
		URL:     url,
	}

	return ch, nil
}
