package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
)

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

func (f *Finder) search(fileName string) {
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
	// fmt.Println("mod me", found.ModifiedByMeTime)
	// fmt.Println("crea", found.CreatedTime)
	// fmt.Println("shar", found.SharedWithMeTime)
	// fmt.Println("mod", found.ModifiedTime)
	var searchTime time.Time
	if found.ModifiedByMe {
		searchTime, _ = time.Parse(time.RFC3339, found.ModifiedByMeTime)
	} else {
		searchTime, _ = time.Parse(time.RFC3339, found.ModifiedTime)
	}
	mediaItem, ok := f.photos.Search(fileName, MediaType_Photo, searchTime)
	if ok {
		fmt.Println("found copy on Google Photos: ", mediaItem.ProductURL)
	} else {
		fmt.Println("not found on Google Photos")
	}
}

func (f *Finder) matches(entry string) (list []string) {
	if f.driveFilesKind == "folders" {
		fmt.Println("cannot match on folders")
		return list
	}
	if strings.Contains(entry, "*") {
		entry = strings.Replace(entry, "*", ".*", -1)
		for _, each := range f.lastListing {
			if m, _ := regexp.MatchString(entry, each.Name); m {
				list = append(list, each.Name)
			}
		}
	} else {
		list = append(list, entry)
	}
	return list
}
func (f *Finder) rm(fileName string) bool {
	if f.driveFilesKind == "folders" {
		fmt.Println("cannot copy folders")
		return false
	}
	if fileName == "*" {
		for _, each := range f.lastListing {
			if !f.rm(each.Name) {
				return false
			}
		}
		return true
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
		return false
	}
	if f.drive.Delete(found) {
		fmt.Println("... done")
	}
	return true
}

func (f *Finder) cp(fileName string) bool {
	if f.driveFilesKind == "folders" {
		fmt.Println("cannot copy folders")
		return false
	}
	if fileName == "*" {
		for _, each := range f.lastListing {
			if !f.cp(each.Name) {
				return false
			}
		}
		return true
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
		return false
	}
	searchTime, err := time.Parse(time.RFC3339, found.ModifiedTime)
	if err != nil {
		fmt.Println("cannot parse created time", found.ModifiedTime)
		return false
	}
	mediaItem, ok := f.photos.Search(fileName, MediaType_Photo, searchTime)
	if ok {
		fmt.Println("found copy on Google Photos, no copy needed: ", mediaItem.ProductURL)
		return true
	}
	data, ok := f.drive.Download(found)
	if !ok {
		return false
	}
	fmt.Println("... done")
	if !f.photos.Upload(found, data) {
		return false
	}
	fmt.Println("... done")
	return true
}
