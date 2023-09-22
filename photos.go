package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	MediaType_Photo = "PHOTO"
	MediaType_Video = "VIDEO"
)

type PhotosService struct {
	client *http.Client
}

func (s *PhotosService) Upload(payloadReader io.Reader, mimeType string) bool {
	resp, err := s.client.Post("https://photoslibrary.googleapis.com/v1/uploads",
		mimeType,
		payloadReader)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("error:", resp.Status)
		return false
	}
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	return true
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
