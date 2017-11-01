package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	pageRegexp = `\S+\s+(\w+)\s+\S+\s+(\w+)`
)

type MusicInfo struct {
	Title       string
	Description string
	Genres      string
	Artist      string
	MusicDetail MusicDetail
	DownloadUrl string
	Tags        string
}

type MusicDetail struct {
	Quality   string
	Duration  string
	Tempo     string
	AudioSize string
}

func main() {
	urls, err := readLines("./ofs_urls.txt")
	if err != nil {
		log.Fatal(err)
	}

	for _, baseUrl := range urls {
		run(baseUrl)
	}
}

func run(baseUrl string) {
	doc, err := goquery.NewDocument(baseUrl)
	if err != nil {
		log.Fatal(err)
	}

	pageMatch, err := getTotalPageWithRegexp(doc.Find(".pages").Text())
	if err != nil {
		fmt.Println("page maybe has only one or none")
		os.Exit(1)
	}

	totalPage, err := strconv.Atoi(pageMatch[2])
	if err != nil {
		fmt.Println("parse page from string to int failed")
		os.Exit(1)
	}

	pageUrls := getPageUrls(baseUrl, totalPage)
	for _, pageUrl := range pageUrls {
		downloadUrls := getDownloadUrlFromEachPage(pageUrl)
		for _, url := range downloadUrls {
			musicDetails := musicDetails(url)
			b, err := json.Marshal(musicDetails)
			if err != nil {
				fmt.Printf("url: [%s] detail struct to json failed\n", url)
				continue
			}
			println(string(b))
		}
	}
}

func getDownloadUrlFromEachPage(pageLink string) []string {
	var downloadUrls []string
	doc, err := goquery.NewDocument(pageLink)
	if err != nil {
		log.Fatal("Get downloadurl error in Function [getDownloadUrlFromEachPage]", err)
	}
	doc.Find(".post-title").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Find("a").Attr("href")
		downloadUrls = append(downloadUrls, href)
	})

	return downloadUrls
}

func getPageUrls(baseUrl string, totalPage int) []string {
	var hrefs []string
	for i := 1; i <= totalPage; i++ {
		href := fmt.Sprintf(baseUrl + "page/" + strconv.Itoa(i))
		hrefs = append(hrefs, href)
	}

	return hrefs
}

func getTotalPageWithRegexp(pageString string) ([]string, error) {
	re := regexp.MustCompile(pageRegexp)
	matches := re.FindStringSubmatch(pageString)
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not parse page string")
	}

	return matches, nil
}

func musicDetails(url string) *MusicInfo {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}
	musicInfo := &MusicInfo{}

	// Get Title
	title := doc.Find(".post-inner.group h1").Text()
	musicInfo.Title = title

	// Get downloadUrl
	re := regexp.MustCompile(".zip$")
	var downloadUrl string
	doc.Find("table tbody tr td a").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		if re.MatchString(url) {
			downloadUrl = url
		}
	})
	musicInfo.DownloadUrl = downloadUrl

	// Get Basic Info
	text := doc.Find(".entry-inner p").Text()
	re = regexp.MustCompile("Description:\\s+(.*?)\\s+Genres:\\s+(.*?)\\s+Artist:\\s+(\\w+)")
	infoes := re.FindStringSubmatch(text)
	if len(infoes) > 1 {
		musicInfo.Description = infoes[1]
		musicInfo.Genres = infoes[2]
		musicInfo.Artist = infoes[3]
	}

	// Get Music Detail
	//musicDetail := &MusicDetail{}
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		td := strings.Split(s.Find("td").Text(), ":")
		if td[0] == "Quality" {
			musicInfo.MusicDetail.Quality = td[1]
		}
		if td[0] == "Duration" {
			durationValue := td[1] + ":" + td[2]
			musicInfo.MusicDetail.Duration = durationValue
		}
		if td[0] == "Tempo" {
			musicInfo.MusicDetail.Tempo = td[1]
		}
		if td[0] == "Audio Size" {
			musicInfo.MusicDetail.AudioSize = td[1]
		}

		if len(td) < 2 {
			return
		}
	})

	// Get tags
	var tags []string
	doc.Find(".post-tags a").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, s.Text())
	})
	tagsString := strings.Join(tags, ",")
	musicInfo.Tags = tagsString

	return musicInfo
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}
