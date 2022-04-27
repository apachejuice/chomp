package chomp

import (
	"fmt"
	"math"
)

func contains(a []Position, e Position) bool {
	for _, v := range a {
		if e == v {
			return true
		}
	}

	return false
}

// Represents a position in a chess board.
type Position struct {
	// The x coordinate
	X int8 `json:"x"`
	// The y coordinate
	Y int8 `json:"y"`
}

func (p Position) String() string {
	return fmt.Sprintf("[x=%d y=%d]", p.X, p.Y)
}

// Represents an erroneous position.
var ErrorPos = Position{X: -1, Y: -1}

// Creates a new position.
func NewPos(x, y int8) Position {
	if x < 0 || y < 0 {
		return ErrorPos
	}

	return Position{X: x, Y: y}
}

// Represents a location on the chessboard: a corner, a side, a middle or out of bounds.
type Location int8

const (
	LocationCorner      = 1
	LocationSide        = 2
	LocationMiddle      = 3
	LocationOutOfBounds = 4
)

/////////////////////////////////////////
///// Position comparison functions /////
// These functions do no bound checks. //
/////////////////////////////////////////

var corners = []Position{
	NewPos(0, 0), NewPos(0, 7), NewPos(7, 0), NewPos(7, 7),
}

// Returns where this position is on the chessboard: in a corner, on the sides, in the middle or not in the board at all.
func (p Position) Locate() Location {
	if p == ErrorPos || (p.X > 7 || p.X < 0) || (p.Y > 7 || p.Y < 0) {
		return LocationOutOfBounds
	}

	if contains(corners, p) {
		return LocationCorner
	}

	if p.X == 7 || p.X == 0 || p.Y == 7 || p.Y == 0 {
		return LocationSide
	}

	return LocationMiddle
}

// Returns whether position a corners this position.
func (p Position) IsCornering(a Position) bool {
	return abs8(p.X-a.X) == 1 && abs8(p.Y-a.Y) == 1
}

// Returns whether position a is touching this position by one side.
func (p Position) IsAdjacent(a Position) bool {
	// So here deltaX or deltaY must be 1, but not both.
	return abs8(p.X-a.X) == 1 || abs8(p.Y-a.Y) == 1 && !p.IsCornering(a)
}

// Returns whether position a is one of the 8 positions next to this.
func (p Position) IsNextTo(a Position) bool {
	return abs8(p.X-a.X) == 1 || abs8(p.Y-a.Y) == 1
}

func (p Position) IsLinedUpWith(a Position) bool {
	return p.X == a.X || p.Y == a.Y
}

func (p Position) IsOnSameSlope(a Position) bool {
	return (p.Y-a.Y)/(p.X-a.X) == 1
}

func (p Position) IsKnightAvailable(a Position) bool {
	slope := math.Abs(float64((p.Y - a.Y)) / float64((p.X - a.X)))
	diffX := abs8(p.X - a.X)
	diffY := abs8(p.Y - a.Y)

	return (slope == 2 || slope == 0.5) && ((diffX == 1 && diffY == 2) || (diffX == 2 && diffY == 1))
}

func PositionsBetween(a, b Position) ([]Position, error) {
	if !a.IsLinedUpWith(b) && !a.IsOnSameSlope(b) || a == b {
		return nil, fmt.Errorf("the given positions were invalid")
	}

	if a.IsNextTo(b) {
		return []Position{}, nil
	}

	if a.IsLinedUpWith(b) {
		// I feel like these repetitive pieces of code should be cleaner but i don't know how to.
		if a.X == b.X {
			start, end := minmax(a.Y, b.Y)
			positions := make([]Position, end-start-1)
			for i := int8(0); i < end-start-1; i++ {
				positions[i] = Position{X: a.X, Y: start + i + 1}
			}

			return positions, nil
		} else {
			start, end := minmax(a.X, b.X)
			positions := make([]Position, end-start-1)
			for i := int8(0); i < end-start-1; i++ {
				positions[i] = Position{X: start + i + 1, Y: a.Y}
			}

			return positions, nil
		}

	} else if a.IsOnSameSlope(b) {
		slope := (a.Y - b.Y) / (a.X - b.X)
		start, end := minmax(a.X, b.X)
		eb := a.Y - slope*a.X
		positions := make([]Position, end-start-1)

		// +1 to not count the starting position
		for i := start + 1; i < end; i++ {
			y := float64(slope*i + eb)
			rounded := int8(math.Floor(y + 0.5))
			positions[i-start-1] = Position{X: i, Y: rounded}
		}

		return positions, nil
	}

	return nil, nil
}
