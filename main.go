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

// type Feed struct {
// 	XMLName xml.Name "xml:"
// }

// type Channel struct {
// 	Item string
// }

var pages *Pages

func forceRead(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}

	return string(file)
}

func loadPages() {
	pages = new(Pages)
	pages.index = forceRead("html/index.html")
	pages.not_found = forceRead("html/404.html")
	pages.posts = make(map[int]string, 0)
	folder := "./.tmp"
	dirs, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal("error:", err)
	}

	for _, dir := range dirs {
		fileName := dir.Name()
		fullName := folder + "/" + dir.Name()
		id, err := strconv.Atoi(fileName[0 : len(fileName)-5])
		if err != nil {
			log.Panic(err)
		}

		pages.posts[id] = forceRead(fullName)

	}

}

func writeLegacyPosts(legacyPosts []Post, outFolder string, template string) error {
	// template_parts := strings.Split(template, "{{post}}")
	// if len(template_parts) != 2 {
	// 	return errors.New("bad template, expected {{post}} got " + strconv.Itoa(len(template_parts)) + "parts")
	// }

	for _, legacypost := range legacyPosts {
		file_path := outFolder + "/" + strconv.Itoa(legacypost.id) + ".html"
		f, err := os.Create(file_path)
		if err != nil {
			return err
		}

		// fmt.Fprint(f, "<head>")
		// fmt.Fprint(f, "<title>")
		// fmt.Fprint(f, legacypost.title)
		// fmt.Fprint(f, "</title>")
		// fmt.Fprint(f, "</head>")

		// fmt.Fprint(f, template_parts[0])
		// fmt.Fprint(f, "<body>")
		templaten := strings.ReplaceAll(template, "{{title}}", legacypost.title)
		templaten = strings.ReplaceAll(templaten, "{{post}}", legacypost.content)
		fmt.Fprint(f, templaten)
		// fmt.Fprint(f, "<h1>")
		// fmt.Fprint(f, legacypost.title)
		// fmt.Fprint(f, "</h1>")
		// fmt.Fprint(f, "<p>")
		// fmt.Fprint(f, legacypost.content)
		// fmt.Fprint(f, "</p>")
		// fmt.Fprint(f, template_parts[1])
		// fmt.Fprint(f, "</body>")
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	/*
		if r.Method == "" || r.Method == "GET" {
			fmt.Fprint(w, "hello")
			return
		}
	*/

	// posts, err := parsePosts("wp_blog_2025-10-09.xml")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	slug := r.URL.EscapedPath()
	if slug == "/favicon.ico" {
		// TODO
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

	// log.Fatal("unimpmeneted page id", id)

	post, ok := pages.posts[id]
	if !ok {
		fmt.Fprint(w, "invalid page")
	}

	fmt.Fprint(w, post)
	return

	// if slug == "/" || slug == "" {
	// 	fmt.Fprintf(w, "<head>")
	// 	fmt.Fprint(w, "</head>")
	// 	fmt.Fprint(w, "<body>")
	// 	fmt.Fprintf(w, "<ul>")
	// 	for _, post := range legacy_posts {
	// 		fmt.Fprint(w, "<li>")
	// 		fmt.Fprintf(w, "<a href=\"/%s\">", strconv.Itoa(post.id))
	// 		fmt.Fprint(w, post.title)
	// 		fmt.Fprint(w, "</a>")
	// 		fmt.Fprint(w, "</li>")
	// 	}
	// 	fmt.Fprintf(w, "</ul>")
	// 	fmt.Fprint(w, "</body>")
	// 	return
	// }

	// idx := slices.IndexFunc(legacy_posts, func(a Post) bool {
	// 	return a.id == id
	// })

	// if idx < 0 {
	// 	fmt.Fprint(w, "unexpected url")
	// 	return
	// }

	// post := legacy_posts[idx]
	// fmt.Println("displaying ", post.title)

	// fmt.Fprint(w, "<head>")
	// fmt.Fprint(w, "<title>")
	// fmt.Fprint(w, post.title)
	// fmt.Fprint(w, "</title>")
	// fmt.Fprint(w, "</head>")

	// fmt.Fprint(w, "<body>")
	// fmt.Fprint(w, "<h1>")
	// fmt.Fprint(w, post.title)
	// fmt.Fprint(w, "</h1>")
	// fmt.Fprint(w, "<p>")
	// fmt.Fprint(w, post.content)
	// fmt.Fprint(w, "</p>")
	// fmt.Fprint(w, "</body>")
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

func parsePosts(file_path string) ([]Post, error) {
	data, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	results, err := parseItems(string(data))
	if err != nil {
		return nil, err
	}

	contents := make([]Post, 0, len(results))
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
			return contents, err
		}

		wp_id := item["1.2:post_id"].(string)
		id, err := strconv.Atoi(wp_id)
		if err != nil {
			return contents, err
		}

		page := Post{
			title:     title,
			content:   content,
			timestamp: date,
			id:        id,
		}

		contents = append(contents, page)
	}

	return contents, nil
}

type Post struct {
	title     string
	content   string
	timestamp time.Time
	id        int
}

var legacy_posts []Post

type Pages struct {
	index     string
	not_found string
	posts     map[int]string
}

func writeIndex() {
	index := forceRead("html/index.html")
	posts_list := "<ul>"
	for _, post := range legacy_posts {
		posts_list = posts_list + "<a href='/" + strconv.Itoa(post.id) + "'> <li>" + post.title + "</li></a>"
	}
	posts_list = posts_list + "</ul>"

	index = strings.ReplaceAll(index, "{{posts_list}}", posts_list)
	pages.index = index
}

func main() {

	// load pages into memory

	var err error
	legacy_posts, err = parsePosts("posts/legacy/wp_blog_2025-10-09.xml")
	if err != nil {
		log.Fatal(err)
	}

	template := forceRead("html/post_template1.html")
	err = writeLegacyPosts(legacy_posts, ".tmp", template)
	if err != nil {
		log.Fatal("error", err)
	}

	loadPages()
	writeIndex()

	http.HandleFunc("/", handler)
	err = http.ListenAndServe("localhost:10000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
