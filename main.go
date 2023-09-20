package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/peterh/liner"
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
	s := PhotosService{client: client}
	/**
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

		m, ok := s.Search("IMG_1483203832100.JPG", MediaType_Photo, c)
		if ok {
			log.Println(m)
		}
	**/
	f := Finder{drive: d, photos: s, driveStack: new(Stack[*drive.File]), driveFilesKind: "folders"}
	f.driveStack.Push(&drive.File{Id: "root", Name: "/"})
	f.repl()
}

type Finder struct {
	drive            DriveService
	driveStack       *Stack[*drive.File]
	lastDriveFolders []*drive.File
	photos           PhotosService
	driveFilesKind   string
}

func (f *Finder) repl() {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	for {
		entry, err := line.Prompt(fmt.Sprintf("<%s::%s> ", f.driveFilesKind, Path(f.driveStack)))
		if err != nil {
			break
		}
		if entry == ":q" {
			break
		}
		if entry == ":f" {
			f.driveFilesKind = "folders"
			continue
		}
		if entry == ":p" {
			f.driveFilesKind = "photos"
			continue
		}
		if entry == "ls" {
			if f.driveFilesKind == "folders" {
				f.lastDriveFolders = f.drive.Folders(f.driveStack.Top().Id)
				for _, each := range f.lastDriveFolders {
					fmt.Println(each.Name)
				}
			}
			if f.driveFilesKind == "photos" {
				for _, each := range f.drive.Photos(f.driveStack.Top().Id) {
					fmt.Println(each.Name)
				}
			}
			continue
		}
		if strings.HasPrefix(entry, "cd") {
			parts := strings.Split(entry, " ")
			if len(parts) > 1 {
				dir := parts[1]
				if dir == ".." {
					// keep root
					if f.driveStack.Size() > 1 {
						f.driveStack.Pop()
					}
					continue
				}
				var found *drive.File
				for _, each := range f.lastDriveFolders {
					if each.Name == dir {
						found = each
						break
					}
				}
				if found == nil {
					fmt.Println(dir, " no such folder")
				} else {
					f.driveStack.Push(found)
				}
			}
			continue
		}
		line.AppendHistory(entry)
	}
}

func Path(s *Stack[*drive.File]) string {
	b := new(strings.Builder)
	for i, each := range s.items {
		if i > 1 {
			fmt.Fprintf(b, "/")
		}
		fmt.Fprintf(b, "%s", each.Name)
	}
	return b.String()
}
