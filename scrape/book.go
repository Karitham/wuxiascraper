package scrape

import "bytes"

type Book struct {
	Title        string
	Author       string
	Cover        string
	Chapters     []Chapter
	ChapterCount int
}

type Chapter struct {
	Title   string
	Content *bytes.Buffer
	URL     string
	Number  int
}
