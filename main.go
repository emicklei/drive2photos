package main

import (
	"context"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// https://developers.google.com/photos/library/guides/authorization
	config, err := google.ConfigFromJSON(b,
		drive.DrivePhotosReadonlyScope,
		drive.DriveFileScope,
		"https://www.googleapis.com/auth/drive.readonly.metadata",
		"https://www.googleapis.com/auth/photoslibrary.readonly",
		"https://www.googleapis.com/auth/photoslibrary.appendonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	d := DriveService{service: srv, owner: "ernest.micklei@gmail.com"}

	for _, each := range d.Folders("root") {
		log.Println(each.Name)
	}
	for _, each := range d.Photos("root") {
		when, err := time.Parse(time.RFC3339, each.CreatedTime)
		if err != nil {
			log.Println("cannot parse", each.CreatedTime)
		}
		log.Println(each.Name, when)
	}
	c, _ := time.Parse("2006-01-02", "2018-09-07")
	s := PhotosService{client: client}
	m, ok := s.Search("IMG_1483203832100.JPG", MediaType_Photo, c)
	if ok {
		log.Println(m)
	}
}
