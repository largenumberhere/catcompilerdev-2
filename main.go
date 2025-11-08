package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sbabiv/xml2map"
)

func forceRead(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}

	return string(file)
}

func loadPages(templates_folder string, temp_folder string, posts *[]Post) Pages {
	var pages_out Pages

	pages_out.index = forceRead(templates_folder + "/index.html")
	pages_out.not_found = forceRead(templates_folder + "/404.html")
	pages_out.posts = make(map[int]string, 0)

	dirs, err := os.ReadDir(temp_folder)
	if err != nil {
		log.Fatal("error:", err)
	}

	for _, dir := range dirs {
		fileName := dir.Name()
		fullName := temp_folder + "/" + dir.Name()
		id, err := strconv.Atoi(fileName[0 : len(fileName)-5])
		if err != nil {
			log.Panic(err)
		}

		pages_out.posts[id] = forceRead(fullName)

	}

	loadIndex(&pages_out, posts, templates_folder)

	return pages_out
}

func generate_legacy_posts(legacyPosts []Post, outFolder string, template string) error {

	for _, legacypost := range legacyPosts {
		file_path := outFolder + "/" + strconv.Itoa(legacypost.id) + ".html"
		f, err := os.Create(file_path)
		if err != nil {
			return err
		}

		templaten := strings.ReplaceAll(template, "{{title}}", legacypost.title)
		templaten = strings.ReplaceAll(templaten, "{{post}}", legacypost.content)
		fmt.Fprint(f, templaten)

	}

	return nil
}

func parseItems(xml_data string) ([]map[string]interface{}, error) {
	decoder := xml2map.NewDecoder(strings.NewReader(xml_data))
	result, err := decoder.Decode()

	if err != nil {
		return nil, err
	}

	result, ok := result["rss"].(map[string]interface{})
	if !ok {
		return nil, errors.New("xml string does not contain the key ['rss']")
	}

	result, ok = result["channel"].(map[string]interface{})
	if !ok {
		return nil, errors.New("xml string does not contain the key ['rss']['channel']")
	}

	results, ok := result["item"].([]map[string]interface{})

	if !ok {
		return nil, errors.New("xml string does not contain the key ['rss']['channel']['item']")
	}

	return results, nil
}

func parseLegacyPosts(file_path string, posts *[]Post) error {
	data, err := os.ReadFile(file_path)
	if err != nil {
		return err
	}

	results, err := parseItems(string(data))
	if err != nil {
		return err
	}

	// contents := make([]Post, 0, len(results))
	for _, item := range results {
		post_type := item["1.2:post_type"]
		if post_type != "post" {
			continue
		}

		content := item["content:encoded"].(string)
		title := item["title"].(string)

		post_date := item["1.2:post_date"].(string)
		date, err := time.Parse(time.DateTime, post_date)
		if err != nil {
			return err
		}

		wp_id := item["1.2:post_id"].(string)
		id, err := strconv.Atoi(wp_id)
		if err != nil {
			return err
		}

		page := Post{
			title:     title,
			content:   content,
			timestamp: date,
			id:        id,
		}

		(*posts) = append((*posts), page)
		// contents = append(contents, page)
	}

	return nil
	// return contents, nil
}

type Post struct {
	title     string
	content   string
	timestamp time.Time
	id        int
}

type Pages struct {
	index     string
	not_found string
	posts     map[int]string
}

func loadIndex(pages *Pages, posts *[]Post, templates_folder string) {
	index := forceRead(templates_folder + "/index.html")
	posts_list := "<ul>"
	for _, post := range *posts {
		posts_list = posts_list + "<a href='/" + strconv.Itoa(post.id) + "'> <li>" + post.title + "</li></a>"
	}
	posts_list = posts_list + "</ul>"

	index = strings.ReplaceAll(index, "{{posts_list}}", posts_list)
	pages.index = index
}

func makeHandler(pages Pages) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.EscapedPath()
		if slug == "/favicon.ico" {
			// TODO
			fmt.Fprint(w, "unimplemented")
			return
		}

		if slug == "/" || slug == "" {
			// index
			fmt.Fprint(w, pages.index)
			return
		}

		id, err := strconv.Atoi(slug[1:])
		if err != nil {
			fmt.Print(err)
			fmt.Fprint(w, "unexpected url")
			return
		}

		post, ok := pages.posts[id]
		if !ok {
			fmt.Fprint(w, "invalid page")
		}

		fmt.Fprint(w, post)
	}
}

func main() {

	var err error
	posts := make([]Post, 0)
	err = parseLegacyPosts("posts/legacy/wp_blog_2025-10-09.xml", &posts)
	if err != nil {
		log.Fatal(err)
	}

	legacy_posts_template := forceRead("html/post_template1.html")
	err = generate_legacy_posts(posts, ".tmp", legacy_posts_template)
	if err != nil {
		log.Fatal("error", err)
	}

	pages := loadPages("html", ".tmp", &posts)

	handler := makeHandler(pages)
	http.HandleFunc("/", handler)
	err = http.ListenAndServe("localhost:10000", nil)
	if err != nil {
		log.Panic(err)
	}
}
