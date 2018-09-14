package main

import (
	"strings"
	"fmt"
	"strconv"

	"net/http"
	"io"
	"os"

	"github.com/devincarr/goarxiv"
	"golang.org/x/tools/blog/atom"
)

type article struct {
	title string
	url   string
}

func getArticles(term string, count int) []article {
	s := goarxiv.New()
	s.AddQuery("search_query", fmt.Sprintf("all:%v", term))
	s.AddQuery("max_results", strconv.Itoa(count))
	result, error := s.Get()
	if error != nil {
		fmt.Println(error)
	}

	articles := make([]article, count)
	for _, entry := range result.Entry {
		if article, ok := getArticle(entry); ok {
			articles = append(articles, *article)
		}
	}
	return articles
}

func getArticle(e *atom.Entry) (*article, bool) {
	for _, link := range e.Link {
		if link.Type == "application/pdf" {
			return &article{e.Title, link.Href}, true
		}
	}
	return nil, false
}

func downloadFile(filepath string, url string) error {

	_ = os.Mkdir("articles", 0755)
	// Create the file
	out, err := os.Create("articles/" + filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	 articles := getArticles("google", 10)
	 for _, article := range articles {
		 downloadFile(fmt.Sprintf("%v.pdf", strings.TrimSpace(article.title)), article.url)
	 }
}
