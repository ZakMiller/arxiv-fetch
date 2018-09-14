package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"io"
	"net/http"
	"os"
	"runtime"

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

func downloadFileC(filepath string, url string, wg *sync.WaitGroup) error {

	_ = os.Mkdir("articlesC", 0755)
	// Create the file
	out, err := os.Create("articlesC/" + filepath)
	if err != nil {
		wg.Done()
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		wg.Done()
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		wg.Done()
		return err
	}

	wg.Done()
	return nil
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

func downloadArticlesC(url string, count int) {
	articles := getArticles(url, count)
	var wg sync.WaitGroup
	wg.Add(len(articles))
	for _, article := range articles {
		go downloadFileC(fmt.Sprintf("%v.pdf", strings.TrimSpace(article.title)), article.url, &wg)
	}

	wg.Wait()

}

func downloadArticles(url string, count int) {
	articles := getArticles(url, count)
	for _, article := range articles {
		downloadFile(fmt.Sprintf("%v.pdf", strings.TrimSpace(article.title)), article.url)
	}
}

func main() {
	t := time.Now()
	downloadArticles("google", 10)
	fmt.Printf("Not concurrent - %v\n", time.Since(t))

	t = time.Now()
	downloadArticlesC("google", 10)
	fmt.Printf("Concurrent - %v\n", time.Since(t))

	fmt.Printf("Core count - %v\n", runtime.NumCPU())
}
