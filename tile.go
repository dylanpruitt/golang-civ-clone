package main

import (
	"slices"
)

type Feature int

const (
	FeatureNone Feature = iota
	FeatureVillage
	FeatureCity
	FeatureCrop
	FeatureFarm
)

const FeatureChars string = " +@,#"

type TileType int

const (
	TilePlains TileType = iota
	TileMountain
)

const TileChars string = ".^"

type Tile struct {
	tileType       TileType
	feature        Feature
	hasRoad        bool
	city           *City
	validForAction bool
	discoveredBy   []int
    cameFrom *[2]int
}

func (t Tile) discoveredByPlayer() bool {
	// assumes player ID will always be 0
	return slices.Contains(t.discoveredBy, 0)
}

func (t Tile) getDescription() string {
	s := ""

	if t.discoveredByPlayer() {
		if t.city != nil {
			styledCityName := t.city.owner.tileStyle.Render(t.city.name)
			s += styledCityName + "\n"
		}
		moveCostString := ""
		switch t.tileType {
		case TilePlains:
			s += "Plains\n"
			moveCostString = "  1 movement cost\n"
		case TileMountain:
			s += "Mountain\n"
			moveCostString = "  2 movement cost\n"
		}
		if t.hasRoad {
			s += "Road\n"
			moveCostString = "  0.25 movement cost\n"
		}
		s += moveCostString

		switch t.feature {
		case FeatureVillage:
			s += "Village\n  Move a unit here to capture"
		case FeatureCrop:
			s += "Crop\n  Can build a Farm here"
		case FeatureFarm:
			s += "Farm\n"
		}
	} else {
		s += "??? - Unexplored Tile"
	}

	return s
}

func (t Tile) getCursorHint() string {
	s := ""
	switch t.tileType {
	case TilePlains:
		s += "Plains"
	case TileMountain:
		s += "Mountain"
	}
	if t.hasRoad {
		s += ", Road"
	}
	switch t.feature {
	case FeatureVillage:
		s += ", Village"
	case FeatureCrop:
		s += ", Crop"
	case FeatureFarm:
		s += ", Farm"
	}

	return s
}
