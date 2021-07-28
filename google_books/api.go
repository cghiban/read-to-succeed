package google_books

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	Authors             []string             `json:"authors"`
	IndustryIdentifiers []IndustryIdentifier `json:"industryIdentifiers"`
	Description         string               `json:"description"`
	PublishedDate       string               `json:"publishedDate"`
	PrintType           string               `json:"printType"`
	MainCategory        string               `json:"mainCategory"`
	Categories          []string             `json:"categories"`
	ImageLinks          ImageLinks           `json:"imageLinks"`
	Language            string               `json:"language"`
}

type ImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"small"`
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

//curl -sk "https://www.googleapis.com/books/v1/volumes?q=Mihai%20Eminescu&fields=totalItems,kind,items(id,volumeInfo/title,volumeInfo/authors,volumeInfo/subtitle,volumeInfo/description,volumeInfo/imageLinks,volumeInfo/language)
func DoSearch(query string) VolumeSearchResult {

	req, _ := http.NewRequest("GET", endPoint, nil)
	req.Header.Add("Accept", "application/json")

	//uri := endPoint + "?fields=totalItems,items(id,volumeInfo/title,volumeInfo/authors,volumeInfo/subtitle,volumeInfo/description,volumeInfo/imageLinks,volumeInfo/language)"
	uri := endPoint + "?projection=lite"
	uri += "&printType=books"
	uri += "&q=" + url.QueryEscape(query)

	req.URL.RawQuery = uri
	fmt.Println("url:", uri)

	/*
		q := req.URL.Query()
		//q.Add("key", "key_from_environment_or_flag")
		q.Add("q", query)
		q.Add("fields", "totalItems,items(id,volumeInfo/title,volumeInfo/authors,volumeInfo/subtitle,volumeInfo/description,volumeInfo/imageLinks,volumeInfo/language)")
		fmt.Println("query:", q.Encode())
	*/

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return VolumeSearchResult{}
	}

	defer resp.Body.Close()

	fmt.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
	fmt.Printf("resp.Status = %+v\n", resp.Status)
	if resp.StatusCode != 200 {
		// return err
	}

	var output VolumeSearchResult
	//respBody, _ := ioutil.ReadAll(resp.Body)
	//json.Unmarshal(respBody, &output)
	//fmt.Println(string(respBody))
	//return VolumeSearchResult{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&output)

	if err != nil {
		fmt.Println(err)
		return VolumeSearchResult{}
	}

	return output
}
