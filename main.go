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
var highlightColor = lipgloss.Color("#dfdf00")

type keyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Quit},
	}
}

var keys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
}

const mapSizeX int = 30
const mapSizeY int = 15

type TileType int

const (
	TilePlains TileType = iota
	TileMountain
)

const TileChars string = ".^"

type Feature int

const (
	FeatureNone Feature = iota
	FeatureVillage
	FeatureCity
	FeatureCrop
)

const FeatureChars string = " +@,"

type Civ struct {
	name      string
	tileStyle lipgloss.Style
}

type City struct {
	name       string
	population int
	owner      Civ
	positionX  int
	positionY  int
}

type Tile struct {
	tileType TileType
	feature  Feature
	city     *City
}

type UnitType int

const (
	UnitWarrior UnitType = iota
)

const UnitChars string = "w"

type Unit struct {
	name      string
	unitType  UnitType
	positionX int
	positionY int
	owner     *Civ
}

func (u *Unit) moveTo(x, y int) {
	u.positionX = x
	u.positionY = y
}

type UIState int

const (
	UIStateWaitingForInput UIState = iota
	UIStatePickingAction
)

type model struct {
	uiState      UIState
	tileMap      [mapSizeY][mapSizeX]Tile
	cursorX      int
	cursorY      int
	civs         []Civ
	units        []Unit
	selectedUnit *Unit
	help         help.Model
	keys         keyMap
}

func initialModel() model {
	civ0 := Civ{
		name:      "TestCiv",
		tileStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#6f0000")),
	}
	civ1 := Civ{
		name:      "TestCiv2",
		tileStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#001f5f")),
	}
	m := model{
		uiState: UIStateWaitingForInput,
		tileMap: [mapSizeY][mapSizeX]Tile{},
		cursorX: 5,
		cursorY: 7,
		civs: []Civ{
			civ0,
			civ1,
		},
		units: []Unit{
			Unit{
				name:      "Warrior",
				unitType:  UnitWarrior,
				positionX: 6,
				positionY: 6,
				owner:     &civ0,
			},
			Unit{
				name:      "Warrior",
				unitType:  UnitWarrior,
				positionX: 8,
				positionY: 6,
				owner:     &civ1,
			},
		},
		selectedUnit: nil,
		help:         help.New(),
		keys:         keys,
	}
	m.tileMap[5][9].tileType = TileMountain
	m.tileMap[6][7].feature = FeatureVillage
	m.tileMap[8][8].feature = FeatureVillage
	m.tileMap[11][22].feature = FeatureVillage

	return m
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Grocery List")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
				unitOnTile := m.getUnitOnTile(m.cursorX, m.cursorY)
				if unitOnTile != nil {
					m.selectedUnit = unitOnTile
					m.uiState = UIStatePickingAction
				}
			case UIStatePickingAction:
				if m.selectedUnit != nil {
					if m.cursorX == m.selectedUnit.positionX && m.cursorY == m.selectedUnit.positionY && m.tileMap[m.cursorY][m.cursorX].feature == FeatureVillage {
						m.captureVillageAtPositionWithUnit(m.cursorX, m.cursorY, m.selectedUnit)
					} else {
						m.selectedUnit.moveTo(m.cursorX, m.cursorY)
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
		unitOnTile := m.getUnitOnTile(m.cursorX, m.cursorY)
		if unitOnTile != nil {
			m.keys.Enter.SetHelp("enter", "select unit")
		} else {
			m.keys.Enter.SetEnabled(false)
		}
	case UIStatePickingAction:
		if m.selectedUnit.positionX == m.cursorX && m.selectedUnit.positionY == m.cursorY && m.tileMap[m.cursorY][m.cursorX].feature == FeatureVillage {
			m.keys.Enter.SetHelp("enter", "capture village")
		} else {
			m.keys.Enter.SetHelp("enter", "move unit")
		}
	}
}

func (m *model) captureVillageAtPositionWithUnit(x, y int, u *Unit) {
	if m.tileMap[m.cursorY][m.cursorX].feature != FeatureVillage {
		return
	}
	m.createCity(*u.owner, x, y)
}

func (m *model) createCity(civ Civ, x, y int) {
	name := "Rome"
	if civ.name == "TestCiv2" {
		name = "London"
	}
	city := City{
		name:       name,
		population: 1,
		owner:      civ,
		positionX:  x,
		positionY:  y,
	}

	m.tileMap[y][x].feature = FeatureCity
	m.cultureBombTile(city, x, y)
}

func (m *model) cultureBombTile(city City, x, y int) {
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if x+j >= 0 && x+j < mapSizeX && y+i >= 0 && y+i < mapSizeY && m.tileMap[y+i][x+j].city == nil {
				m.tileMap[y+i][x+j].city = &city
			}
		}
	}
}

func (m model) View() string {
	s := ""
	for i := 0; i < mapSizeY; i++ {
		for j := 0; j < mapSizeX; j++ {
			textStyle := normalStyle
			if m.cursorX == j && m.cursorY == i {
				textStyle = cursorStyle
			} else if m.tileMap[i][j].city != nil {
				textStyle = m.tileMap[i][j].city.owner.tileStyle
			}
			tileChar := TileChars[m.tileMap[i][j].tileType]
			unitOnTile := m.getUnitOnTile(j, i)
			if unitOnTile != nil {
				tileChar = UnitChars[unitOnTile.unitType]
				if m.cursorX != j || m.cursorY != i {
					textStyle = unitOnTile.owner.tileStyle
					if unitOnTile == m.selectedUnit {
						textStyle = textStyle.Foreground(highlightColor)
					}
				}
			} else {
				if m.tileMap[i][j].feature != FeatureNone {
					tileChar = FeatureChars[m.tileMap[i][j].feature]
				}
			}

			s += textStyle.Render(string(tileChar))
		}
		s += "\n"
	}
	s += m.getCursorHint()

	s += "\nPress q to quit.\n"
	s += m.help.View(m.keys)

	return s
}

func (m model) getCursorHint() string {
	cursorTile := m.tileMap[m.cursorY][m.cursorX]
	s := ""

	unitOnTile := m.getUnitOnTile(m.cursorX, m.cursorY)
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
	}

	return s
}

func (m model) getUnitOnTile(x, y int) *Unit {
	unitIndex := -1
	for i, u := range m.units {
		if u.positionX == x && u.positionY == y {
			unitIndex = i
		}
	}
	if unitIndex > -1 {
		return &m.units[unitIndex]
	}
	return nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
