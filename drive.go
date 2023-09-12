package main

import (
	"fmt"
	"log"

	"google.golang.org/api/drive/v3"
)

// https://developers.google.com/drive/api/guides/ref-search-terms
type DriveService struct {
	service *drive.Service
	owner   string
}

func (s *DriveService) Folders(parent string) (list []*drive.File) {
	done := false
	pageToken := ""
	for !done {
		r, err := s.service.Files.List().
			Q(fmt.Sprintf(`
		'%s' in parents and
		mimeType = 'application/vnd.google-apps.folder' and 
		trashed=false and 
		'%s' in owners
		`, parent, s.owner)).
			PageToken(pageToken).
			PageSize(100).
			Fields("nextPageToken, files(id, name)").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
		list = append(list, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			done = true
		}
	}
	return
}

func (s *DriveService) Photos(parent string) (list []*drive.File) {
	done := false
	pageToken := ""
	for !done {
		r, err := s.service.Files.List().
			Q(fmt.Sprintf(`
		'%s' in parents and
		mimeType = 'image/png' and 
		trashed=false and 
		'%s' in owners
		`, parent, s.owner)).
			PageSize(100).
			PageToken(pageToken).
			Fields("nextPageToken, files(id, name,createdTime)").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
		list = append(list, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			done = true
		}
	}
	return
}
