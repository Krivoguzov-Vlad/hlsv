package main

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Krivoguzov-Vlad/hlsv/highlighter"
	hl "github.com/Krivoguzov-Vlad/hlsv/highlighter"
	parser "github.com/Krivoguzov-Vlad/hlsv/parser"
	"github.com/Krivoguzov-Vlad/hlsv/parser/hls"
	source "github.com/Krivoguzov-Vlad/hlsv/source"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

type ViewableFile struct {
	File parser.NestedFile
	Idx  int
}

func (f *ViewableFile) NestedLen() int {
	if p, ok := f.File.(parser.Playlist); ok {
		return len(p.NestedFiles())
	}
	return 0
}

func (f *ViewableFile) SelectNext() {
	f.Idx = min(f.NestedLen()-1, f.Idx+1)
}

func (f *ViewableFile) SelectPrev() {
	f.Idx = max(0, f.Idx-1)
}

func (f *ViewableFile) Selected() parser.NestedFile {
	if p, ok := f.File.(parser.Playlist); ok {
		return p.NestedFiles()[f.Idx]
	}
	return nil
}

func (f *ViewableFile) String() string {
	if p, ok := f.File.(fmt.Stringer); ok {
		return p.String()
	}
	return f.File.RelativeURL()
}

type PlaylistViewer struct {
	selectedHighlighter highlighter.Highlighter
	source              source.Source
	masterURL           string

	historyStack []*ViewableFile
}

func NewPlaylistViewer(masterURL string) *PlaylistViewer {
	return &PlaylistViewer{
		masterURL: masterURL,
	}
}

func (pv *PlaylistViewer) Init() tea.Cmd {
	pv.source = source.NewHTTPSource()
	ctx := context.Background()
	pv.GoTo(ctx, pv.masterURL)

	// Highlighter
	green := color.New(color.FgGreen, color.Bold)
	pv.selectedHighlighter = hl.Multi(
		hl.NewColor(green),
		// hl.Pointer("->"),
	)
	return nil
}

func (pv *PlaylistViewer) GoTo(ctx context.Context, url string) (err error) {
	data, err := pv.source.GetFile(ctx, url)
	if err != nil {
		return err
	}

	var nextFile parser.NestedFile
	nextFile, err = hls.NewPlaylist(url, strings.NewReader(data), false)
	if err != nil {
		nextFile = hls.NewNestedFile("", url)
	}
	pv.historyStack = append(pv.historyStack, &ViewableFile{
		File: nextFile,
		Idx:  0,
	})
	return nil
}

func (pv *PlaylistViewer) GoPrev(ctx context.Context) (err error) {
	if len(pv.historyStack) <= 1 {
		return nil
	}
	pv.historyStack = pv.historyStack[:len(pv.historyStack)-1]
	return nil
}

func (pv *PlaylistViewer) GoNext(ctx context.Context) (err error) {
	if selected := pv.Current().Selected(); selected != nil {
		return pv.GoTo(ctx, selected.URL())
	}
	return nil
}

func (pv *PlaylistViewer) UpdateCurrent(ctx context.Context) (err error) {
	if len(pv.historyStack) == 1 {
		pv.historyStack = nil
		return pv.GoTo(ctx, pv.masterURL)
	}
	return cmp.Or(pv.GoPrev(ctx), pv.GoNext(ctx))
}

func (pv *PlaylistViewer) Current() *ViewableFile {
	return pv.historyStack[len(pv.historyStack)-1]
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.

func (pv *PlaylistViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.TODO()
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {
		case "ctrl+c", "q":
			return pv, tea.Quit
		case "up", "k":
			if pv.Current().Idx > 0 {
				pv.Current().SelectPrev()
			} else {
				// TODO: fix
				// media, ok := pv.Current().File.(*hls.MediaPlaylist)
				// if ok {
				// 	u, _ := url.Parse(media.URL())
				// 	query := u.Query()
				// 	v := query.Get("timeshift")
				// 	timeshift, _ := strconv.Atoi(v)
				// 	timeshift += int(media.TargetDuration().Seconds())
				// 	query.Set("timeshift", strconv.Itoa(timeshift))
				// 	u.RawQuery = query.Encode()
				// 	pv.GoPrev(ctx)
				// 	pv.GoTo(ctx, u.String())
				// }
			}
		case "down", "j":
			pv.Current().SelectNext()
		case "backspace", "ctrl+o":
			pv.GoPrev(ctx)
		case "u":
			pv.UpdateCurrent(ctx)
		case "enter", " ":
			pv.GoNext(ctx)
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return pv, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (pv *PlaylistViewer) View() string {
	selected := pv.Current().Selected()
	currentContent := pv.Current().String()
	if selected == nil {
		return currentContent
	}

	strs := strings.Split(currentContent, "\n")

	idx := slices.IndexFunc(strs, func(str string) bool {
		return strings.Contains(str, selected.RelativeURL())
	})
	strs[idx] = pv.selectedHighlighter.Highlight(strs[idx])

	// strs = append([]string{pv.Current().File.URL()}, strs...)

	return strings.Join(strs, "\n")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("where is link?")
		return
	}
	pv := NewPlaylistViewer(os.Args[1])
	if _, err := tea.NewProgram(pv, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
