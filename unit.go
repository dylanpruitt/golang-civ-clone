package main

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
