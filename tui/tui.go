package tui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
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
}

type scanMsg struct {
	keys []string
}

// Scan CQL Keyspace for tables, functions, aggregates, views, and user types.
func (m model) Scan() tea.Cmd {

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
func (m model) Init() tea.Cmd {
	return m.Scan()
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}
func (m model) View() string {
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
