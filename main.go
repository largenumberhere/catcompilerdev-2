package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
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

func handler(w http.ResponseWriter, r *http.Request) {
	/*
		if r.Method == "" || r.Method == "GET" {
			fmt.Fprint(w, "hello")
			return
		}
	*/

	posts, err := parsePosts("wp_blog_2025-10-09.xml")
	if err != nil {
		log.Fatal(err)
	}

	slug := r.URL.EscapedPath()
	if slug == "/favicon.ico" {
		// TODO
		return
	}

	if slug == "/" || slug == "" {
		fmt.Fprintf(w, "<head>")
		fmt.Fprint(w, "</head>")
		fmt.Fprint(w, "<body>")
		fmt.Fprintf(w, "<ul>")
		for _, post := range posts {
			fmt.Fprint(w, "<li>")
			fmt.Fprintf(w, "<a href=\"/%s\">", strconv.Itoa(post.id))
			fmt.Fprint(w, post.title)
			fmt.Fprint(w, "</a>")
			fmt.Fprint(w, "</li>")
		}
		fmt.Fprintf(w, "</ul>")
		fmt.Fprint(w, "</body>")
		return
	}

	id, err := strconv.Atoi(slug[1:])
	if err != nil {
		fmt.Print(err)
		fmt.Fprint(w, "unexpected url")
		return
	}

	idx := slices.IndexFunc(posts, func(a Page) bool {
		return a.id == id
	})

	if idx < 0 {
		fmt.Fprint(w, "unexpected url")
		return
	}

	post := posts[idx]
	fmt.Println("displaying ", post.title)

	fmt.Fprint(w, "<head>")
	fmt.Fprint(w, "<title>")
	fmt.Fprint(w, post.title)
	fmt.Fprint(w, "</title>")
	fmt.Fprint(w, "</head>")

	fmt.Fprint(w, "<body>")
	fmt.Fprint(w, "<h1>")
	fmt.Fprint(w, post.title)
	fmt.Fprint(w, "</h1>")
	fmt.Fprint(w, "<p>")
	fmt.Fprint(w, post.content)
	fmt.Fprint(w, "</p>")
	fmt.Fprint(w, "</body>")
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

func parsePosts(file_path string) ([]Page, error) {
	data, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	results, err := parseItems(string(data))
	if err != nil {
		return nil, err
	}

	contents := make([]Page, 0, len(results))
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

		page := Page{
			title:     title,
			content:   content,
			timestamp: date,
			id:        id,
		}

		contents = append(contents, page)
	}

	return contents, nil
}

type Page struct {
	title     string
	content   string
	timestamp time.Time
	id        int
}

func main() {

	// posts, err := parsePosts("wp_blog_2025-10-09.xml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, page := range posts {
	// 	fmt.Println(page.title)
	// 	fmt.Println(page.content)
	// }

	// fmt.Println(contents[1].title)
	// return
	// result := results[1]["content:encoded"].(string)
	// fmt.Println(result)
	// fmt.Println("end")
	// for _, item := range results {

	// 	fmt.Println(item[0])
	// 	break
	// }
	// decoder := xml2map.NewDecoder(strings.NewReader(string(data)))
	// result, err := decoder.Decode()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// result = result["rss"].(map[string]interface{})
	// result = result["channel"].(map[string]interface{})
	// results := result["item"].([]map[string]interface{})

	// for k, v := range results {
	// value := (map[string]string) v
	// fmt.Printf(value)
	// fmt.Fprint("%s", k)
	// fmt.Println(k)
	// fmt.Println(v)

	// if i == 3 {
	// 	break
	// }
	// }

	// fmt.Println(result)
	// channel := result["channel"]
	// fmt.Printf("%s\n", channel)

	// var feed string
	// set := make(map[string]string)

	// xml.Unmarshal(data, &set)
	// fmt.Print(feed)

	http.HandleFunc("/", handler)

	err := http.ListenAndServe("localhost:10000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
