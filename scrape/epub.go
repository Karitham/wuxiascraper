package scrape

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"fyne.io/fyne/v2"
	epub "github.com/mdigger/epub3"
)

func (b *Book) Write(a fyne.App, file fyne.URIWriteCloser) error {
	writer, err := epub.New(file)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// COVER
	resp, err := http.Get(b.Cover)
	if err == nil {
		defer resp.Body.Close()

		buf := &bytes.Buffer{}
		err = CoverTmpl.Execute(buf, map[string]string{"Filename": "images/cover.jpg"})
		if err != nil {
			return err
		}
		err = writer.AddContent(buf, "text/cover.xhtml", epub.Primary)
		if err != nil {
			return err
		}

		err = writer.AddContent(resp.Body, "images/cover.jpg", epub.Media)
		if err != nil {
			return err
		}
	}

	// CHAPTERS
	var Chapters []map[string]string
	for _, c := range b.Chapters {
		buf := &bytes.Buffer{}
		err = ChapTmpl.Execute(buf, map[string]string{
			"Title": c.Title,
			"Body":  c.Content.String(),
		})
		if err != nil {
			return err
		}

		err = writer.AddContent(buf, fmt.Sprintf("text/Section-%d.xhtml", c.Number), epub.Primary)
		if err != nil {
			return err
		}

		Chapters = append(Chapters, map[string]string{"Path": fmt.Sprintf("text/Section-%d.xhtml", c.Number), "Title": c.Title})
	}

	// NAV
	buf := &bytes.Buffer{}
	err = NavTmpl.Execute(buf, Chapters)
	if err != nil {
		return err
	}
	err = writer.AddContent(buf, "text/nav.xhtml", epub.Auxiliary)
	if err != nil {
		return nil
	}

	// TOC
	buf.Reset()
	err = TOCTmpl.Execute(buf, map[string]interface{}{"Title": b.Title, "Chapters": Chapters})
	if err != nil {
		return err
	}

	err = writer.AddContent(buf, "toc.ncx", epub.Media)
	if err != nil {
		return err
	}

	writer.SetLang("en")
	writer.SetUUID(epub.NewUUID())
	writer.AddAuthors(b.Author)
	writer.AddTitle(b.Title)
	return writer.Close()
}

var ChapTmpl = template.Must(template.New("").Parse(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<title>{{.Title}}</title>
</head>
<body>
<h2>{{.Title}}</h2>
{{.Body}}
</body>
</html>`))

var CoverTmpl = template.Must(template.New("").Parse(`<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<title>Cover</title>
</head>
<body>
<div style="text-align: center; padding: 0pt; margin: 0pt;">
<img src="../{{.Filename}}" />
</div>
</body>
</html>
`))

var NavTmpl = template.Must(template.New("").Parse(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en" xml:lang="en">
<head>
  <meta charset="utf-8"/>
  <style type="text/css">
    nav#landmarks, nav#page-list { display:none; }
    ol { list-style-type: none; }
  </style>
  <title>Table of Contents</title>
</head>
<body epub:type="frontmatter">
  <nav epub:type="toc" id="toc">
    <h1>Table of Contents</h1>
    <ol>
	  {{range .}}
      <li>
        <a href="../{{.Path}}">{{.Title}}</a>
      </li>
	  {{end}}
    </ol>
  </nav>
  <nav epub:type="landmarks" id="landmarks" hidden="">
    <h1>Landmarks</h1>
    <ol>
      <li>
        <a epub:type="toc" href="#toc">Table of Contents</a>
      </li>
	  {{range .}}
      <li>
        <a epub:type="chapter" href="../{{.Path}}">Chapter</a>
      </li>
	  {{end}}
    </ol>
  </nav>
</body>
</html>`))

var TOCTmpl = template.Must(template.New("").Parse(`<?xml version="1.0" encoding="utf-8" ?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
<head>
</head>
<docTitle>
<text>{{.Title}}</text>
</docTitle>
<navMap>
{{range .Chapters}}
<navPoint>
<navLabel>
<text>{{.Title}}</text>
</navLabel>
<content src="{{.Path}}"/>
</navPoint>
{{end}}
</navMap>
</ncx>`))
