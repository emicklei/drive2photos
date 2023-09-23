package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
)

const (
	MediaType_Photo = "PHOTO"
	MediaType_Video = "VIDEO"
)

type PhotosService struct {
	client *http.Client
}

func (s *PhotosService) Upload(file *drive.File, content []byte) bool {
	fmt.Println("uploading", file.Name, "with", len(content), "bytes created on", file.CreatedTime, "mime", file.MimeType)

	// first upload bytes
	payloadReader := bytes.NewReader(content)
	req, err := http.NewRequest("POST", "https://photoslibrary.googleapis.com/v1/uploads", payloadReader)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-File-Name", file.Name)
	req.Header.Set("X-Goog-Upload-Protocol", "raw")
	req.Header.Set("X-Goog-Upload-Content-Type", file.MimeType)

	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("error:", resp.Status)
		return false
	}
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	defer resp.Body.Close()
	uploadTokenData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	uploadToken := string(uploadTokenData)
	if len(uploadToken) == 0 {
		fmt.Println("error:", "no upload token")
		return false
	}

	// now create media item
	req, err = http.NewRequest("POST", "https://photoslibrary.googleapis.com/v1/mediaItems:batchCreate", nil)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	// payload
	doc := map[string][]NewMediaItem{}
	doc["newMediaItems"] = []NewMediaItem{
		{
			Description: file.Description,
			SimpleMediaItem: SimpleMediaItem{
				Filename:    file.Name,
				UploadToken: uploadToken,
			},
		},
	}
	body, err := json.Marshal(doc)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	resp, err = s.client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("error:", resp.Status)
		return false
	}
	content, _ = io.ReadAll(resp.Body)
	fmt.Println("response:", string(content))

	return true
}

type NewMediaItem struct {
	Description     string          `json:"description,omitempty"`
	SimpleMediaItem SimpleMediaItem `json:"simpleMediaItem,omitempty"`
}

type SimpleMediaItem struct {
	Filename    string `json:"fileName,omitempty"`
	UploadToken string `json:"uploadToken,omitempty"`
}

func (s *PhotosService) Search(fileName, mediaType string, created time.Time) (MediaItem, bool) {
	year := created.Year()
	month := int(created.Month())
	day := created.Day()
	queryReader := strings.NewReader(fmt.Sprintf(`
	{"pageSize": 100
	,"filters": {		
		"mediaTypeFilter": {
			"mediaTypes": [
			  "%s"
			]
		},
		"dateFilter": {
			"ranges": [
				{
					"startDate": {
						"year": %d,
						"month": %d,
						"day": %d
					},
					"endDate": {
						"year": %d,
						"month": %d,
						"day": %d
					}
				}
			]
		}
	}
}
`, mediaType, year, month, day, year, month, day+1))
	resp, err := s.client.Post("https://photoslibrary.googleapis.com/v1/mediaItems:search",
		"application/json",
		queryReader)
	if err != nil {
		return MediaItem{}, false
	}
	defer resp.Body.Close()
	items := MediaItems{}
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		panic(err)
	}
	if len(items.MediaItems) == 0 {
		log.Println("no matching media items found")
	}
	for _, each := range items.MediaItems {
		if each.Filename == fileName {
			return each, true
		}
	}
	return MediaItem{}, false
}
