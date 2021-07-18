package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	wg "fyne.io/fyne/v2/widget"
	"github.com/Karitham/wuxiascraper/scrape"
)

func main() {
	a := app.New()
	wd := a.NewWindow("WuxiaScraper")
	wd.Resize(fyne.NewSize(600, 400))

	urlStr := binding.NewString()
	entry := wg.NewEntry()
	entry.Bind(urlStr)
	entry.Resize(fyne.NewSize(200, 1))

	urlbox := container.NewPadded(container.NewVBox(
		entry,

		wg.NewButton("Get", func() {
			u, _ := urlStr.Get()
			nu, _ := url.Parse(u)
			if !strings.Contains(nu.Host, "wuxiaworld.com") {
				dialog.NewInformation("error", "not a valid wuxiaworld novel URL", wd).Show()
				return
			}

			pg := binding.NewFloat()
			pgdialog := dialog.NewCustom("Progress", "OK", wg.NewProgressBarWithData(pg), wd)
			pgdialog.Resize(fyne.NewSize(200, 50))

			fs := dialog.NewFileSave(func(uc fyne.URIWriteCloser, e error) {
				pgdialog.Show()

				err := scrape.Run(u, a, pg, uc)
				if err != nil {
					dialog.NewInformation("error", err.Error(), wd).Show()
				}
				pgdialog.Hide()
				dialog.NewInformation("done", fmt.Sprintf("downloaded %s to %s", filepath.Base(u), uc.URI().String()), wd).Show()
			},
				wd)

			fs.SetFileName(filepath.Base(nu.Path) + ".epub")
			fs.SetDismissText("Cancel")
			fs.Show()
		})),
	)

	urlbox.Resize(fyne.NewSize(400, 25))

	wd.SetContent(container.NewVBox(
		container.New(layout.NewCenterLayout(), wg.NewLabel("Enter the URL of the novel you want")),
		urlbox,
	),
	)

	wd.ShowAndRun()
}
