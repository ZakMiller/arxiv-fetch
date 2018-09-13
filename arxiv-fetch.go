package main

import (
	"fmt"
	"strconv"

	"github.com/devincarr/goarxiv"
	"golang.org/x/tools/blog/atom"
)

type article struct {
	title string
	url   string
}

func getArticles(term string, count int) []*article{
	s := goarxiv.New()
	s.AddQuery("search_query", fmt.Sprintf("all:%v", term))
	s.AddQuery("max_results", strconv.Itoa(count))
	result, error := s.Get()
	if error != nil {
		fmt.Println(error)
	}


	articles := make([]*article, count)
	for _, entry := range result.Entry {
		if article, ok := getArticle(entry); ok {
			articles = append(articles, article)
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

func main() {
	articles := getArticles("google", 10)
	for _, article := range articles {
		fmt.Println(article)
	}
}
