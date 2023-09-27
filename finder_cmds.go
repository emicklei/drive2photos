package main

import (
	"fmt"
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
	createdTime, err := time.Parse(time.RFC3339, found.CreatedTime)
	if err != nil {
		fmt.Println("cannot parse created time", found.CreatedTime)
		return
	}
	mediaItem, ok := f.photos.Search(fileName, MediaType_Photo, createdTime)
	if ok {
		fmt.Println("found copy on Google Photos: ", mediaItem.ProductURL)
	} else {
		fmt.Println("not found on Google Photos")
	}
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
	createdTime, err := time.Parse(time.RFC3339, found.CreatedTime)
	if err != nil {
		fmt.Println("cannot parse created time", found.CreatedTime)
		return false
	}
	mediaItem, ok := f.photos.Search(fileName, MediaType_Photo, createdTime)
	if ok {
		fmt.Println("found copy on Google Photos, no copy needed: ", mediaItem.ProductURL)
		return false
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
