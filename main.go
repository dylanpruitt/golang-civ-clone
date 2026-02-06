package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

var normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dfdfdf")).Background(lipgloss.Color("#000000"))
var cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#dfdfdf"))
var cursorLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#9f9f9f"))

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
    FeatureCrop
)

const FeatureChars string = " +,"

type Tile struct {
	tileType TileType
	feature  Feature
}

type model struct {
	hello   string
	tileMap [mapSizeY][mapSizeX]Tile
	cursorX int
	cursorY int
}

func initialModel() model {
	m := model{
		hello:   "Hello World",
		tileMap: [mapSizeY][mapSizeX]Tile{},
		cursorX: 5,
		cursorY: 7,
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
			if m.cursorY < mapSizeY - 1 {
				m.cursorY += 1
			}
		case "left":
			if m.cursorX > 0 {
				m.cursorX -= 1
			}
		case "right":
			if m.cursorX < mapSizeX - 1 {
				m.cursorX += 1
			}
		case "enter":
			m.tileMap[m.cursorY][m.cursorX].tileType = TileMountain
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	s := m.hello + "\n"
	for i := 0; i < mapSizeY; i++ {
		for j := 0; j < mapSizeX; j++ {
			textStyle := normalStyle
			if m.cursorX == j && m.cursorY == i {
				textStyle = cursorStyle
			} else if m.cursorX == j || m.cursorY == i {
				textStyle = cursorLineStyle
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
