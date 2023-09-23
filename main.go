package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/peterh/liner"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var owner = flag.String("email", "", "Google email address")

func main() {
	flag.Parse()
	fmt.Println("drive2photos --- [:q :p :f cd ls cp rm mv]")

	if *owner == "" {
		fmt.Println("email flag is required")
		return
	}

	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// https://developers.google.com/photos/library/guides/authorization
	config, err := google.ConfigFromJSON(b,
		drive.DrivePhotosReadonlyScope,
		drive.DriveScope,
		"https://www.googleapis.com/auth/drive.readonly.metadata",
		"https://www.googleapis.com/auth/photoslibrary.readonly",
		"https://www.googleapis.com/auth/photoslibrary.appendonly",
		"https://www.googleapis.com/auth/photoslibrary.edit.appcreateddata")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	d := DriveService{service: srv, owner: *owner, client: new(http.Client)}
	s := PhotosService{client: client}
	f := Finder{drive: d, photos: s, driveStack: new(Stack[*drive.File]), driveFilesKind: "folders"}
	f.driveStack.Push(&drive.File{Id: "root", Name: "/"})
	f.ls()
	f.repl()
}

type Finder struct {
	drive          DriveService
	driveStack     *Stack[*drive.File]
	lastListing    []*drive.File
	photos         PhotosService
	driveFilesKind string
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
			f.ls()
			continue
		}
		if entry == ":p" {
			f.driveFilesKind = "photos"
			f.ls()
			continue
		}
		if entry == "ls" {
			f.ls()
			continue
		}
		if strings.HasPrefix(entry, "cp") {
			parts := strings.Split(entry, " ")
			if len(parts) > 1 {
				obj := parts[1]
				f.cp(obj)
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
					f.ls()
					continue
				}
				var found *drive.File
				for _, each := range f.lastListing {
					if each.Name == dir {
						found = each
						break
					}
				}
				if found == nil {
					fmt.Println(dir, " no such folder (did you run ls?)")
				} else {
					f.driveStack.Push(found)
				}
			}
			f.ls()
			continue
		} else {
			// fallback
			fmt.Println("unknown command, try :q :p :f cd ls cp rm mv")
		}
		line.AppendHistory(entry)
	}
}

func (f *Finder) ls() {
	found := false
	if f.driveFilesKind == "folders" {
		f.lastListing = f.drive.Folders(f.driveStack.Top().Id)
		for _, each := range f.lastListing {
			found = true
			fmt.Println(each.Name)
		}
		if !found {
			fmt.Println("no folders found")
		}
		return
	}
	if f.driveFilesKind == "photos" {
		f.lastListing = f.drive.Photos(f.driveStack.Top().Id)
		for _, each := range f.lastListing {
			found = true
			fmt.Println(each.Name)
		}
		if !found {
			fmt.Println("no photos found")
		}
		return
	}
}

func (f *Finder) cp(fileName string) {
	if f.driveFilesKind == "folders" {
		fmt.Println("cannot copy folders")
		return
	}
	var found *drive.File
	for _, each := range f.lastListing {
		if each.OriginalFilename == fileName || each.Name == fileName {
			found = each
			break
		}
	}
	if found == nil {
		fmt.Println(fileName, " no such file (did you run ls?)")
		return
	}
	createdTime, err := time.Parse(time.RFC3339, found.CreatedTime)
	if err != nil {
		fmt.Println("cannot parse created time", found.CreatedTime)
		return
	}
	mediaItem, ok := f.photos.Search(fileName, MediaType_Photo, createdTime)
	if ok {
		fmt.Println("found copy on Google Photos, no copy needed: ", mediaItem.ProductURL)
		return
	}
	data, ok := f.drive.Download(found)
	if !ok {
		return
	}
	fmt.Println("... done")
	if !f.photos.Upload(found, data) {
		return
	}
	fmt.Println("... done")
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
