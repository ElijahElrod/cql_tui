package tui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gocql/gocql"
	"log"
	"os"
	"strings"
)

const MARGIN = 2

type model struct {
	clusterSession   *gocql.Session
	cursor           int
	selected         map[int]struct{}
	keyspaceMetaData *gocql.KeyspaceMetadata
	prettyPrintJson  bool

	// models
	help      help.Model
	details   *Details
	searchBar *Search
	tpl       *TablePrintList
}

type scanMsg struct {
	keys []string
}

// Scan CQL Keyspace for tables, functions, aggregates, views, and user types.
func (m *model) Scan() tea.Cmd {

	keys := make([]string, 5)
	keysCursor := 0

	if cnt := len(m.keyspaceMetaData.Tables); cnt > 0 {
		keys[keysCursor] = "Tables"
		keysCursor++
	}

	if cnt := len(m.keyspaceMetaData.Functions); cnt > 0 {
		keys[keysCursor] = "Functions"
		keysCursor++
	}

	if cnt := len(m.keyspaceMetaData.Aggregates); cnt > 0 {
		keys[keysCursor] = "Aggregates"
		keysCursor++
	}

	if cnt := len(m.keyspaceMetaData.MaterializedViews); cnt > 0 {
		keys[keysCursor] = "Views"
		keysCursor++
	}

	if cnt := len(m.keyspaceMetaData.UserTypes); cnt > 0 {
		keys[keysCursor] = "UserTypes"
	}

	return func() tea.Msg {
		return scanMsg{keys: keys}
	}
}

func CreateCqlSession(conn, keyspace, username, password string) (*gocql.Session, error) {
	if conn == "" {
		conn = "localhost:7199"
	}

	cluster := gocql.NewCluster(strings.Split(conn, ",")...)
	if username != "" && password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: username,
			Password: password,
		}
	}
	if keyspace != "" {
		cluster.Keyspace = keyspace

	}

	session, err := cluster.CreateSession()

	if err != nil {
		return nil, err
	}
	return session, nil
}

func initialModel(cqlSession *gocql.Session, keyspace string, prettyPrintJson bool) (*model, error) {

	helpMenu := help.New()
	helpMenu.ShowAll = false

	keyspaceMetaData, err := cqlSession.KeyspaceMetadata(keyspace)
	if err != nil {
		log.Fatal(err)
	}

	return &model{
		clusterSession:   cqlSession,
		help:             helpMenu,
		searchBar:        NewSearch(),
		prettyPrintJson:  prettyPrintJson,
		keyspaceMetaData: keyspaceMetaData,
	}, nil
}
func (m *model) Init() tea.Cmd {
	return m.Scan()
}
func (m *model) UpdateSize(width int, height int) {
	m.tpl.width = width/2 - MARGIN
	m.details.width = width/2 - MARGIN

	m.tpl.height = height - 7
	m.details.height = height - 7
}

func setWindowSize(width int, height int) tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  width,
			Height: height,
		}
	}
}
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UpdateSize(msg.Width, msg.Height)

	case scanMsg: // new scan results
		search := strings.ReplaceAll(m.search, "*", "")
		for _, key := range msg.keys {
			split := strings.Split(key, ":")
			m.Node.AddChild(split, key, m.redis, search)
		}
		m.tpl.Update(updatePL{&m.Node})
		m.ScanCursor = msg.cursor

	case setTextMessage:
		m.reset(msg.text)
		cmds := tea.Sequence(m.Scan(), m.tpl.ResetCursor)
		return m, cmds

	case tea.KeyMsg:
		// if the search bar is active, pass all messages to search
		if m.search_bar.active {
			ms, cmd := m.search_bar.Update(msg)
			m.search_bar = ms.(*Search)

			return m, cmd
		}
		switch {
		case key.Matches(msg, default_keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, default_keys.Search):
			m.search_bar.ToggleActive(true)
			return m, nil

		case key.Matches(msg, default_keys.Enter):
			node := m.tpl.GetCurrent()
			if node != nil && len(node.Children) == 0 {
				cmd := m.GetDetails(node)
				return m, cmd
			} else if node != nil {
				node.Expanded = !node.Expanded
				m.tpl.Update(updatePL{&m.Node})
			}

		case key.Matches(msg, default_keys.Scan):
			return m, m.Scan()

		case key.Matches(msg, default_keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, default_keys.Search):
			m.search_bar.ToggleActive(true)
			return m, nil
		}
	}

	// update the other models on the screen
	res, cmd := m.tpl.Update(msg)
	if a, ok := res.(*TablePrintList); ok {
		m.tpl = a
	} else {
		return res, cmd
	}

	res, cmd = m.search_bar.Update(msg)
	if a, ok := res.(*Search); ok {
		m.search_bar = a
	} else {
		return res, cmd
	}

	res, cmd = m.details.Update(msg)
	if a, ok := res.(*Details); ok {
		m.details = a
	}
	return m, nil
}
func (m *model) View() string {
	// The header
	s := "Welcome to CQL_TUI!\n\n"

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func RunTUI(conn, keyspace, username, password string, prettyPrintJson bool) {
	cqlSession, err := CreateCqlSession(conn, keyspace, username, password)
	if err != nil {
		log.Fatal(err)
	}

	initialCqlModel, err := initialModel(cqlSession, keyspace, prettyPrintJson)
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(
		initialCqlModel,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
