package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

var normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#000000"))
var cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#dfdfdf"))
var fogOfWarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#9f9f9f"))
var highlightColor = lipgloss.Color("#dfdf00")

type keyMap struct {
	MoveCursor key.Binding
	Enter      key.Binding
	Quit       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.MoveCursor, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Quit},
	}
}

var keys = keyMap{
	// MoveCursor is currently ONLY used to render help a message for the cursor
	MoveCursor: key.NewBinding(
		key.WithKeys(""),
		key.WithHelp("←/↑/↓/→", "move cursor"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
}

type UIState int

const (
	UIStateWaitingForInput UIState = iota
	UIStatePickingAction
)

type Log struct {
	message        string
	hideNextUpdate bool
}

type model struct {
	uiState      UIState
	gameState    GameState
	cursorX      int
	cursorY      int
	selectedUnit *Unit
	help         help.Model
	keys         keyMap
	log          Log
	screenWidth  int
	screenHeight int
}

func initialModel() model {
	return model{
		uiState:      UIStateWaitingForInput,
		gameState:    initialGameState(),
		cursorX:      5,
		cursorY:      7,
		selectedUnit: nil,
		help:         help.New(),
		keys:         keys,
		log: Log{
			message:        "",
			hideNextUpdate: true,
		},
		screenWidth:  80,
		screenHeight: 24,
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Civ Clone")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.log.message = ""
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursorY > 0 {
				m.cursorY -= 1
			}
		case "down":
			if m.cursorY < mapSizeY-1 {
				m.cursorY += 1
			}
		case "left":
			if m.cursorX > 0 {
				m.cursorX -= 1
			}
		case "right":
			if m.cursorX < mapSizeX-1 {
				m.cursorX += 1
			}
		case "enter":
			switch m.uiState {
			case UIStateWaitingForInput:
				unitOnTile := m.gameState.getUnitOnTile(m.cursorX, m.cursorY)
				if unitOnTile != nil && m.gameState.tileMap[m.cursorY][m.cursorX].discoveredByPlayer() {
					m.selectedUnit = unitOnTile
					m.gameState.setValidMoveTilesForUnit(m.selectedUnit)
					m.uiState = UIStatePickingAction
				}

				if m.gameState.tileMap[m.cursorY][m.cursorX].feature == FeatureCrop {
					m.gameState.createFarm(m.cursorX, m.cursorY)
				}
			case UIStatePickingAction:
				if m.selectedUnit != nil {
					if m.cursorX == m.selectedUnit.positionX && m.cursorY == m.selectedUnit.positionY {
						if m.gameState.tileMap[m.cursorY][m.cursorX].feature == FeatureVillage {
							m.gameState.captureVillageAtPositionWithUnit(m.cursorX, m.cursorY, m.selectedUnit)
						} else if m.gameState.tileMap[m.cursorY][m.cursorX].feature == FeatureCity && m.gameState.tileMap[m.cursorY][m.cursorX].city.owner != m.selectedUnit.owner {
							m.gameState.captureCityAtPositionWithUnit(m.cursorX, m.cursorY, m.selectedUnit)
						}
					} else if m.gameState.tileMap[m.cursorY][m.cursorX].validForAction {
						unitOnTile := m.gameState.getUnitOnTile(m.cursorX, m.cursorY)
						if unitOnTile != nil && unitOnTile.owner != m.selectedUnit.owner {
							m.gameState.runCombatBetween(m.selectedUnit, unitOnTile)
						} else {
							m.gameState.moveUnitTo(m.selectedUnit, m.cursorX, m.cursorY)
							m.log.message = "You move the Warrior."
						}
					}
				}
				m.selectedUnit = nil
				m.uiState = UIStateWaitingForInput
			}
		case "esc":
			switch m.uiState {
			case UIStatePickingAction:
				m.selectedUnit = nil
				m.uiState = UIStateWaitingForInput
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	}

	m.setContextAwareHelpMessages()

	return m, nil
}

func (m *model) setContextAwareHelpMessages() {
	m.keys.Enter.SetEnabled(true)
	switch m.uiState {
	case UIStateWaitingForInput:
		unitOnTile := m.gameState.getUnitOnTile(m.cursorX, m.cursorY)
		if unitOnTile != nil && m.gameState.tileMap[m.cursorY][m.cursorX].discoveredByPlayer() {
			m.keys.Enter.SetHelp("enter", "select unit")
		} else {
			m.keys.Enter.SetEnabled(false)
		}
	case UIStatePickingAction:
		if m.selectedUnit.positionX == m.cursorX && m.selectedUnit.positionY == m.cursorY {
			if m.gameState.tileMap[m.cursorY][m.cursorX].feature == FeatureVillage {
				m.keys.Enter.SetHelp("enter", "capture village")
			} else if m.gameState.tileMap[m.cursorY][m.cursorX].feature == FeatureCity && m.gameState.tileMap[m.cursorY][m.cursorX].city.owner != m.selectedUnit.owner {
				m.keys.Enter.SetHelp("enter", "capture city")
			}
		} else {
			if m.gameState.tileMap[m.cursorY][m.cursorX].validForAction {
				unitOnTile := m.gameState.getUnitOnTile(m.cursorX, m.cursorY)
				if unitOnTile != nil && m.selectedUnit.owner != unitOnTile.owner {
					m.keys.Enter.SetHelp("enter", "attack unit")
				} else {
					// units can't move onto tiles with other units from the same civ
					m.keys.Enter.SetHelp("enter", "move unit")
				}
			} else {
				m.keys.Enter.SetHelp("enter", "unselect unit")
			}
		}
	}
}

func (m model) View() string {
	s := ""
	for i := 0; i < mapSizeY; i++ {
		for j := 0; j < mapSizeX; j++ {
			textStyle := normalStyle
			var tileChar byte

			if m.gameState.tileMap[i][j].discoveredByPlayer() {
				if m.cursorX == j && m.cursorY == i {
					textStyle = cursorStyle
				} else if m.gameState.tileMap[i][j].city != nil {
					textStyle = m.gameState.tileMap[i][j].city.owner.tileStyle
				}

				tileChar = TileChars[m.gameState.tileMap[i][j].tileType]
				unitOnTile := m.gameState.getUnitOnTile(j, i)
				if unitOnTile != nil {
					tileChar = UnitChars[unitOnTile.unitType]
					if m.cursorX != j || m.cursorY != i {
						textStyle = unitOnTile.owner.tileStyle
						if unitOnTile == m.selectedUnit {
							textStyle = textStyle.Foreground(highlightColor)
						}
					}
				} else {
					if m.gameState.tileMap[i][j].feature != FeatureNone {
						tileChar = FeatureChars[m.gameState.tileMap[i][j].feature]
					}
				}
			} else {
				if m.cursorX == j && m.cursorY == i {
					textStyle = cursorStyle
				} else {
					textStyle = fogOfWarStyle
				}

				tileChar = '?'
			}

			if m.uiState == UIStatePickingAction && m.gameState.tileMap[i][j].validForAction && (m.cursorX != j || m.cursorY != i) {
				textStyle = textStyle.Foreground(highlightColor)
			}

			s += textStyle.Render(string(tileChar))
		}
		if i < mapSizeY-1 {
			s += "\n"
		}
	}

	s = lipgloss.JoinHorizontal(lipgloss.Top, s, " ", m.getInfoPanel())
	s = lipgloss.JoinVertical(lipgloss.Left, s, m.getCursorHint())

	logString := ""
	if m.log.message != "" {
		logString = m.getLog()
	}
	s = lipgloss.Place(m.screenWidth, m.screenHeight-2, lipgloss.Left, lipgloss.Top, lipgloss.JoinVertical(lipgloss.Left, s, logString))
	s = lipgloss.JoinVertical(lipgloss.Left, s, lipgloss.PlaceVertical(2, lipgloss.Bottom, m.help.View(m.keys)))

	return s
}

func (m model) getInfoPanel() string {
	cursorTile := m.gameState.tileMap[m.cursorY][m.cursorX]
	s := "Info\n"

	tileString := cursorTile.getDescription()
	if cursorTile.discoveredByPlayer() {
		unitOnTile := m.gameState.getUnitOnTile(m.cursorX, m.cursorY)
		if unitOnTile != nil {
			isSelected := unitOnTile == m.selectedUnit
			tileString = unitOnTile.getDescription(isSelected) + tileString
		}
	}

	s += tileString

	return s
}

func (m model) getLog() string {
	return fmt.Sprintf("[!] %s", m.log.message)
}

func (m model) getCursorHint() string {
	cursorTile := m.gameState.tileMap[m.cursorY][m.cursorX]
	s := ""

	if cursorTile.discoveredByPlayer() {
		unitOnTile := m.gameState.getUnitOnTile(m.cursorX, m.cursorY)
		if unitOnTile != nil {
			s += unitOnTile.owner.tileStyle.Render(unitOnTile.name)
			if m.selectedUnit != nil && unitOnTile == m.selectedUnit {
				s += " (Selected)"
			}
			s += ", "
		}
		if cursorTile.city != nil {
			styledCityName := cursorTile.city.owner.tileStyle.Render(cursorTile.city.name)
			s = fmt.Sprintf("%s - %s", styledCityName, s)
			if cursorTile.city.positionX == m.cursorX && cursorTile.city.positionY == m.cursorY {
				s += "City, "
			}
		}
		switch cursorTile.tileType {
		case TilePlains:
			s += "Plains"
		case TileMountain:
			s += "Mountain"
		}
		switch cursorTile.feature {
		case FeatureVillage:
			s += ", Village"
		case FeatureCrop:
			s += ", Crop"
		case FeatureFarm:
			s += ", Farm"
		}
	} else {
		s += "Unexplored"
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
