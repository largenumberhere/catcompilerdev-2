package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sbabiv/xml2map"
)

// type Feed struct {
// 	XMLName xml.Name "xml:"
// }

// type Channel struct {
// 	Item string
// }

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "" || r.Method == "GET" {
		fmt.Fprint(w, "hello")
		return
	}

	fmt.Fprintf(w, "meow!")

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

}

type Page struct {
	title   string
	content string
}

func main() {
	data, err := os.ReadFile("wp_blog_2025-10-09.xml")
	if err != nil {
		log.Fatal(err)
	}

	results, err := parseItems(string(data))
	if err != nil {
		log.Fatal(err)
	}

	contents := make([]Page, 0, len(results))
	for _, item := range results {
		post_type := item["1.2:post_type"]
		if post_type != "post" {
			continue
		}

		content := item["content:encoded"].(string)
		title := item["title"].(string)
		page := Page{
			title:   title,
			content: content,
		}

		contents = append(contents, page)

	}

	for _, page := range contents {
		fmt.Println(page.title)
		fmt.Println(page.content)
	}

	// fmt.Println(contents[1].title)
	return
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

	err = http.ListenAndServe("localhost:10000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
