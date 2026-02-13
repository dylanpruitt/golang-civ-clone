package main

import "fmt"

type UnitType int

const (
	UnitWarrior UnitType = iota
)

const UnitChars string = "w"

type Unit struct {
	name       string
	unitType   UnitType
	hp         int
	maxHp      int
	attack     int
	defense    int
	kills      int
	positionX  int
	positionY  int
	owner      *Civ
	movePoints int
}

func makeWarrior(x, y int, owner *Civ) Unit {
	return Unit{
		name:       "Warrior",
		unitType:   UnitWarrior,
		hp:         4,
		maxHp:      4,
		attack:     3,
		defense:    3,
		positionX:  x,
		positionY:  y,
		owner:      owner,
		movePoints: 2,
	}
}

func (u Unit) getDescription(isSelected bool) string {
	s := fmt.Sprintf("%s %d/%d HP", u.owner.tileStyle.Render(u.name), u.hp, u.maxHp)
	if isSelected {
		s += " (Selected)"
	}
	s += fmt.Sprintf("\n  %d ATK %d DEF %d MOVE\n", u.attack, u.defense, u.movePoints)
	if u.kills > 0 {
		s += fmt.Sprintf("  %d/2 kills to promotion\n", u.kills)
	}
	return s
}
