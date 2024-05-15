package hls

import (
	"fmt"
	"strings"
	"time"

	"github.com/bldsoft/m3u8"
)

type Segment struct {
	baseURL string
	segment *m3u8.MediaSegment

	playlist *MediaPlaylist
}

func NewSegment(baseURL string, segment *m3u8.MediaSegment) *Segment {
	return &Segment{baseURL: baseURL, segment: segment}
}

func (s *Segment) WithMediaPlaylist(p *MediaPlaylist) *Segment {
	s.playlist = p
	return s
}

func (s *Segment) MediaPlaylist() *MediaPlaylist {
	return s.playlist
}

func (s *Segment) BaseURL() string {
	return s.baseURL
}

func (s *Segment) RelativeURL() string {
	return s.segment.URI
}

func (s *Segment) URL() string {
	return fmt.Sprintf("%s/%s",
		strings.TrimRight(s.BaseURL(), "/"),
		strings.TrimLeft(s.RelativeURL(), "/"),
	)
}

func (s *Segment) Title() string {
	return s.segment.Title
}

func (s *Segment) SeqID() uint64 {
	return s.segment.SeqId
}

func (s *Segment) Time() time.Time {
	return s.segment.ProgramDateTime
}

func (s *Segment) SetSeqID(seqID uint64) {
	s.segment.SeqId = seqID
}

func (s *Segment) Duration() time.Duration {
	return time.Duration(s.segment.Duration) * time.Second
}
