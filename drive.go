package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// https://developers.google.com/drive/api/guides/ref-search-terms
type DriveService struct {
	service *drive.Service
	owner   string
	client  *http.Client
}

func (s *DriveService) FolderByName(dir string) *drive.File {
	call := s.service.Files.List().Q(fmt.Sprintf("name = '%s'", dir))
	f, err := call.Do()
	if err != nil {
		fmt.Printf("Unable to retrieve files: %v\n", err)
		return nil
	}
	if len(f.Files) == 0 {
		return nil
	}
	return f.Files[0]
}

func (s *DriveService) Delete(f *drive.File) bool {
	fmt.Println("deleting", f.Name)

	err := s.service.Files.Delete(f.Id).Do()
	if err != nil {
		fmt.Printf("unable to delete file: %v/n", err)
		return false
	}
	return true
}

func (s *DriveService) Download(f *drive.File) ([]byte, bool) {
	fmt.Println("downloading", f.Name)

	resp, err := s.service.Files.Get(f.Id).Download()
	if err != nil {
		fmt.Printf("unable to download file: %v/n", err)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("unable to download file: %v/n", resp.Status)
		return nil, false
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("unable to download file: %v/n", err)
		return nil, false
	}
	return data, true
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
			if uerr, ok := err.(*url.Error); ok {
				if oerr, ok := uerr.Err.(*oauth2.RetrieveError); ok {
					if oerr.ErrorCode == "invalid_grant" {
						fmt.Println("Your saved access token (token.json) is no longer valid ; retry after deleting it")
						return list
					}
				}
			}
			fmt.Printf("Unable to retrieve files: %v (%T)\n", err, err)
		}
		list = append(list, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			done = true
		}
	}
	return
}

// https://developers.google.com/drive/api/reference/rest/v3/files
func (s *DriveService) Photos(parent string) (list []*drive.File) {
	done := false
	pageToken := ""
	for !done {
		r, err := s.service.Files.List().
			Q(fmt.Sprintf(`
		'%s' in parents and
		(mimeType = 'image/png' or mimeType = 'image/jpeg' or name contains '.JPG') and 
		trashed=false and 
		'%s' in owners
		`, parent, s.owner)).
			PageSize(100).
			PageToken(pageToken).
			Fields("nextPageToken, files(id, name,createdTime,modifiedTime,modifiedByMeTime,originalFilename)").Do()
		if err != nil {
			if uerr, ok := err.(*url.Error); ok {
				if oerr, ok := uerr.Err.(*oauth2.RetrieveError); ok {
					if oerr.ErrorCode == "invalid_grant" {
						fmt.Println("Your saved access token (token.json) is no longer valid ; retry after deleting it")
						return list
					}
				}
			}
			fmt.Printf("Unable to retrieve files: %v\n", err)
		}
		list = append(list, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			done = true
		}
	}
	return
}
