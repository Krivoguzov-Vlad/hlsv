package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"tmp/highlighter"
	hl "tmp/highlighter"
	"tmp/source"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"golang.org/x/net/webdav"
)

type PlaylistViewer struct {
	selectedHighlighter highlighter.Highlighter
	source              source.Source
	masterURL           string

	historyStack []string

	currentPlaylistData string
	selectedLineIdx     int
	selectMax           int
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
		hl.Pointer("->"),
	)

	pv.selectedLineIdx = 0
	pv.selectMax = 4
	return nil
}

func (pv *PlaylistViewer) GoTo(ctx context.Context, url string) (err error) {
	pv.currentPlaylistData, err = pv.source.GetFile(ctx, pv.masterURL)
	if err != nil {
		return err
	}
	pv.historyStack = append(pv.historyStack, url)
	return nil
}

func (pv *PlaylistViewer) GoPrev(ctx context.Context) (err error) {
	if len(pv.historyStack) <= 1 {
		return nil
	}
	pv.historyStack = pv.historyStack[:len(pv.historyStack)-1]
	current := pv.historyStack[len(pv.historyStack)-1]
	pv.currentPlaylistData, err = pv.source.GetFile(ctx, current)
	if err != nil {
		return err
	}
	return nil
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

		// These keys should exit the program.
		case "ctrl+c", "q":
			return pv, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if pv.selectedLineIdx > 0 {
				pv.selectedLineIdx--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if pv.selectedLineIdx < pv.selectMax {
				pv.selectedLineIdx++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "backspace", "ctrl+o":
			pv.GoPrev(ctx)
		case "enter", " ":
			pv.GoTo(ctx)
			// _, ok := m.selected[pv.selectedLineIdx]
			// if ok {
			// 	delete(m.selected, m.cursor)
			// } else {
			// 	m.selected[m.cursor] = struct{}{}
			// }
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return pv, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (pv *PlaylistViewer) View() string {

	strs := strings.Split(pv.currentPlaylistData, "\n")

	strs[pv.selectedLineIdx] = "->" + pv.selectedHighlighter.Highlight(strs[pv.selectedLineIdx])
	return strings.Join(strs, "\n")
	return pv.currentPlaylistData
}

func main() {
	webdav.Condition
	pv := NewPlaylistViewer("https://by-streampool-ext.spnode.net/10005/nodvr/hls/setplex_test/playlist.m3u8")
	if _, err := tea.NewProgram(pv, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
