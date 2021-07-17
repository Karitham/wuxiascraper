package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bmaupin/go-epub"
)

type Book struct {
	Title        string
	Author       string
	Cover        string
	Chapters     []Chapter
	ChapterCount int
}

type Chapter struct {
	Title   string
	Content string
	URL     string
	Number  int
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected novel url")
	}
	if err := Run(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func Run(wuxiaURL string) error {
	wg := &sync.WaitGroup{}
	ch := make(chan Chapter)

	b := Book{}

	urls, err := b.FindAllChapters(wuxiaURL)
	if err != nil {
		return err
	}

	go b.Consume(ch)
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
			log.Printf("scraped chapter %s", url)
		}(url, i)
	}
	wg.Wait()
	close(ch)

	return b.Write()
}

func (b *Book) Write() error {
	e := epub.NewEpub(b.Title)
	e.SetAuthor(b.Author)
	s, err := e.AddImage(b.Cover, b.Title+".jpg")
	if err != nil {
		return err
	}
	e.SetCover(s, "")

	for _, c := range b.Chapters {
		_, err = e.AddSection(c.Content, c.Title, "", "")
		if err != nil {
			log.Println(err)
		}
	}

	err = e.Write(b.Title + ".epub")
	if err != nil {
		return err
	}

	return nil
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

func (b *Book) Consume(ch chan Chapter) {
	ts := []Chapter{}
	for c := range ch {
		ts = append(ts, c)
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Number < ts[j].Number
	})

	b.Chapters = ts
	b.ChapterCount = len(ts)
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

	content := strings.Builder{}
	doc.Find("#chapter-content > p").Each(func(i int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			log.Println(err)
		}
		content.WriteString("<p>")
		content.WriteString(html)
		content.WriteString("</p>")
	})

	ch := Chapter{
		Title:   doc.Find(".caption > div > h4").Text(),
		Content: content.String(),
		Number:  count,
		URL:     url,
	}

	return ch, nil
}
