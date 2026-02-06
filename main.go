package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

var normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#000000"))
var cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#dfdfdf"))

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
	cityNames []string
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

type model struct {
	hello        string
	tileMap      [mapSizeY][mapSizeX]Tile
	cursorX      int
	cursorY      int
	civs         []Civ
	TEMPcivIndex int
}

func initialModel() model {
	m := model{
		hello:   "Hello World",
		tileMap: [mapSizeY][mapSizeX]Tile{},
		cursorX: 5,
		cursorY: 7,
		civs: []Civ{
			Civ{
				name:      "TestCiv",
				tileStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#6f0000")),
			},
			Civ{
				name:      "TestCiv2",
				tileStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#001f5f")),
			},
		},
		TEMPcivIndex: 0,
	}
	m.tileMap[5][9].tileType = TileMountain
	m.tileMap[6][7].feature = FeatureVillage

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
			m.createCity(m.civs[m.TEMPcivIndex], m.cursorX, m.cursorY)
			m.TEMPcivIndex = 1 - m.TEMPcivIndex
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		return m, nil
	}

	return m, nil
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
			if x+j >= 0 && x+j < mapSizeX && y+i >= 0 && y+i <= mapSizeY && m.tileMap[y+i][x+j].city == nil {
				m.tileMap[y+i][x+j].city = &city
			}
		}
	}
}

func (m model) View() string {
	s := m.hello + "\n"
	for i := 0; i < mapSizeY; i++ {
		for j := 0; j < mapSizeX; j++ {
			textStyle := normalStyle
			if m.cursorX == j && m.cursorY == i {
				textStyle = cursorStyle
			} else if m.tileMap[i][j].city != nil {
				textStyle = m.tileMap[i][j].city.owner.tileStyle
			}
			tileChar := TileChars[m.tileMap[i][j].tileType]
			if m.tileMap[i][j].feature != FeatureNone {
				tileChar = FeatureChars[m.tileMap[i][j].feature]
			}
			s += textStyle.Render(string(tileChar))
		}
		s += "\n"
	}
	s += m.getCursorHint()
	s += "\nPress q to quit.\n"

	return s
}

func (m model) getCursorHint() string {
	cursorTile := m.tileMap[m.cursorY][m.cursorX]
	s := ""
	if cursorTile.city != nil {
		s += fmt.Sprintf("%s - ", cursorTile.city.name)
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

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
