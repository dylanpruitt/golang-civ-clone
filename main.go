package main

import (
	"fmt"
	"os"
	"slices"

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
	FeatureFarm
)

const FeatureChars string = " +@,#"

type Civ struct {
	name      string
	id        int
	tileStyle lipgloss.Style
}

type City struct {
	name       string
	population int
	owner      *Civ
	positionX  int
	positionY  int
}

type Tile struct {
	tileType       TileType
	feature        Feature
	city           *City
	validForAction bool
}

type UnitType int

const (
	UnitWarrior UnitType = iota
)

const UnitChars string = "w"

type Unit struct {
	name       string
	unitType   UnitType
	positionX  int
	positionY  int
	owner      *Civ
	movePoints int
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

type Log struct {
	message        string
	hideNextUpdate bool
}

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
	log          Log
}

func initialModel() model {
	civ0 := Civ{
		name:      "TestCiv",
		id:        0,
		tileStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#6f0000")),
	}
	civ1 := Civ{
		name:      "TestCiv2",
		id:        1,
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
				name:       "Warrior",
				unitType:   UnitWarrior,
				positionX:  6,
				positionY:  6,
				owner:      &civ0,
				movePoints: 2,
			},
			Unit{
				name:       "Warrior",
				unitType:   UnitWarrior,
				positionX:  8,
				positionY:  6,
				owner:      &civ1,
				movePoints: 2,
			},
		},
		selectedUnit: nil,
		help:         help.New(),
		keys:         keys,
		log: Log{
			message:        "",
			hideNextUpdate: true,
		},
	}
	m.tileMap[5][9].tileType = TileMountain
	m.tileMap[7][7].tileType = TileMountain
	m.tileMap[7][10].tileType = TileMountain
	m.tileMap[4][2].tileType = TileMountain
	m.tileMap[4][3].tileType = TileMountain
	m.tileMap[4][4].tileType = TileMountain
	m.tileMap[6][7].feature = FeatureVillage
	m.tileMap[8][8].feature = FeatureVillage
	m.tileMap[8][9].feature = FeatureCrop
	m.tileMap[8][10].feature = FeatureCrop
	m.tileMap[8][11].feature = FeatureCrop
	m.tileMap[8][12].feature = FeatureCrop
	m.tileMap[11][22].feature = FeatureVillage

	return m
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Civ Clone")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.log.message = ""
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

				if m.tileMap[m.cursorY][m.cursorX].feature == FeatureCrop {
					m.createFarm(m.cursorX, m.cursorY)
				}
			case UIStatePickingAction:
				if m.selectedUnit != nil {
					if m.cursorX == m.selectedUnit.positionX && m.cursorY == m.selectedUnit.positionY {
						if m.tileMap[m.cursorY][m.cursorX].feature == FeatureVillage {
							m.captureVillageAtPositionWithUnit(m.cursorX, m.cursorY, m.selectedUnit)
						} else if m.tileMap[m.cursorY][m.cursorX].feature == FeatureCity && m.tileMap[m.cursorY][m.cursorX].city.owner != m.selectedUnit.owner {
							m.captureCityAtPositionWithUnit(m.cursorX, m.cursorY, m.selectedUnit)
						}
					} else {
						m.selectedUnit.moveTo(m.cursorX, m.cursorY)
						m.log.message = "You move the Warrior."
					}
				}
				m.selectedUnit = nil
				m.uiState = UIStateWaitingForInput
			}
		case "t":
			if m.selectedUnit != nil {
				m.setValidMoveTilesForUnit(m.selectedUnit)
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
		if m.selectedUnit.positionX == m.cursorX && m.selectedUnit.positionY == m.cursorY {
			if m.tileMap[m.cursorY][m.cursorX].feature == FeatureVillage {
				m.keys.Enter.SetHelp("enter", "capture village")
			} else if m.tileMap[m.cursorY][m.cursorX].feature == FeatureCity && m.tileMap[m.cursorY][m.cursorX].city.owner != m.selectedUnit.owner {
				m.keys.Enter.SetHelp("enter", "capture city")
			}
		} else {
			m.keys.Enter.SetHelp("enter", "move unit")
		}
	}
}

func (m *model) createFarm(x, y int) {
	if m.tileMap[m.cursorY][m.cursorX].city == nil {
		return
	}
	m.tileMap[m.cursorY][m.cursorX].feature = FeatureFarm
	m.cultureBombTile(*m.tileMap[m.cursorY][m.cursorX].city, x, y)
}

func (m *model) captureVillageAtPositionWithUnit(x, y int, u *Unit) {
	if m.tileMap[m.cursorY][m.cursorX].feature != FeatureVillage {
		return
	}
	m.createCity(u.owner, x, y)
}

func (m *model) createCity(civ *Civ, x, y int) {
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
			if x+j >= 0 && x+j < mapSizeX && y+i >= 0 && y+i < mapSizeY && m.tileMap[y+i][x+j].city == nil && tileInCityRange(x+j, y+i, city) {
				m.tileMap[y+i][x+j].city = &city
			}
		}
	}
}

func (m *model) captureCityAtPositionWithUnit(x, y int, u *Unit) {
	if m.tileMap[m.cursorY][m.cursorX].feature != FeatureCity || m.tileMap[m.cursorY][m.cursorX].city.owner == u.owner {
		return
	}
	m.tileMap[m.cursorY][m.cursorX].city.owner = u.owner
}

func tileInCityRange(x, y int, city City) bool {
	xDist := city.positionX - x
	if xDist < 0 {
		xDist *= -1
	}
	yDist := city.positionY - y
	if yDist < 0 {
		yDist *= -1
	}
	return xDist <= 3 && yDist <= 3
}

type TileCost struct {
	totalCost int
	baseCost  int
}

func (m *model) setValidMoveTilesForUnit(u *Unit) {
	tileCosts := [mapSizeY][mapSizeX]TileCost{}
	validTiles := [][2]int{}
	// reset tile costs, set tile unit is on to valid
	for i := 0; i < mapSizeY; i++ {
		for j := 0; j < mapSizeX; j++ {
			if i == u.positionY && j == u.positionX {
				tileCosts[i][j].totalCost = 0
				tileCosts[i][j].baseCost = 0
				validTiles = append(validTiles, [2]int{u.positionX, u.positionY})
			} else {
				tileCosts[i][j].totalCost = 99
				tileCosts[i][j].baseCost = m.getTileMoveCost(j, i, u)
				m.tileMap[i][j].validForAction = false
			}
		}
	}

	t := 0
	for t < len(validTiles) {
		tilePos := validTiles[t]
		thisTc := tileCosts[tilePos[1]][tilePos[0]]
		for i := -1; i < 2; i++ {
			for j := -1; j < 2; j++ {
				if tilePos[0]+j >= 0 && tilePos[0]+j < mapSizeX && tilePos[1]+i >= 0 && tilePos[1]+i < mapSizeY {
					tc := &tileCosts[tilePos[1]+i][tilePos[0]+j]
					oldTotalCost := tc.totalCost
					newTotalCost := thisTc.totalCost + tc.baseCost
					if newTotalCost < oldTotalCost {
						tc.totalCost = newTotalCost
					}

					tilePos := [2]int{tilePos[0] + j, tilePos[1] + i}
					if tc.totalCost <= u.movePoints && !slices.Contains(validTiles, tilePos) {
						validTiles = append(validTiles, tilePos)
					}
				}
			}
		}
		t++
	}

	for i := 0; i < len(validTiles); i++ {
		t := validTiles[i]
		m.tileMap[t[1]][t[0]].validForAction = true
	}
}

func (m *model) getTileMoveCost(x, y int, u *Unit) int {
	const IMPASSABLE int = 99
	unitOnTile := m.getUnitOnTile(x, y)
	if unitOnTile != nil && (u.positionX != unitOnTile.positionX || u.positionY != unitOnTile.positionY) {
		return IMPASSABLE
	} else {
		switch m.tileMap[y][x].tileType {
		case TileMountain:
			// TODO handle case where Mountains are impassable without Climbing tech
			return 2
		default:
			return 1
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
				if m.tileMap[i][j].validForAction {
					textStyle = textStyle.Foreground(highlightColor)
				}
			}

			s += textStyle.Render(string(tileChar))
		}
		if i < mapSizeY-1 {
			s += "\n"
		}
	}

	s = lipgloss.JoinHorizontal(lipgloss.Top, s, " ", m.getInfoPanel())
	s = lipgloss.JoinVertical(lipgloss.Left, s, m.getCursorHint())
	if m.log.message != "" {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.getLog())
	}
	s = lipgloss.JoinVertical(lipgloss.Left, s, m.help.View(m.keys))

	return s
}

func (m model) getInfoPanel() string {
	cursorTile := m.tileMap[m.cursorY][m.cursorX]
	s := "Info\n"

	unitOnTile := m.getUnitOnTile(m.cursorX, m.cursorY)
	if unitOnTile != nil {
		s += unitOnTile.owner.tileStyle.Render(unitOnTile.name)
		if m.selectedUnit != nil && unitOnTile == m.selectedUnit {
			s += " (Selected)"
		}
		// TODO replace with Unit describe function
		s += "\n  Basic unit.\n"
	}
	if cursorTile.city != nil {
		styledCityName := cursorTile.city.owner.tileStyle.Render(cursorTile.city.name)
		s += styledCityName + "\n"
	}
	switch cursorTile.tileType {
	case TilePlains:
		s += "Plains\n  1 movement cost\n"
	case TileMountain:
		s += "Mountain\n  2 movement cost\n"
	}
	switch cursorTile.feature {
	case FeatureVillage:
		s += "Village\n  Move a unit here to capture"
	case FeatureCrop:
		s += "Crop\n  Can build a Farm here"
	case FeatureFarm:
		s += "Farm\n"
	}

	return s
}

func (m model) getLog() string {
	return fmt.Sprintf("[!] %s", m.log.message)
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
	case FeatureFarm:
		s += ", Farm"
	}

	return s
}

func (m model) getUnitOnTile(x, y int) *Unit {
	for i := 0; i < len(m.units); i++ {
		if m.units[i].positionX == x && m.units[i].positionY == y {
			return &m.units[i]
		}
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
