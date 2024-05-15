package hls

import (
	"fmt"
	"strings"
)

// key or init file
type NestedFile struct {
	baseURL     string
	relativeURL string

	playlist *MediaPlaylist
}

func NewNestedFile(baseURL string, relativeURL string) *NestedFile {
	return &NestedFile{baseURL: baseURL, relativeURL: relativeURL}
}

func (s *NestedFile) WithMediaPlaylist(p *MediaPlaylist) *NestedFile {
	s.playlist = p
	return s
}

func (s *NestedFile) MediaPlaylist() *MediaPlaylist {
	return s.playlist
}

func (f NestedFile) RelativeURL() string {
	return f.relativeURL
}

func (f NestedFile) BaseURL() string {
	return f.baseURL
}

func (f NestedFile) URL() string {
	return fmt.Sprintf("%s/%s",
		strings.TrimRight(f.BaseURL(), "/"),
		strings.TrimLeft(f.RelativeURL(), "/"),
	)
}
