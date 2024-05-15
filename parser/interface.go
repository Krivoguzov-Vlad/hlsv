package parser

type NestedFile interface {
	RelativeURL() string
	URL() string
}

type Playlist interface {
	NestedFile
	NestedFiles() []NestedFile
}
