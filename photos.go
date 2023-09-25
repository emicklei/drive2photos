package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
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

// https://developers.google.com/photos/library/guides/upload-media#creating-media-bp
func (s *PhotosService) Upload(file *drive.File, content []byte) bool {
	mimeType := file.MimeType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(file.Name))
	}
	fmt.Println("uploading", file.Name, "with", len(content), "bytes created on", file.CreatedTime, "mime", mimeType)

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
	req.Header.Set("X-Goog-Upload-Content-Type", mimeType)

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
	// fmt.Println(string(body))
	req, err = http.NewRequest("POST", "https://photoslibrary.googleapis.com/v1/mediaItems:batchCreate", bytes.NewReader(body))
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

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
	content, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	resultDoc := NewMediaItemResultsDoc{}
	err = json.Unmarshal(content, &resultDoc)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}

	// fmt.Println(string(content))

	createdTime, err := time.Parse(time.RFC3339, file.CreatedTime)
	if err != nil {
		fmt.Println("cannot parse created time", file.CreatedTime)
		return false
	}
	fmt.Println("photo stored on timeline at", createdTime)

	return true
}

type NewMediaItemResultsDoc struct {
	NewMediaItemResults []struct {
		UploadToken string `json:"uploadToken"`
		Status      struct {
			Message string `json:"message"`
		} `json:"status"`
		MediaItem struct {
			ID            string `json:"id"`
			ProductURL    string `json:"productUrl"`
			MimeType      string `json:"mimeType"`
			MediaMetadata struct {
				CreationTime time.Time `json:"creationTime"`
				Width        string    `json:"width"`
				Height       string    `json:"height"`
			} `json:"mediaMetadata"`
			Filename string `json:"filename"`
		} `json:"mediaItem"`
	} `json:"newMediaItemResults"`
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
