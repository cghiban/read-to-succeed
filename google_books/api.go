package google_books

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Volume struct {
	ID         string     `json:"id"`
	VolumeInfo VolumeInfo `json:"volumeInfo"`
	SearchInfo struct {
		TextSnippet string `json:"textSnippet"`
	} `json:"searchInfo"`
}

type VolumeInfo struct {
	Title               string               `json:"title"`
	Subtitle            string               `json:"subtitle,omitempty"`
	Authors             []string             `json:"authors"`
	IndustryIdentifiers []IndustryIdentifier `json:"industryIdentifiers,omitempty"`
	Description         string               `json:"description"`
	PublishedDate       string               `json:"publishedDate"`
	PrintType           string               `json:"printType"`
	MainCategory        string               `json:"mainCategory"`
	Categories          []string             `json:"categories,omitempty"`
	ImageLinks          ImageLinks           `json:"imageLinks"`
	Language            string               `json:"language"`
}

type ImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
}

type VolumeSearchResult struct {
	TotalItems int      `json:"totalItems"`
	Items      []Volume `json:"items"`
	Kind       string   `json:"kind"`
	ID         string   `json:"id"`
}

type IndustryIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type VolumeSearchResultError struct {
	Status string       `json:"status"`
	Error  ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	endPoint string = "https://www.googleapis.com/books/v1/volumes"
)

func DoSearch(query, lang string) VolumeSearchResult {

	if lang == "" {
		lang = "en"
	}

	q := url.Values{}
	q.Set("projection", "lite") // Returns only certain fields
	q.Set("printType", "books") // Returns only results that are books
	q.Set("langRestrict", lang)
	q.Set("q", query)

	// q.Add("fields", "totalItems,items(id,volumeInfo/title,volumeInfo/authors,volumeInfo/subtitle,volumeInfo/description,volumeInfo/imageLinks,volumeInfo/language)")
	// fmt.Println("query:", q.Encode())

	fmt.Println("url:", q.Encode())
	apiKey := os.Getenv("GOOGLE_BOOKS_API_KEY")
	if apiKey != "" {
		q.Set("key", apiKey)
	}

	uri := endPoint + "?" + q.Encode()
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return VolumeSearchResult{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("resp.Status = %+v\n", resp.Status)
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ERROR reading data: %s\n", err)
		} else {
			fmt.Printf("ERROR: %s\n", respBody)
		}
		return VolumeSearchResult{}
	}

	var output VolumeSearchResult
	// respBody, _ := ioutil.ReadAll(resp.Body)
	// json.Unmarshal(respBody, &output)
	// fmt.Println(string(respBody))
	// return VolumeSearchResult{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&output)

	if err != nil {
		fmt.Println(err)
		return VolumeSearchResult{}
	}

	return output
}
