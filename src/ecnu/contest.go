package ecnu

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/feeds"
	"net/http"
)

func BuildFeeds(c *gin.Context) {
	host := "https://acm.ecnu.edu.cn"
	uri := host + "/contest/"

	// Request the HTML page.
	res, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		panic(fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}

	feed := &feeds.Feed{
		Title:       "ECNU ACM 公开比赛",
		Link:        &feeds.Link{Href: uri},
		Description: "ECNU ACM 公开比赛",
		Author:      &feeds.Author{Email: "meowhuang@163.com"},
	}

	// Find the review items
	doc.Find("div > div > table > tbody > tr").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		if icon := s.Find("i[class*='green']"); icon.Nodes == nil {
			return
		}
		title := s.Find("a:nth-child(1)")
		description := fmt.Sprintf("Title: %s\nDate: %s\nLength: %s",
			title.Text(),
			s.Find("td:nth-child(2)").Text(),
			s.Find("td:nth-child(3)").Text())

		feed.Items = append(feed.Items, &feeds.Item{
			Id:          uuid.NewMD5(uuid.NameSpaceOID, []byte(description)).String(),
			Title:       title.Text(),
			Link:        &feeds.Link{Href: host + title.AttrOr("href", "")},
			Description: description,
			Content:     description,
		})
	})

	rss, err := feed.ToRss()
	if err != nil {
		panic(err)
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(200, rss)
}
