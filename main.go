package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/peterh/liner"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var owner = flag.String("email", "", "Google email address")

var cmds = ":q :p :f cd ls cp rm mv ff"

func main() {
	flag.Parse()
	fmt.Println("drive2photos --- [" + cmds + "]")

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
		if strings.HasPrefix(entry, "ff") {
			obj := parameterFromEntry(entry)
			if obj != "" {
				f.search(obj)
			}
			continue
		}
		if strings.HasPrefix(entry, "cp") {
			obj := parameterFromEntry(entry)
			if obj != "" {
				f.cp(obj)
			}
			continue
		}
		if strings.HasPrefix(entry, "rm") {
			obj := parameterFromEntry(entry)
			if obj != "" {
				f.rm(obj)
			}
			f.ls()
			continue
		}
		if strings.HasPrefix(entry, "mv") {
			obj := parameterFromEntry(entry)
			if obj != "" {
				if f.cp(obj) {
					f.rm(obj)
				}
			}
			f.ls()
			continue
		}
		if strings.HasPrefix(entry, "cd") {
			dir := parameterFromEntry(entry)
			if dir != "" {
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
					found := f.drive.FolderByName(dir)
					if found == nil {
						fmt.Println(dir, " no such folder (did you run ls?)")
					} else {
						f.driveStack.Push(found)
					}
				} else {
					f.driveStack.Push(found)
				}
			}
			f.ls()
			continue
		} else {
			// fallback
			fmt.Println("unknown command, try " + cmds)
		}
		line.AppendHistory(entry)
	}
}

func parameterFromEntry(entry string) string {
	space := strings.Index(entry, " ")
	if space == -1 {
		return entry
	}
	return strings.TrimSpace(entry[space+1:])
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
