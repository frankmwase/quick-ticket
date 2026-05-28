package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// app state machine 
type state int

const (
	stateLogin     state = iota
	stateDashboard
)

// btea model 
type model struct {
	state     state
	baseURL   string
	tenantID  string
	apiKey    string
	apiClient *Client

	// Login
	loginFocus int // 0=url, 1=tenant, 2=apikey, 3=submit
	loginErr   string

	// 3D wireframe rotation
	angleX float64
	angleY float64
	angleZ float64

	// Dashboard
	menuIndex    int
	menuItems    []string
	resultText   string
	profileData  string

	// Sub-prompts (for verify token, batch ID, etc.)
	promptActive bool
	promptLabel  string
	promptValue  string
	promptAction func(m *model, value string) string

	// Terminal size
	width  int
	height int
}

func initialModel() model {
	return model{
		state:      stateLogin,
		baseURL:    "http://localhost:8080",
		tenantID:   "",
		apiKey:     "",
		loginFocus: 1,
		menuIndex:  0,
		menuItems: []string{
			"Generate Ticket",
			"Verify Ticket",
			"View Profile",
			"List Members",
			"Health Check",
		},
		width:  120,
		height: 40,
	}
}

// tick animation
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		key := msg.String()

		// Global quit
		if key == "ctrl+c" {
			return m, tea.Quit
		}

		// sub prompts
		if m.promptActive {
			switch key {
			case "enter":
				if m.promptAction != nil {
					m.resultText = m.promptAction(&m, m.promptValue)
				}
				m.promptActive = false
				m.promptValue = ""
			case "esc":
				m.promptActive = false
				m.promptValue = ""
			case "backspace":
				if len(m.promptValue) > 0 {
					m.promptValue = m.promptValue[:len(m.promptValue)-1]
				}
			default:
				if len(key) == 1 {
					m.promptValue += key
				}
			}
			return m, nil
		}

		// login screen
		if m.state == stateLogin {
			switch key {
			case "tab", "down":
				m.loginFocus = (m.loginFocus + 1) % 4
			case "shift+tab", "up":
				m.loginFocus--
				if m.loginFocus < 0 {
					m.loginFocus = 3
				}
			case "enter":
				if m.loginFocus == 3 {
					m.apiClient = NewClient(m.baseURL, m.tenantID, m.apiKey)

					// health check
					_, err := m.apiClient.Health()
					if err != nil {
						m.loginErr = "Connection failed: " + err.Error()
					} else {
						// trying profile, failures are ok
						prof, err := m.apiClient.GetProfile()
						if err != nil {
							m.profileData = "(auth error: " + err.Error() + ")"
						} else {
							m.profileData = prettyJSON(prof)
						}
						m.state = stateDashboard
						m.loginErr = ""
					}
				}
			case "backspace":
				switch m.loginFocus {
				case 0:
					if len(m.baseURL) > 0 {
						m.baseURL = m.baseURL[:len(m.baseURL)-1]
					}
				case 1:
					if len(m.tenantID) > 0 {
						m.tenantID = m.tenantID[:len(m.tenantID)-1]
					}
				case 2:
					if len(m.apiKey) > 0 {
						m.apiKey = m.apiKey[:len(m.apiKey)-1]
					}
				}
			default:
				if len(key) == 1 {
					switch m.loginFocus {
					case 0:
						m.baseURL += key
					case 1:
						m.tenantID += key
					case 2:
						m.apiKey += key
					}
				}
			}
		}

		// dashboard screen
		if m.state == stateDashboard {
			switch key {
			case "q":
				return m, tea.Quit
			case "up", "k":
				if m.menuIndex > 0 {
					m.menuIndex--
				}
			case "down", "j":
				if m.menuIndex < len(m.menuItems)-1 {
					m.menuIndex++
				}
			case "backspace":
				// Back to login
				m.state = stateLogin
				m.apiClient = nil
				m.resultText = ""
				m.profileData = ""
			case "enter":
				m.resultText = m.executeMenuAction()
			}
		}

	case tickMsg:
		m.angleX += 0.02
		m.angleY += 0.03
		m.angleZ += 0.01
		return m, tickCmd()
	}

	return m, nil
}

func (m *model) executeMenuAction() string {
	if m.apiClient == nil {
		return "X No connection"
	}

	switch m.menuIndex {
	case 0: // Generate Ticket
		t, err := m.apiClient.GenerateTicket(1, "tui-user")
		if err != nil {
			return "X " + err.Error()
		}
		return "Generated:\n" + prettyJSON(t)

	case 1: // Verify Ticket
		m.promptActive = true
		m.promptLabel = "ENTER TOKEN TO VERIFY"
		m.promptAction = func(mod *model, value string) string {
			res, err := mod.apiClient.VerifyTicket(value)
			if err != nil {
				return "X " + err.Error()
			}
			return "Verification:\n" + prettyJSON(res)
		}
		return "Awaiting token input..."

	case 2: // View Profile
		prof, err := m.apiClient.GetProfile()
		if err != nil {
			return "X " + err.Error()
		}
		return "Profile:\n" + prettyJSON(prof)

	case 3: // List Members
		mems, err := m.apiClient.GetMembers()
		if err != nil {
			return "X " + err.Error()
		}
		return "Members:\n" + prettyJSON(mems)

	case 4: // Health Check
		h, err := m.apiClient.Health()
		if err != nil {
			return "X " + err.Error()
		}
		return "" + prettyJSON(h)
	}

	return ""
}

func (m model) View() string {
	if m.state == stateLogin {
		return m.loginView()
	}
	return m.dashboardView()
}

// ui styles
var (
	colorGreen  = lipgloss.Color("#00ff41")
	colorAmber  = lipgloss.Color("#ffb000")
	colorRed    = lipgloss.Color("#ff003c")
	colorDim    = lipgloss.Color("#333333")
	colorDimGrn = lipgloss.Color("#006611")

	styleTitle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleLabel = lipgloss.NewStyle().
			Foreground(colorAmber).
			Bold(true)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDimGrn)

	styleErr = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	styleFocused = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDimGrn).
			Padding(1, 2)

	styleSelectedMenu = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true).
				Background(lipgloss.Color("#003300"))

	styleNormalMenu = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// login view
func (m model) loginView() string {
	focus := func(idx int) lipgloss.Style {
		if m.loginFocus == idx {
			return styleFocused
		}
		return styleNormal
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(styleDim.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")
	sb.WriteString(styleTitle.Render("        QUICK-TICKET ∕∕ TERMINAL ACCESS") + "\n")
	sb.WriteString(styleDim.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")
	sb.WriteString("\n")

	sb.WriteString(focus(0).Render(fmt.Sprintf("  ENDPOINT    │ %s", m.baseURL)) + "\n")
	sb.WriteString(focus(1).Render(fmt.Sprintf("  TENANT ID   │ %s", m.tenantID)) + "\n")
	sb.WriteString(focus(2).Render(fmt.Sprintf("  API KEY     │ %s", maskKey(m.apiKey))) + "\n")
	sb.WriteString("\n")
	sb.WriteString(focus(3).Render("  [ ESTABLISH CONNECTION ]") + "\n")
	sb.WriteString("\n")

	if m.loginErr != "" {
		sb.WriteString(styleErr.Render("  ⚠ "+m.loginErr) + "\n")
	}

	sb.WriteString(styleDim.Render("  tab/↑↓ navigate · enter confirm · ctrl+c quit") + "\n")

	return sb.String()
}

func maskKey(key string) string {
	if len(key) == 0 {
		return ""
	}
	if len(key) <= 4 {
		return strings.Repeat("•", len(key))
	}
	return strings.Repeat("•", len(key)-4) + key[len(key)-4:]
}

// dashboard view
func (m model) dashboardView() string {
	// Left: 3D wireframe
	wireW := 42
	wireH := 18
	if m.width > 0 && m.width < 100 {
		wireW = 30
		wireH = 14
	}
	wireframe := RenderWireframe(wireW, wireH, m.angleX, m.angleY, m.angleZ)
	leftPane := stylePanel.Copy().
		Width(wireW + 6).
		Render(wireframe)

	// Right: Menu + Results
	var rightContent strings.Builder

	// Header
	rightContent.WriteString(styleTitle.Render("QT∕∕DASHBOARD") + "\n")
	rightContent.WriteString(styleDim.Render("tenant: "+m.tenantID) + "\n")
	rightContent.WriteString(styleDim.Render(strings.Repeat("─", 44)) + "\n\n")

	// Menu
	rightContent.WriteString(styleLabel.Render("OPERATIONS") + "\n")
	for i, item := range m.menuItems {
		if i == m.menuIndex {
			rightContent.WriteString(styleSelectedMenu.Render(" "+item+" ") + "\n")
		} else {
			rightContent.WriteString(styleNormalMenu.Render("  "+item) + "\n")
		}
	}

	// Prompt overlay
	if m.promptActive {
		rightContent.WriteString("\n")
		rightContent.WriteString(styleLabel.Render("  "+m.promptLabel) + "\n")
		rightContent.WriteString(styleFocused.Render("  > "+m.promptValue+"█") + "\n")
		rightContent.WriteString(styleDim.Render("  enter submit · esc cancel") + "\n")
	}

	// Results
	rightContent.WriteString("\n" + styleDim.Render(strings.Repeat("─", 44)) + "\n")
	rightContent.WriteString(styleLabel.Render("OUTPUT") + "\n")
	result := m.resultText
	if result == "" {
		result = styleDim.Render("(select an action and press enter)")
	}
	// Truncate long results
	lines := strings.Split(result, "\n")
	maxLines := 12
	if m.height > 0 {
		maxLines = m.height/2 - 8
		if maxLines < 6 {
			maxLines = 6
		}
	}
	if len(lines) > maxLines {
		lines = append(lines[:maxLines], styleDim.Render("... (truncated)"))
	}
	rightContent.WriteString(strings.Join(lines, "\n") + "\n")

	rightPane := stylePanel.Copy().
		Width(50).
		Height(wireH + 2).
		Render(rightContent.String())

	// Compose
	main := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	// Footer
	footer := styleDim.Render("  ↑↓ navigate · enter execute · backspace logout · q quit")

	return main + "\n" + footer + "\n"
}

// Main

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting TUI: %v\n", err)
		os.Exit(1)
	}
}
