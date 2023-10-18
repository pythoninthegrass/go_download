package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

func main() {
	// load the environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// assign the environment variables
	name := os.Getenv("NAME")
	url := os.Getenv("URL")
	dir := os.Getenv("DIR")
	ext := os.Getenv("EXT")

	if name == "" {
		name = "noaa"
	}
	if url == "" {
		url = "https://www.ncei.noaa.gov/pub/data/swdi/stormevents/csvfiles/"
	}
	if dir == "" {
		dir = "data"
	}
	if ext == "" {
		ext = ".gz"
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
	}

	// fetch the response
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	defer resp.Body.Close()

	// parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// extract the links
	var links []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasSuffix(href, ext) {
			links = append(links, href)
		}
	})

	// download the files
	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()

			fileUrl := url + link
			fileName := filepath.Join(dir, link)

			// check if file already exists
			if _, err := os.Stat(fileName); err == nil {
				fmt.Println("File already exists:", fileName)
				return
			}

			// fetch the file
			resp, err := http.Get(fileUrl)
			if err != nil {
				fmt.Println("Error fetching file:", err)
				return
			}
			defer resp.Body.Close()

			// write the file
			file, err := os.Create(fileName)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = file.ReadFrom(resp.Body)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			fmt.Println("Downloaded file:", fileName)
		}(link)
	}

	wg.Wait()

	fmt.Println("Success! Exiting now...")
}
