// Package main provides a CLI setup wizard for the gopl-server.
// It guides the user through database configuration, validates connections,
// and generates the necessary .config.yaml files for local development.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopl-dev/server/app"
	"github.com/jackc/pgx/v5"
	"gopkg.in/yaml.v3"
)

var (
	errYAMLNodeUnsupportedValueType = errors.New("unsupported value type")
	errEmptyYAML                    = errors.New("empty yaml document")
	errSectionNotFound              = errors.New("section not found")
	errSubsectionNotFound           = errors.New("subsection not found")
	errKeyNotFound                  = errors.New("key not found")
	errDatabaseNotEmpty             = errors.New("database already has tables")
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	swTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	swSectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	swFocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	swBlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	swHelpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	swSuccessStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	swErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	swDimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
)

// ── Async messages ────────────────────────────────────────────────────────────

type msgConnectResult struct{ err error }
type msgDBCheckResult struct {
	created bool
	err     error
}

// ── Step state ────────────────────────────────────────────────────────────────

// stepState holds the runtime state for a single step.
// extra is used for step-specific non-input state (e.g. selects, flags).
type stepState struct {
	inputs  []textinput.Model
	focused int
	extra   any
	err     string
}

// val returns the input value at index i, falling back to the placeholder if empty.
func (s stepState) val(i int) string {
	v := s.inputs[i].Value()
	if v == "" {
		return s.inputs[i].Placeholder
	}
	return v
}

// focusInput blurs the currently focused input and focuses the one at index to.
func (s stepState) focusInput(to int) stepState {
	if s.focused < len(s.inputs) {
		s.inputs[s.focused].Blur()
		s.inputs[s.focused].PromptStyle = swBlurredStyle
		s.inputs[s.focused].TextStyle = swBlurredStyle
	}
	s.inputs[to].Focus()
	s.inputs[to].PromptStyle = swFocusedStyle
	s.inputs[to].TextStyle = swFocusedStyle
	s.focused = to
	return s
}

// makeInput creates a styled textinput with optional password echo mode.
func makeInput(secret bool) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.PromptStyle = swBlurredStyle
	ti.TextStyle = swBlurredStyle
	if secret {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}
	return ti
}

// renderInput renders a labeled textinput row, highlighted when focused.
func renderInput(s stepState, idx int, label string) string {
	var b strings.Builder
	if idx == s.focused {
		b.WriteString(swFocusedStyle.Render("▶ "+label) + "\n")
	} else {
		b.WriteString(swBlurredStyle.Render("  "+label) + "\n")
	}
	b.WriteString(s.inputs[idx].View() + "\n\n")
	return b.String()
}

// renderError renders a red error line prefixed with ✗. Returns empty string if err is empty.
func renderError(err string) string {
	if err == "" {
		return ""
	}
	return "\n" + swErrorStyle.Render("✗ "+err) + "\n"
}

// ── Step definition ───────────────────────────────────────────────────────────

// Step describes a single wizard step.
type Step struct {
	// Init returns the initial state for this step.
	Init func() stepState

	// View renders the step UI given its current state.
	View func(s stepState) string

	// HandleKey processes a key event and returns the updated state.
	// Returns (state, cmd, advance) where advance=true means move to next step.
	HandleKey func(s stepState, msg tea.KeyMsg) (stepState, tea.Cmd, bool)

	// HandleMsg processes an async tea.Msg (e.g. DB connect result).
	// Returns (state, cmd, advance, ok) where ok=false means stay on this step.
	HandleMsg func(s stepState, msg tea.Msg) (stepState, tea.Cmd, bool, bool)

	// Values returns section→key→value triples used when writing the config.
	Values func(s stepState) []yamlValue
}

// ── DB credentials step ───────────────────────────────────────────────────────

const (
	dbIdxHost = 0
	dbIdxUser = 1
	dbIdxPass = 2
)

var stepDBCredentials = Step{
	Init: func() stepState {
		inputs := []textinput.Model{
			makeInput(false),
			makeInput(false),
			makeInput(true),
		}
		s := stepState{inputs: inputs}
		return s.focusInput(dbIdxHost)
	},

	View: func(s stepState) string {
		var b strings.Builder
		b.WriteString(swSectionStyle.Render("── Database connection ──") + "\n\n")
		b.WriteString(renderInput(s, dbIdxHost, "Host:Port"))
		b.WriteString(renderInput(s, dbIdxUser, "User"))
		b.WriteString(renderInput(s, dbIdxPass, "Password"))
		b.WriteString(renderError(s.err))
		if s.focused == dbIdxPass {
			b.WriteString("\n" + swDimStyle.Render("Press enter to test connection") + "\n")
		}
		return b.String()
	},

	HandleKey: func(s stepState, msg tea.KeyMsg) (stepState, tea.Cmd, bool) {
		switch msg.Type {
		case tea.KeyEnter, tea.KeyTab, tea.KeyDown:
			if s.focused < dbIdxPass {
				return s.focusInput(s.focused + 1), textinput.Blink, false
			}
			host := s.val(dbIdxHost)
			user := s.val(dbIdxUser)
			pass := s.val(dbIdxPass)
			return s, func() tea.Msg {
				conn, err := pgx.Connect(context.Background(), buildConnStr(host, user, pass, "postgres"))
				if err != nil {
					return msgConnectResult{err: err}
				}
				_ = conn.Close(context.Background())
				return msgConnectResult{}
			}, false
		case tea.KeyShiftTab, tea.KeyUp:
			if s.focused > dbIdxHost {
				return s.focusInput(s.focused - 1), textinput.Blink, false
			}
		}
		return s, nil, false
	},

	HandleMsg: func(s stepState, msg tea.Msg) (stepState, tea.Cmd, bool, bool) {
		r, ok := msg.(msgConnectResult)
		if !ok {
			return s, nil, false, false
		}
		if r.err != nil {
			s.err = r.err.Error()
			return s, nil, false, true
		}
		s.err = ""
		return s, nil, true, true
	},

	Values: func(s stepState) []yamlValue {
		host, port, _ := strings.Cut(s.val(dbIdxHost), ":")
		if port == "" {
			port = "5432"
		}
		return []yamlValue{
			yv("db", "host", host),
			yv("db", "port", port),
			yv("db", "user", s.val(dbIdxUser)),
			yv("db", "password", s.val(dbIdxPass)),
		}
	},
}

// ── DB name step ──────────────────────────────────────────────────────────────

const dbNameIdx = 0

var stepDBName = Step{
	Init: func() stepState {
		inputs := []textinput.Model{makeInput(false)}
		s := stepState{inputs: inputs}
		return s.focusInput(dbNameIdx)
	},

	View: func(s stepState) string {
		var b strings.Builder
		b.WriteString(swSuccessStyle.Render("✓ Connected successfully") + "\n\n")
		b.WriteString(swSectionStyle.Render("── Database name ──") + "\n\n")
		b.WriteString(renderInput(s, dbNameIdx, "DB Name"))
		b.WriteString(swHelpStyle.Render("  If the database does not exist, we'll try to create it") + "\n")
		b.WriteString(swHelpStyle.Render("  Tip: use a \"_local_dev\" suffix (e.g. gopl_local_dev) —") + "\n")
		b.WriteString(swHelpStyle.Render("  the reset tool uses this convention to prevent accidents") + "\n")
		b.WriteString(renderError(s.err))
		return b.String()
	},

	HandleKey: func(s stepState, msg tea.KeyMsg) (stepState, tea.Cmd, bool) {
		switch msg.Type {
		case tea.KeyEnter, tea.KeyTab, tea.KeyDown:
			dbName := s.val(dbNameIdx)

			creds, ok := s.extra.(dbCreds)
			if !ok {
				s.err = "internal error: database credentials not found"
				return s, nil, false
			}

			if dbName != "" && dbName == creds.originalName {
				return s, func() tea.Msg {
					return msgDBCheckResult{created: false, err: nil}
				}, false
			}

			return s, func() tea.Msg {
				return checkDB(creds.host, creds.user, creds.pass, dbName)
			}, false
		}
		return s, nil, false
	},

	HandleMsg: func(s stepState, msg tea.Msg) (stepState, tea.Cmd, bool, bool) {
		r, ok := msg.(msgDBCheckResult)
		if !ok {
			return s, nil, false, false
		}
		if r.err != nil {
			s.err = r.err.Error()
			return s, nil, false, true
		}
		s.err = ""
		return s, nil, true, true
	},

	Values: func(_ stepState) []yamlValue { return nil },
}

// dbCreds is passed as extra to stepDBName so it can fire the async check.
type dbCreds struct {
	host, user, pass string
	originalName     string // Добавляем это поле
}

// checkDB connects to dbName, creating it if it doesn't exist, and verifies it has no tables.
func checkDB(hostPort, user, pass, dbName string) tea.Msg {
	ctx := context.Background()
	connStr := buildConnStr(hostPort, user, pass, dbName)
	conn, err := pgx.Connect(ctx, connStr)
	created := false

	if err != nil {
		pgConn, pgErr := pgx.Connect(ctx, buildConnStr(hostPort, user, pass, "postgres"))
		if pgErr != nil {
			return msgDBCheckResult{err: pgErr}
		}
		_, pgErr = pgConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %q", dbName))
		_ = pgConn.Close(ctx)
		if pgErr != nil {
			return msgDBCheckResult{err: fmt.Errorf("create database: %w", pgErr)}
		}
		created = true

		conn, err = pgx.Connect(ctx, connStr)
		if err != nil {
			return msgDBCheckResult{err: err}
		}
	}
	defer func() {
		_ = conn.Close(ctx)
	}()

	var count int
	err = conn.QueryRow(ctx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'",
	).Scan(&count)
	if err != nil {
		return msgDBCheckResult{err: fmt.Errorf("query tables: %w", err)}
	}
	if count > 0 {
		return msgDBCheckResult{err: fmt.Errorf("%w: %s (%d tables)", errDatabaseNotEmpty, dbName, count)}
	}
	return msgDBCheckResult{created: created}
}

// ── Step registry ─────────────────────────────────────────────────────────────

// allSteps defines the wizard flow in order.
// To add a new step: append a Step{} value here.
// To remove a step: delete its entry.
var allSteps = []Step{
	stepDBCredentials,
	stepDBName,
}

const (
	stepIdxDBCreds = 0
	stepIdxDBName  = 1
)

// ── Model ─────────────────────────────────────────────────────────────────────

type swModel struct {
	stepIdx  int
	states   []stepState
	done     bool
	dbInfo   dbCreds // passed to stepDBName as extra
	dbName   string  // resolved after stepDBName
	created  bool    // whether the DB was just created
	writeErr string  // error from writing .config.yaml
}

// newSWModel initializes the wizard model. If ex is non-nil, inputs are pre-populated
// from the existing config so the user can edit rather than retype unchanged values.
// sample is used for placeholder defaults (loaded from config.sample.yaml).
func newSWModel(ex *app.ConfigT, sample *app.ConfigT) swModel {
	states := make([]stepState, len(allSteps))
	for i, step := range allSteps {
		states[i] = step.Init()
	}

	// apply sample defaults as placeholders
	if sample != nil {
		sampleHostPort := sample.DB.Host
		if sample.DB.Port != "" && !strings.Contains(sampleHostPort, ":") {
			sampleHostPort = sampleHostPort + ":" + sample.DB.Port
		}
		states[stepIdxDBCreds].inputs[dbIdxHost].Placeholder = sampleHostPort
		states[stepIdxDBCreds].inputs[dbIdxUser].Placeholder = sample.DB.User
		states[stepIdxDBName].inputs[dbNameIdx].Placeholder = sample.DB.Name
	}

	m := swModel{states: states}

	if ex == nil {
		return m
	}

	// pre-populate inputs from existing .config.yaml
	hostPort := ex.DB.Host
	if ex.DB.Port != "" && !strings.Contains(hostPort, ":") {
		hostPort = hostPort + ":" + ex.DB.Port
	}

	setInputDefault(states[stepIdxDBCreds].inputs, dbIdxHost, hostPort)
	setInputDefault(states[stepIdxDBCreds].inputs, dbIdxUser, ex.DB.User)
	setInputDefault(states[stepIdxDBCreds].inputs, dbIdxPass, ex.DB.Password)
	setInputDefault(states[stepIdxDBName].inputs, dbNameIdx, ex.DB.Name)

	return m
}

// setInputDefault pre-fills an input with value so it appears as editable text.
// The slice element must be passed directly (not via pointer) because textinput.Model is a struct.
func setInputDefault(inputs []textinput.Model, idx int, value string) {
	if value == "" {
		return
	}
	inputs[idx].SetValue(value)
}

// ── Init / Update ─────────────────────────────────────────────────────────────

func (m swModel) Init() tea.Cmd { return textinput.Blink }

func (m swModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// handle summary screen
	if m.done {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyEnter:
				err := swWriteConfig(m)
				if err != nil {
					m.writeErr = err.Error()
					return m, nil
				}
				return m, tea.Quit
			}
		}
		return m, nil
	}

	step := m.currentStep()
	state := m.currentState()

	// handle async messages first
	if step.HandleMsg != nil {
		if newState, cmd, advance, handled := step.HandleMsg(state, msg); handled {
			m.states[m.stepIdx] = newState
			if advance {
				if m.stepIdx == stepIdxDBCreds {
					m.dbInfo = dbCreds{
						host:         newState.val(dbIdxHost),
						user:         newState.val(dbIdxUser),
						pass:         newState.val(dbIdxPass),
						originalName: "",
					}
				}
				if m.stepIdx == stepIdxDBName {
					if r, ok := msg.(msgDBCheckResult); ok {
						m.dbName = newState.val(dbNameIdx)
						m.created = r.created
					}
				}
				m = m.advance()
			}
			return m, cmd
		}
	}

	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		if state.focused >= 0 && state.focused < len(state.inputs) {
			var cmd tea.Cmd
			state.inputs[state.focused], cmd = state.inputs[state.focused].Update(msg)
			m.states[m.stepIdx] = state
			return m, cmd
		}
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	}

	newState, cmd, advance := step.HandleKey(state, keyMsg)
	m.states[m.stepIdx] = newState
	if advance {
		m = m.advance()
	}

	if cmd == nil && !advance && newState.focused >= 0 && newState.focused < len(newState.inputs) {
		var inputCmd tea.Cmd
		m.states[m.stepIdx].inputs[newState.focused], inputCmd = newState.inputs[newState.focused].Update(keyMsg)
		return m, inputCmd
	}

	return m, cmd
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m swModel) View() string {
	if m.done {
		return m.summaryView()
	}

	var b strings.Builder
	b.WriteString(swTitleStyle.Render("gopl-server · Setup Wizard") + "\n")
	b.WriteString(swHelpStyle.Render("tab/enter: next · shift+tab: prev · esc: cancel") + "\n\n")
	b.WriteString(m.currentStep().View(m.currentState()))
	return b.String()
}

// currentStep returns the Step definition for the active step index.
func (m swModel) currentStep() Step { return allSteps[m.stepIdx] }

// currentState returns the runtime state for the active step.
func (m swModel) currentState() stepState { return m.states[m.stepIdx] }

// advance moves to the next step, injecting dbCreds into stepDBName when needed.
// Sets done=true when there are no more steps.
func (m swModel) advance() swModel {
	next := m.stepIdx + 1
	if next >= len(allSteps) {
		m.done = true
		return m
	}
	m.stepIdx = next
	// inject dbCreds into stepDBName so it can fire the async check
	if next == stepIdxDBName {
		info := m.dbInfo
		info.originalName = m.states[next].val(dbNameIdx)
		m.states[next].extra = info
	}
	return m
}

// summaryView renders the final confirmation screen showing collected values.
func (m swModel) summaryView() string {
	var b strings.Builder
	b.WriteString(swTitleStyle.Render("gopl-server · Setup Wizard") + "\n\n")
	b.WriteString(swSuccessStyle.Render("✓ Database is ready") + "\n")
	if m.created {
		b.WriteString(swDimStyle.Render("  (database was created)") + "\n")
	}
	b.WriteString("\n" + swSectionStyle.Render("── Summary ──") + "\n\n")

	dbState := m.states[stepIdxDBCreds]
	b.WriteString(swDimStyle.Render("  host:port   ") + dbState.val(dbIdxHost) + "\n")
	b.WriteString(swDimStyle.Render("  db user     ") + dbState.val(dbIdxUser) + "\n")
	b.WriteString(swDimStyle.Render("  db password ") + strings.Repeat("•", len(dbState.val(dbIdxPass))) + "\n")
	b.WriteString(swDimStyle.Render("  db name     ") + m.dbName + "\n")

	b.WriteString("\n" + swSuccessStyle.Render("  Press enter to write .config.yaml") + "\n")
	if m.writeErr != "" {
		b.WriteString("\n" + swErrorStyle.Render("✗ "+m.writeErr) + "\n")
	}
	return b.String()
}

// ── Config writer ─────────────────────────────────────────────────────────────

// buildConnStr builds a PostgreSQL DSN from hostPort (host:port), user, pass, and dbName.
func buildConnStr(hostPort, user, pass, dbName string) string {
	host, port, _ := strings.Cut(hostPort, ":")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbName)
}

// ── YAML writer ───────────────────────────────────────────────────────────────

// yamlValue represents a value to set at a path in a YAML document.
// section is the top-level key, subsection is an optional nested key, key is the field.
type yamlValue struct {
	section    string
	subsection string
	key        string
	value      any
}

// yv is a shorthand constructor for yamlValue without a subsection.
func yv(section, key string, value any) yamlValue {
	return yamlValue{section: section, key: key, value: value}
}

// yvs is a shorthand constructor for yamlValue with a subsection.
func yvs(section, subsection, key, value string) yamlValue {
	return yamlValue{section: section, subsection: subsection, key: key, value: value}
}

// setYAMLValue finds section[.subsection].key in a yaml.Node document and sets its value.
// Preserves comments and structure of the original document.
func setYAMLValue(doc *yaml.Node, v yamlValue) error {
	root := doc
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return errEmptyYAML
		}
		root = root.Content[0]
	}

	sectionNode := mappingValue(root, v.section)
	if sectionNode == nil {
		return fmt.Errorf("%w: %s", errSectionNotFound, v.section)
	}

	target := sectionNode
	if v.subsection != "" {
		target = mappingValue(sectionNode, v.subsection)
		if target == nil {
			return fmt.Errorf("%w: %s in %s", errSubsectionNotFound, v.subsection, v.section)
		}
	}

	valueNode := mappingValue(target, v.key)
	if valueNode == nil {
		return fmt.Errorf("%w: %s in %s", errKeyNotFound, v.key, v.section)
	}

	switch val := v.value.(type) {
	case bool:
		valueNode.Kind = yaml.ScalarNode
		valueNode.Value = strconv.FormatBool(val)
		valueNode.Tag = "!!bool"

	case int:
		valueNode.Kind = yaml.ScalarNode
		valueNode.Value = strconv.Itoa(val)
		valueNode.Tag = "!!int"

	case string:
		valueNode.Kind = yaml.ScalarNode
		valueNode.Value = val
		valueNode.Tag = "!!str"

	default:
		return fmt.Errorf("%w: %T", errYAMLNodeUnsupportedValueType, v.value)
	}

	return nil
}

// mappingValue returns the value node for key in a YAML mapping node, or nil if not found.
func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// applyValues applies a list of yamlValue changes to a parsed YAML document.
func applyValues(doc *yaml.Node, vals []yamlValue) error {
	for _, v := range vals {
		err := setYAMLValue(doc, v)
		if err != nil {
			return fmt.Errorf("set %s.%s: %w", v.section, v.key, err)
		}
	}
	return nil
}

// encodeYAML encodes a yaml.Node back to string, preserving structure and comments.
func encodeYAML(doc *yaml.Node) (string, error) {
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2) //nolint:mnd
	err := enc.Encode(doc)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// swWriteConfig writes .config.yaml from config.sample.yaml with wizard values applied,
// then calls writeTestConfigs to write the test config files.
func swWriteConfig(m swModel) error {
	src, err := os.ReadFile("config.sample.yaml")
	if err != nil {
		return fmt.Errorf("read config.sample.yaml: %w", err)
	}

	var doc yaml.Node
	err = yaml.Unmarshal(src, &doc)
	if err != nil {
		return fmt.Errorf("parse config.sample.yaml: %w", err)
	}

	var vals []yamlValue
	for i, step := range allSteps {
		if step.Values == nil {
			continue
		}
		vals = append(vals, step.Values(m.states[i])...)
	}
	vals = append(vals, yv("db", "name", m.dbName))

	err = applyValues(&doc, vals)
	if err != nil {
		return err
	}

	content, err := encodeYAML(&doc)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	err = os.WriteFile(".config.yaml", []byte(content), 0o600) //nolint:mnd
	if err != nil {
		return fmt.Errorf("write .config.yaml: %w", err)
	}

	return writeTestConfigs(m, src)
}

var testConfigPaths = []string{
	"test/api_test/.config.yaml",
	"test/service_test/.config.yaml",
	"test/worker_test/.config.yaml",
}

// writeTestConfigs writes test configs based on config.sample.yaml with
// test-specific overrides and a "_test" suffix on the DB name.
func writeTestConfigs(m swModel, src []byte) error {
	testDBName := m.dbName + "_test"

	creds := m.states[stepIdxDBCreds]
	err := ensureTestDB(creds.val(dbIdxHost), creds.val(dbIdxUser), creds.val(dbIdxPass), testDBName)
	if err != nil {
		return fmt.Errorf("test database '%s': %w", testDBName, err)
	}

	var doc yaml.Node
	err = yaml.Unmarshal(src, &doc)
	if err != nil {
		return fmt.Errorf("parse config.sample.yaml: %w", err)
	}

	dbVals := stepDBCredentials.Values(m.states[stepIdxDBCreds])
	vals := make([]yamlValue, 0, len(dbVals)+4) //nolint:mnd
	vals = append(vals, dbVals...)
	vals = append(vals,
		yv("db", "name", testDBName),
		yv("email", "driver", "test"),
		yv("tracing", "enabled", "false"), // Обратите внимание: в вашем коде была строка "false"
		yv("files", "storage_driver", "in-memory-fs"),
	)

	err = applyValues(&doc, vals)
	if err != nil {
		return err
	}

	content, err := encodeYAML(&doc)
	if err != nil {
		return fmt.Errorf("encode test config: %w", err)
	}

	for _, path := range testConfigPaths {
		_, statErr := os.Stat(path)
		if statErr == nil {
			continue // already exists — skip
		}
		err = os.WriteFile(path, []byte(content), 0o600) //nolint:mnd
		if err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}

// ensureTestDB connects to dbName and returns nil if it exists.
// If the connection fails, it attempts to create the database.
// Unlike checkDB, it does not require the database to be empty.
func ensureTestDB(hostPort, user, pass, dbName string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, buildConnStr(hostPort, user, pass, dbName))
	if err == nil {
		_ = conn.Close(ctx)
		return nil
	}

	// DB doesn't exist — create it
	pgConn, err := pgx.Connect(ctx, buildConnStr(hostPort, user, pass, "postgres"))
	if err != nil {
		return err
	}
	_, err = pgConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %q", dbName))
	_ = pgConn.Close(ctx)
	return err
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	fmt.Println(swErrorStyle.Render("⚠  WARNING"))
	fmt.Println(swErrorStyle.Render("   This wizard is intended for local development setup only."))
	fmt.Println(swErrorStyle.Render("   It will create or overwrite .config.yaml and may create databases."))
	fmt.Println(swErrorStyle.Render("   Do NOT run this in staging or production environments."))
	fmt.Println()
	fmt.Print("   Proceed? [Y/n]: ")
	var ans string
	_, err := fmt.Scanln(&ans)
	if err != nil && err.Error() != "unexpected newline" && err.Error() != "EOF" {
		fmt.Fprintln(os.Stderr, swErrorStyle.Render("\n✗ input error: "+err.Error()))
		os.Exit(1)
	}
	ans = strings.ToLower(strings.TrimSpace(ans))
	if ans != "" && ans != "y" {
		fmt.Println("Aborted.")
		return
	}
	fmt.Println()

	sample, err := app.ConfigFromFile("config.sample.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, swErrorStyle.Render("⚠  failed to load config.sample.yaml: "+err.Error()))
		os.Exit(1)
	}

	ex, err := app.ConfigFromFile(".config.yaml")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, swErrorStyle.Render("⚠  failed to load .config.yaml: "+err.Error()))
		fmt.Fprintln(os.Stderr, swErrorStyle.Render("   fix the file or delete it to start fresh"))
		os.Exit(1)
	}
	if errors.Is(err, os.ErrNotExist) {
		ex = nil
	}

	p := tea.NewProgram(newSWModel(ex, sample), tea.WithAltScreen())
	raw, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}

	result, ok := raw.(swModel)
	if !ok {
		fmt.Fprintln(os.Stderr, "error: terminal model is not of type swModel")
		os.Exit(1)
	}
	if result.writeErr != "" {
		fmt.Fprintln(os.Stderr, swErrorStyle.Render("✗ "+result.writeErr))
		os.Exit(1)
	}

	_, statErr := os.Stat(".config.yaml")
	if statErr == nil {
		fmt.Println(swSuccessStyle.Render("✓ .config.yaml updated successfully"))
	} else {
		fmt.Println(swSuccessStyle.Render("✓ .config.yaml created successfully"))
	}
	for _, p := range testConfigPaths {
		_, statErr := os.Stat(p)
		if statErr == nil {
			fmt.Println(swDimStyle.Render("  " + p + " already exists, skipped"))
		} else {
			fmt.Println(swSuccessStyle.Render("✓ " + p + " created successfully"))
		}
	}

	fmt.Println("")
	fmt.Println(swDimStyle.Render("   Your server is ready to run."))
	fmt.Println(swDimStyle.Render("   You can manually edit .config.yaml to check and configure remaining options."))
}
