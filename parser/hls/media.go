package hls

import (
	"bytes"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/Krivoguzov-Vlad/hlsv/parser"

	"github.com/bldsoft/m3u8"
)

type MediaType = m3u8.MediaType

const (
	MediaTypeLive  MediaType = 0
	MediaTypeEvent MediaType = m3u8.EVENT
	MediaTypeVOD   MediaType = m3u8.VOD
)

type MediaPlaylist struct {
	NestedFile
	mp *m3u8.MediaPlaylist
}

func NewMediaPlaylist(baseURL, relativeURL string) *MediaPlaylist {
	return &MediaPlaylist{
		NestedFile: *NewNestedFile(baseURL, relativeURL),
	}
}

func (m *MediaPlaylist) WinSize() uint {
	return m.mp.WinSize()
}

func (m *MediaPlaylist) SetWinSize(winsize uint) error {
	return m.mp.SetWinSize(winsize)
}

func (m *MediaPlaylist) Slide(uri string, duration float64, title string) {
	m.mp.Slide(uri, duration, title)
	m.mp.ResetCache()
}

func (m *MediaPlaylist) DecodeFrom(r io.Reader, strict bool) error {
	p, _, err := m3u8.DecodeFrom(r, strict)
	if err != nil {
		return err
	}
	m.mp = p.(*m3u8.MediaPlaylist)
	return nil
}

func (m MediaPlaylist) SetType(t MediaType) {
	m.mp.MediaType = t
	m.mp.Closed = t == MediaTypeVOD
	switch {
	case t == MediaTypeLive && m.WinSize() == 0:
		_ = m.SetWinSize(5)
	case t == MediaTypeVOD || t == MediaTypeEvent:
		_ = m.SetWinSize(0)
	}
}

func (m MediaPlaylist) IsLive() bool {
	return m.mp.MediaType == MediaTypeLive
}

func (m MediaPlaylist) SetMediaSequence(seqNo uint64) {
	m.mp.SeqNo = seqNo
}

func (m MediaPlaylist) MediaSequence() uint64 {
	return m.mp.SeqNo
}

func (m MediaPlaylist) SegmentBaseURL() string {
	withoutQuery, _, _ := strings.Cut(m.URL(), "?")
	segmentBaseURL, _ := filepath.Split(withoutQuery)
	return segmentBaseURL
}

func (m *MediaPlaylist) NestedFiles() []parser.NestedFile {
	segments := m.Segments()
	res := make([]parser.NestedFile, 0, len(segments)+m.mp.Keys.Len()+1)
	if m.mp.Keys != nil {
		for _, k := range *m.mp.Keys {
			if m.isRelativeURL(k.URI) {
				res = append(res, NewNestedFile(m.SegmentBaseURL(), k.URI).WithMediaPlaylist(m))
			}
		}
	}
	if m.mp.Map != nil {
		res = append(res, NewNestedFile(m.SegmentBaseURL(), m.mp.Map.URI).WithMediaPlaylist(m))
	}
	for _, s := range segments {
		res = append(res, s)
	}
	return res
}

func (m *MediaPlaylist) isRelativeURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err != nil || !u.IsAbs()
}

func (m MediaPlaylist) TargetDuration() time.Duration {
	return time.Duration(m.mp.TargetDuration) * time.Second
}

func (m MediaPlaylist) UpdateInterval() time.Duration {
	if m.mp.Closed {
		return 0
	}
	return m.TargetDuration()
}

func (m *MediaPlaylist) Append(uri string, duration time.Duration, title string) error {
	return m.mp.Append(uri, duration.Seconds(), title)
}

func (m *MediaPlaylist) AppendSegment(segment *Segment) error {
	return m.mp.AppendSegment(segment.segment)
}

func (m *MediaPlaylist) segment(s *m3u8.MediaSegment) *Segment {
	return NewSegment(m.SegmentBaseURL(), s).WithMediaPlaylist(m)
}

func (m *MediaPlaylist) FirstSegment() *Segment {
	if segments := m.mp.GetAllSegments(); len(segments) > 0 {
		return m.segment(segments[0])
	}
	return nil
}

func (m *MediaPlaylist) LastSegment() *Segment {
	if segments := m.mp.GetAllSegments(); len(segments) > 0 {
		return m.segment(segments[len(segments)-1])
	}
	return nil
}

func (m *MediaPlaylist) PopSegment() error {
	return m.mp.Remove()
}

func (m *MediaPlaylist) ResetSegments(resetMediaSeq bool) error {
	seqNo := m.mp.SeqNo
	if resetMediaSeq {
		seqNo = 0
	}
	l := len(m.Segments())
	for i := 0; i < l; i++ {
		if err := m.PopSegment(); err != nil {
			return err
		}
	}

	m.mp.SeqNo = seqNo
	return nil
}

func (m *MediaPlaylist) Segments() []*Segment {
	segments := m.mp.GetAllSegments()
	res := make([]*Segment, 0, len(segments))
	for _, seg := range segments {
		res = append(res, m.segment(seg))
	}
	return res
}

func (m *MediaPlaylist) Encode() *bytes.Buffer {
	return m.mp.Encode()
}

func (m *MediaPlaylist) Bytes() []byte {
	return m.mp.Encode().Bytes()
}

func (m *MediaPlaylist) String() string {
	return m.mp.Encode().String()
}
