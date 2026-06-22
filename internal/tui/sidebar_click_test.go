package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rtalexk/demux/internal/config"
)

func newClickSidebar(sessions ...string) SidebarModel {
	initStyles(Theme{IconTmuxSession: "S"}, config.ProcessesConfig{}, nil)
	var s SidebarModel
	for _, name := range sessions {
		s.nodes = append(s.nodes, SidebarNode{Session: name})
	}
	return s
}

func TestNodeAtRow_ListViewDirectMapping(t *testing.T) {
	s := newClickSidebar("a", "b", "c", "d")
	// List view, no stats: each node is one row, no scroll hints. row r -> node r.
	for r := 0; r < 4; r++ {
		if got := s.NodeAtRow(30, 20, r); got != r {
			t.Errorf("NodeAtRow(row=%d) = %d, want %d", r, got, r)
		}
	}
	if got := s.NodeAtRow(30, 20, 4); got != -1 {
		t.Errorf("NodeAtRow(row=4) = %d, want -1 (past last node)", got)
	}
	if got := s.NodeAtRow(30, 20, -1); got != -1 {
		t.Errorf("NodeAtRow(row=-1) = %d, want -1", got)
	}
}

func TestNodeAtRow_AccountsForFocusedRowExpansion(t *testing.T) {
	s := newClickSidebar("a", "b", "c", "d")
	s.cfg.Sidebar.ShowSessionStats = true
	s.cursor = 1
	s.SetSessionStats(map[string]SessionStat{
		"b": {CPUNow: 5, MemNow: 1024 * 1024, CPUPeak: 9, MemPeak: 2 * 1024 * 1024},
	})
	// Rows: 0->a, 1->b(name), 2->b(cpu), 3->b(mem), 4->c, 5->d
	cases := map[int]int{0: 0, 1: 1, 2: 1, 3: 1, 4: 2, 5: 3}
	for row, want := range cases {
		if got := s.NodeAtRow(30, 20, row); got != want {
			t.Errorf("NodeAtRow(row=%d) = %d, want %d", row, got, want)
		}
	}
}

func TestSetCursorIndex(t *testing.T) {
	s := newClickSidebar("a", "b", "c")
	s.SetCursorIndex(2, 10)
	if s.cursor != 2 {
		t.Errorf("cursor = %d, want 2", s.cursor)
	}
	s.SetCursorIndex(99, 10) // out of range: no-op
	if s.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (out-of-range ignored)", s.cursor)
	}
	s.SetCursorIndex(-1, 10) // negative: no-op
	if s.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (negative ignored)", s.cursor)
	}
}

func TestHandleMouseMsg_LeftClickSelectsRow(t *testing.T) {
	initStyles(Theme{IconTmuxSession: "S"}, config.ProcessesConfig{}, nil)
	var m Model
	m.width = 100
	m.height = 40
	m.cfg.Sidebar.Width = 30
	m.cfg.StatusBar.Show = true
	m.sidebar.nodes = []SidebarNode{{Session: "a"}, {Session: "b"}, {Session: "c"}, {Session: "d"}}

	// Click node index 2: Y = searchBoxH(3) + top-border(1) + contentRow(2) = 6.
	msg := tea.MouseMsg{X: 5, Y: 6, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	m2, _ := m.handleMouseMsg(msg)

	if m2.sidebar.cursor != 2 {
		t.Errorf("cursor = %d, want 2 after click on row for node 2", m2.sidebar.cursor)
	}
	if m2.focus != panelSidebar {
		t.Errorf("focus = %v, want panelSidebar after sidebar click", m2.focus)
	}
}

func TestHandleMouseMsg_ClickOutsideSidebarIgnored(t *testing.T) {
	initStyles(Theme{IconTmuxSession: "S"}, config.ProcessesConfig{}, nil)
	var m Model
	m.width = 100
	m.height = 40
	m.cfg.Sidebar.Width = 30
	m.sidebar.nodes = []SidebarNode{{Session: "a"}, {Session: "b"}}
	m.sidebar.cursor = 0

	// X beyond the sidebar (in the proclist panel): ignored.
	msg := tea.MouseMsg{X: 60, Y: 6, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	m2, _ := m.handleMouseMsg(msg)
	if m2.sidebar.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (click outside sidebar ignored)", m2.sidebar.cursor)
	}
}
