package main

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"

	"io"
	"net/http"
	"os"

	"github.com/devincarr/goarxiv"
	"github.com/ogier/pflag"
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
	defer wg.Done()
	return downloadFile(filepath, url)
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
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

func getFullName(a article, p string) string {
	filename := fmt.Sprintf("%v.pdf", strings.TrimSpace(a.title))
 return path.Join(p, filename)
}

func downloadArticles(url string, count int, p string, parallel bool) {
	articles := getArticles(url, count)
	os.MkdirAll(p, 0755)

	if parallel {
		var wg sync.WaitGroup
		wg.Add(len(articles))
		for _, article := range articles {
			go downloadFileC(getFullName(article, p), article.url, &wg)
		}
		wg.Wait()
	} else {
		for _, article := range articles {
			downloadFile(fmt.Sprintf(getFullName(article, p), strings.TrimSpace(article.title)), article.url)
		}
	}
}

func main() {
	search := pflag.String("search", "google", "The types of articles you want to search for. 'google' by default as an example.")
	count := pflag.Int("count", 10, "The number of articles you want to retrieve. 10 by default.")
	path := pflag.String("path", "articles", "the location you want to store the articles. A folder 'articles' in the current directory by default.")
	parallel := pflag.Bool("parallel", true, "Whether or not you want to pull articles down in parallel. Default is yes.")

	pflag.Parse()

	downloadArticles(*search, *count, *path, *parallel)
}
