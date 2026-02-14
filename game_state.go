package main

import (
	"math"
	"slices"

	lipgloss "github.com/charmbracelet/lipgloss"
)

const mapSizeX int = 30
const mapSizeY int = 15

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

type GameState struct {
	tileMap [mapSizeY][mapSizeX]Tile
	civs    []Civ
	units   []Unit
}

func initialGameState() GameState {
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
	g := GameState{
		tileMap: [mapSizeY][mapSizeX]Tile{},
		civs: []Civ{
			civ0,
			civ1,
		},
		units: []Unit{
			makeWarrior(6, 6, &civ0),
			makeWarrior(7, 7, &civ0),
			makeWarrior(8, 6, &civ1),
		},
	}
	g.tileMap[5][9].tileType = TileMountain
	g.tileMap[7][7].tileType = TileMountain
	g.tileMap[7][10].tileType = TileMountain
	g.tileMap[4][2].tileType = TileMountain
	g.tileMap[3][4].tileType = TileMountain
	g.tileMap[4][4].tileType = TileMountain
	g.tileMap[4][5].tileType = TileMountain
	g.tileMap[3][6].tileType = TileMountain
	g.tileMap[4][6].tileType = TileMountain
	g.tileMap[6][7].feature = FeatureVillage
	g.tileMap[5][5].hasRoad = true
	g.tileMap[5][4].hasRoad = true
	g.tileMap[4][3].hasRoad = true
	g.tileMap[3][5].hasRoad = true
	g.tileMap[3][3].hasRoad = true
	g.tileMap[2][4].hasRoad = true
	g.tileMap[2][3].hasRoad = true
	g.tileMap[2][2].hasRoad = true
	g.tileMap[8][8].feature = FeatureVillage
	g.tileMap[8][9].feature = FeatureCrop
	g.tileMap[8][10].feature = FeatureCrop
	g.tileMap[8][11].feature = FeatureCrop
	g.tileMap[8][12].feature = FeatureCrop
	g.tileMap[11][22].feature = FeatureVillage
	g.revealTilesFromPos(6, 6, 1, &civ0)

	return g
}

func (g *GameState) createFarm(x, y int) {
	if g.tileMap[y][x].city == nil {
		return
	}
	g.tileMap[y][x].feature = FeatureFarm
	g.cultureBombTile(*g.tileMap[y][x].city, x, y)
	g.revealTilesFromPos(x, y, 1, g.tileMap[y][x].city.owner)
}

func (g *GameState) captureVillageAtPositionWithUnit(x, y int, u *Unit) {
	if g.tileMap[y][x].feature != FeatureVillage {
		return
	}
	g.createCity(u.owner, x, y)
}

func (g *GameState) createCity(civ *Civ, x, y int) {
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

	g.tileMap[y][x].feature = FeatureCity
    g.tileMap[y][x].hasRoad = true
	g.cultureBombTile(city, x, y)
	g.revealTilesFromPos(x, y, 2, civ)
}

func (g *GameState) moveUnitOnPathTo(u *Unit, x, y int) {
    goal := [2]int{y,x}
    start := [2]int{u.positionY,u.positionX}
    path := [][2]int{}
    currentTile := goal

    for currentTile != start {
        t := g.tileMap[currentTile[0]][currentTile[1]]
        if t.cameFrom == nil {
            return
        }
        
        path = append(path, currentTile)
        currentTile = *t.cameFrom
    }
	u.positionX = x
	u.positionY = y
    
    for i := len(path) - 1; i >= 0; i-- {
        revealRange := 1
        if g.tileMap[path[i][0]][path[i][1]].tileType == TileMountain {
            revealRange = 2
        }
        g.revealTilesFromPos(path[i][1], path[i][0], revealRange, u.owner)
    }
}

func (g *GameState) moveUnitTo(u *Unit, x, y int) {
	u.positionX = x
	u.positionY = y
	revealRange := 1
	if g.tileMap[y][x].tileType == TileMountain {
		revealRange = 2
	}
	g.revealTilesFromPos(x, y, revealRange, u.owner)
}

func (g *GameState) runCombatBetween(u *Unit, o *Unit) {
	attackerDamage := int(math.Round(float64(u.attack) * (float64(u.hp) / float64(u.maxHp))))
	defenderDamage := int(math.Round(float64(o.defense) * (float64(o.hp) / float64(o.maxHp))))

	o.hp -= attackerDamage
	if o.hp > 0 {
		u.hp -= defenderDamage
		if u.hp <= 0 {
			o.kills++
		}
	} else {
		g.moveUnitTo(u, o.positionX, o.positionY)
		u.kills++
	}

	delIndex := -1
	for i := 0; i < len(g.units); i++ {
		if &g.units[i] == u && u.hp <= 0 {
			delIndex = i
		}
	}
	if delIndex > -1 {
		slices.Delete(g.units, delIndex, delIndex+1)
	}
	for i := 0; i < len(g.units); i++ {
		if &g.units[i] == o && o.hp <= 0 {
			delIndex = i
		}
	}
	if delIndex > -1 {
		slices.Delete(g.units, delIndex, delIndex+1)
	}
}

func (g *GameState) cultureBombTile(city City, x, y int) {
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if positionInMapBounds(x+j, y+i) && g.tileMap[y+i][x+j].city == nil && tileInCityRange(x+j, y+i, city) {
				g.tileMap[y+i][x+j].city = &city
			}
		}
	}
}

func (g *GameState) captureCityAtPositionWithUnit(x, y int, u *Unit) {
	if g.tileMap[y][x].feature != FeatureCity || g.tileMap[y][x].city.owner == u.owner {
		return
	}
	g.tileMap[y][x].city.owner = u.owner
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
	totalCost float64
	baseCost  float64
}

func (g *GameState) setValidMoveTilesForUnit(u *Unit) {
	tileCosts := [mapSizeY][mapSizeX]TileCost{}
	validTiles := [][2]int{}
	// reset tile costs, set tile unit is on to valid
	for i := 0; i < mapSizeY; i++ {
		for j := 0; j < mapSizeX; j++ {
			if i == u.positionY && j == u.positionX {
				tileCosts[i][j].totalCost = 0.0
				tileCosts[i][j].baseCost = 0.0
				validTiles = append(validTiles, [2]int{u.positionX, u.positionY})
			} else {
				tileCosts[i][j].totalCost = 99.0
				tileCosts[i][j].baseCost = g.getTileMoveCost(j, i, u)
				g.tileMap[i][j].validForAction = false
			}
            g.tileMap[i][j].cameFrom = nil
		}
	}

	t := 0
	for t < len(validTiles) {
		tilePos := validTiles[t]
		thisTc := tileCosts[tilePos[1]][tilePos[0]]
		for i := -1; i < 2; i++ {
			for j := -1; j < 2; j++ {
				if positionInMapBounds(tilePos[0]+j, tilePos[1]+i) {
					tc := &tileCosts[tilePos[1]+i][tilePos[0]+j]
					oldTotalCost := tc.totalCost
					newTotalCost := thisTc.totalCost + tc.baseCost
					if newTotalCost < oldTotalCost {
						tc.totalCost = newTotalCost
					}

					newTilePos := [2]int{tilePos[0] + j, tilePos[1] + i}
					if tc.totalCost <= u.movePoints && !slices.Contains(validTiles, newTilePos) {
						validTiles = append(validTiles, newTilePos)
                        g.tileMap[newTilePos[1]][newTilePos[0]].cameFrom = &[2]int{tilePos[1],tilePos[0]}
					}
				}
			}
		}
		t++
	}

	for i := 0; i < len(validTiles); i++ {
		t := validTiles[i]
		g.tileMap[t[1]][t[0]].validForAction = true
	}
}

func (g *GameState) getTileMoveCost(x, y int, u *Unit) float64 {
	const IMPASSABLE float64 = 99.0
	unitOnTile := g.getUnitOnTile(x, y)
	if unitOnTile != nil && unitOnTile.owner == u.owner && (u.positionX != unitOnTile.positionX || u.positionY != unitOnTile.positionY) {
		return IMPASSABLE
	} else {
		if g.tileMap[u.positionY][u.positionX].hasRoad && g.tileMap[y][x].hasRoad {
			return 0.25
		}

		switch g.tileMap[y][x].tileType {
		case TileMountain:
			// TODO handle case where Mountains are impassable without Climbing tech
			return 2.0
		default:
			return 1.0
		}
	}
}

func (g *GameState) revealTilesFromPos(x, y, revealRange int, c *Civ) {
	for i := -revealRange; i < revealRange+1; i++ {
		for j := -revealRange; j < revealRange+1; j++ {
			if positionInMapBounds(x+j, y+i) {
				g.tileMap[y+i][x+j].discoveredBy = append(g.tileMap[y+i][x+j].discoveredBy, c.id)
			}
		}
	}
}

func positionInMapBounds(x, y int) bool {
	return x >= 0 && x < mapSizeX && y >= 0 && y < mapSizeY
}

func (g GameState) getUnitOnTile(x, y int) *Unit {
	for i := 0; i < len(g.units); i++ {
		if g.units[i].positionX == x && g.units[i].positionY == y {
			return &g.units[i]
		}
	}
	return nil
}
