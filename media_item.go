package main

import "time"

type MediaItem struct {
	ID            string `json:"id"`
	ProductURL    string `json:"productUrl"`
	BaseURL       string `json:"baseUrl"`
	MimeType      string `json:"mimeType"`
	MediaMetadata struct {
		CreationTime time.Time `json:"creationTime"`
		Width        string    `json:"width"`
		Height       string    `json:"height"`
		Photo        struct {
			CameraMake      string  `json:"cameraMake"`
			CameraModel     string  `json:"cameraModel"`
			FocalLength     float64 `json:"focalLength"`
			ApertureFNumber float64 `json:"apertureFNumber"`
			IsoEquivalent   int     `json:"isoEquivalent"`
			ExposureTime    string  `json:"exposureTime"`
		} `json:"photo"`
	} `json:"mediaMetadata"`
	Filename string `json:"filename"`
}

type MediaItems struct {
	MediaItems    []MediaItem `json:"mediaItems"`
	NextPageToken string      `json:"nextPageToken"`
}
