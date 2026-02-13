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
	city           *City
	validForAction bool
	discoveredBy   []int
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
		switch t.tileType {
		case TilePlains:
			s += "Plains\n  1 movement cost\n"
		case TileMountain:
			s += "Mountain\n  2 movement cost\n"
		}
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
