package yomiuri

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
	"github.com/moovweb/gokogiri/xpath"
)

type Yomiuri struct {
	Title string    `json:"title"`
	URL   string    `json:"url"`
	Time  time.Time `json:"time"`
	Photo bool      `json:"photo"`
}

func Test() {
	fmt.Println(GetYOmiuriNewArrivals())

}

func getNewsInfo(doc *html.HtmlDocument) ([]xml.Node, error) {
	xp := "//body/div/div/div/div/div/div/div/div/div/div/div/div/div/div/ul/li"
	xps := xpath.Compile(xp)
	newDatas, err := doc.Root().Search(xps)
	if err != nil {
		return nil, err
	}
	return newDatas, nil
}

func urlSeparation(newDatas []xml.Node) ([]string, error) {
	var ret []string

	a := xpath.Compile("./a/@href")

	for _, newData := range newDatas {
		urls, err := newData.Search(a)
		if err != nil {
			return nil, err
		}
		ret = append(ret, urls[0].Content())
	}
	return ret, nil
}

func photoSeparation(newDatas []xml.Node) ([]bool, error) {
	var ret []bool
	a := xpath.Compile("./a/span[@class='icon-photo']")
	for _, newData := range newDatas {
		icons, err := newData.Search(a)
		if err != nil {
			return nil, err
		}
		if len(icons) == 0 {
			ret = append(ret, false)
		} else {
			ret = append(ret, true)
		}
	}
	return ret, nil
}

func timeSeparation(newDatas []xml.Node) ([]time.Time, error) {
	var ret []time.Time

	a := xpath.Compile("./a/span/span")
	for _, newData := range newDatas {
		times, err := newData.Search(a)
		if err != nil {
			return nil, err
		}
		timeStr := strings.Trim(strings.Trim(times[0].Content(), "（"), "）")
		var year int
		var month int
		var day int
		var hour int
		var minute int

		fmt.Sscanf(timeStr, "%d年%d月%d日 %d時%d分", &year, &month, &day, &hour, &minute)
		loc, _ := time.LoadLocation("Asia/Tokyo")
		time := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)
		ret = append(ret, time)
	}

	return ret, nil
}

func titleSeparation(newDatas []xml.Node) ([]string, error) {
	var ret []string

	a := xpath.Compile("./a/span")
	for _, newData := range newDatas {
		titles, err := newData.Search(a)
		if err != nil {
			return nil, err
		}
		for _, title := range titles {
			newsAndTime := title.Content()
			timePath := xpath.Compile("./span")
			time, err := title.Search(timePath)
			if err != nil {
				return nil, err
			}
			if len(time) != 0 {
				cutstr := time[0].Content()
				ret = append(ret, strings.Trim(newsAndTime, cutstr))
			}
		}
	}

	return ret, nil
}

func GetYOmiuriNewArrivals() ([]Yomiuri, error) {
	var ret []Yomiuri
	page, err := getYomiuriNewArrivalsPage()
	if err != nil {
		return nil, err
	}

	doc, err := gokogiri.ParseHtml(page)
	if err != nil {
		return nil, err
	}
	defer doc.Free()
	datas, err := getNewsInfo(doc)
	if err != nil {
		return nil, err
	}
	titles, err := titleSeparation(datas)
	if err != nil {
		return nil, err
	}
	urls, err := urlSeparation(datas)
	if err != nil {
		return nil, err
	}
	times, err := timeSeparation(datas)
	if err != nil {
		return nil, err
	}
	phots, err := photoSeparation(datas)
	if err != nil {
		return nil, err
	}

	l := len(datas)
	for i := 0; i < l; i++ {
		y := Yomiuri{titles[i], urls[i], times[i], phots[i]}
		ret = append(ret, y)
	}

	return ret, nil
}

func getYomiuriNewArrivalsPage() ([]byte, error) {
	sourceURL := "http://www.yomiuri.co.jp/latestnews/"
	resp, err := http.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return page, nil
}
