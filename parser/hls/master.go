package hls

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/Krivoguzov-Vlad/hlsv/parser"

	"github.com/bldsoft/m3u8"
)

type MasterPlaylist struct {
	url string
	mp  *m3u8.MasterPlaylist
}

func NewPlaylist(url string, reader io.Reader, strict bool) (parser.Playlist, error) {
	playlist, playlistType, err := m3u8.DecodeFrom(reader, strict)
	if err != nil {
		return nil, err
	}

	if playlistType == m3u8.MASTER {
		return &MasterPlaylist{url, playlist.(*m3u8.MasterPlaylist)}, nil
	}
	base, relative := filepath.Split(url)
	return &MediaPlaylist{NestedFile: *NewNestedFile(base, relative), mp: playlist.(*m3u8.MediaPlaylist)}, nil
}

func NewMasterPlaylist(url string) *MasterPlaylist {
	return &MasterPlaylist{
		url: url,
	}
}

func (m *MasterPlaylist) DecodeFrom(r io.Reader, strict bool) error {
	p, _, err := m3u8.DecodeFrom(r, strict)
	if err != nil {
		return err
	}
	m.mp = p.(*m3u8.MasterPlaylist)
	return nil
}

func (m MasterPlaylist) URL() string {
	return m.url
}

func (m MasterPlaylist) RelativeURL() string {
	return m.URL()
}

func (m MasterPlaylist) BaseURL() string {
	withoutQuery, _, _ := strings.Cut(m.url, "?")
	baseURL, _ := filepath.Split(withoutQuery)
	return baseURL
}

func (m MasterPlaylist) NestedFiles() []parser.NestedFile {
	if len(m.mp.Variants) == 0 {
		return nil
	}
	baseURL := m.BaseURL()
	res := make([]parser.NestedFile, 0, len(m.mp.Variants)*(1+len(m.mp.Variants[0].Alternatives)))
	alternativesAdded := make(map[string]struct{})
	for _, variant := range m.mp.Variants {
		res = append(res, NewMediaPlaylist(baseURL, variant.URI))
		for _, alt := range variant.Alternatives {
			if _, ok := alternativesAdded[alt.URI]; !ok {
				res = append(res, NewMediaPlaylist(baseURL, alt.URI))
				alternativesAdded[alt.URI] = struct{}{}
			}
		}
	}
	return res
}

func (m *MasterPlaylist) ChangeNestedLinks(changeLink func(path string) string) {
	for _, variant := range m.mp.Variants {
		variant.URI = changeLink(variant.URI)
		for _, alt := range variant.Alternatives {
			alt.URI = changeLink(alt.URI)
		}
	}
}

func (m *MasterPlaylist) Encode() *bytes.Buffer {
	return m.mp.Encode()
}

func (m *MasterPlaylist) Bytes() []byte {
	return m.mp.Encode().Bytes()
}

func (m *MasterPlaylist) String() string {
	return m.mp.Encode().String()
}
