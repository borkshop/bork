// Package parking provides an animated simulation of cars (emoji) entering and
// exiting a parking lot (line art).
package parking

import (
	"image"
	"image/draw"
	"math/rand"

	"github.com/borkshop/bork/internal/borkmark"
	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/text"
	"github.com/borkshop/bork/internal/point"
)

// Rendition of a parking lot tile:
//
//  +-- A. --+-- A. --
//  |LllarrR.|LllarrR.
//  +-- A. --+-- A. --
//  |LllarrR.|LllarrR.
//  +-- A. --+-- A. --
//
// Legend:
//  A aisle1
//  a aisle2
//  l left1
//  l left2
//  L left3
//  r right1
//  r right2
//  R right3
//
// Each tick, we project "claims" for each car onto a logical tile, where each
// value is the identifier of the car that claims right-of-way for that cell of
// the tile.
// Then, each car has an opportunity to move to a position that covers that
// tile.
//
// Invariants:
//
// 1. Cars occupy two cells of a terminal display.
// 2. Cars *logically* recognize occupying a parking space or aisle if their
// position overlaps that space even partially, thus can occupy multiple
// logical areas.
// 3. Cars have the right-of-way to remain at their current position.
// 4. All things being equal, cars with higher identifiers have right of way
// over lower cars.
// 5. Cars exiting the parking lot have right-of-way over cars seeking parking
// spaces, because cars are greedy for the space closest to the BØRK entrance.
//
// Behaviors of cars:
// 1. Cars queue to enter the parking lot.
// 2. While seeking a parking space, cars will advance down the aisle
// until they find an unoccupied space on either side of the aisle,
// choosing randomly between any two opposing spaces if they are both
// vacant.
// 3. Cars will then wait in their parking space for a random, arbitrary, and
// hopefully aesthetically pleasing interval.
// 4. Cars will then exit the parking lot, claiming right-of-way over the
// aisle when it becomes vaccant.

type tile struct {
	aisle1, aisle2, left, right int
}

const (
	tileWidth  = 9
	tileHeight = 2
)

var tileBox = image.Rectangle{
	image.ZP,
	image.Point{tileWidth, tileHeight},
}

var firstTileTemplate *display.Display
var tileTemplate *display.Display

func init() {
	tileTemplate = display.New(tileBox)
	text.Write(tileTemplate, tileBox, "┼──    ──\n│", borkmark.Yellow)
	firstTileTemplate = display.New(tileBox)
	text.Write(firstTileTemplate, tileBox, "┬──    ──\n│", borkmark.Yellow)
}

type pos int

const (
	aisle1 pos = iota
	left3
	left2
	left1
	aisle2
	right1
	right2
	right3
)

func (p pos) String() string {
	switch p {
	case aisle1:
		return "A1"
	case aisle2:
		return "A2"
	case left1:
		return "L1"
	case left2:
		return "L2"
	case left3:
		return "L3"
	case right1:
		return "R1"
	case right2:
		return "R2"
	case right3:
		return "R3"
	}
	return ""
}

var offsets = [8]image.Point{
	{4, 0}, // aisle1
	{1, 1}, // left3
	{2, 1}, // left2
	{3, 1}, // left1
	{4, 1}, // aisle2
	{5, 1}, // right1
	{6, 1}, // right2
	{7, 1}, // right3
}

type car struct {
	color  color
	motive motive
	tile   image.Point
	pos    pos
	wait   int
}

type motive int

const (
	// entering indicates the car is queued to enter the parking lot
	entering motive = iota
	// parking indicates the car is searching for a parking space
	parking
	// waiting indicates the car is parked and waiting for borkshoppers
	waiting
	// exiting indicates the car is leaving the parking lot
	exiting
)

func (m motive) String() string {
	switch m {
	case entering:
		return "entering"
	case parking:
		return "parking"
	case waiting:
		return "waiting"
	case exiting:
		return "exiting"
	}
	return ""
}

type color int

const (
	blueCar = iota
	redCar
	yellowCar
	numColors
)

// NewLotForBounds creates a parking lot that covers the given rectangle,
// rounding up.
func NewLotForBounds(bounds image.Rectangle) *Lot {
	size := bounds.Size()

	w := (size.X + tileWidth) / tileWidth
	h := (size.Y + tileHeight) / tileHeight

	numCars := w * h * 2 * 10 / 11
	numTicks := numCars
	lot := newLot(w, h, numCars)
	lot.prep()
	for i := 0; i < numTicks; i++ {
		lot.Tick()
	}

	return lot
}

func newLot(w, h, numCars int) *Lot {
	numTiles := w * h
	return &Lot{
		width:  w,
		height: h,
		tiles:  make([]tile, numTiles),
		zero:   make([]tile, numTiles),
		cars:   make([]car, numCars),
	}
}

// Lot captures the states of all cars and the data structures necessary
// to coordinate their right of way for each tick.
type Lot struct {
	width, height int
	tiles         []tile
	zero          []tile
	cars          []car
}

// Bounds returns the maximum rectangle that this lot can draw over.
func (lot *Lot) Bounds() image.Rectangle {
	return image.Rect(0, 0, lot.width*tileWidth, lot.height*tileHeight)
}

func (lot *Lot) prep() {
	car := &lot.cars[0]
	car.color = color(rand.Intn(numColors))
	car.motive = parking
	car.tile = image.ZP
	car.pos = aisle1
	car.wait = 1

	tile := &lot.tiles[0]
	tile.aisle1 = 1
}

// Tick advances the parking lot state by one tick.
func (lot *Lot) Tick() {
	// cars claim where they want to be next
	copy(lot.tiles, lot.zero)

	// cars mark where they already are, unconditionally claiming their current
	// location.
	lot.claimOwnTiles()

	// exiting cars try to exit, parking cars wait for them
	lot.exitingCarsClaimTiles()

	// parking cars try to get a spot or move on
	lot.parkingCarsClaimTiles()

	// entering cars try to find an aisle
	lot.enteringCarsClaimTiles()

	// cars mark where they already are, unconditionally claiming their current
	// location, again, because sigh.
	lot.claimOwnTiles()

	// cars move if they won their claim
	lot.moveCars()
}

func (lot *Lot) claimOwnTiles() {
	for i, car := range lot.cars {
		cid := i + 1
		if car.motive != entering {
			tile := &lot.tiles[car.tile.X+car.tile.Y*lot.width]
			switch car.pos {
			case aisle1:
				tile.aisle1 = cid
			case aisle2:
				tile.aisle2 = cid
			case left1:
				tile.aisle2 = cid
				tile.left = cid
			case right1:
				tile.aisle2 = cid
				tile.right = cid
			case left2, left3:
				tile.left = cid
			case right2, right3:
				tile.right = cid
			}
		}
	}
}

func (lot *Lot) exitingCarsClaimTiles() {
	for i, car := range lot.cars {
		cid := i + 1
		if car.motive == exiting {
			tile := &lot.tiles[car.tile.X+car.tile.Y*lot.width]
			switch car.pos {
			case aisle1, left2, left3, right2, right3:
				if cid > tile.aisle2 {
					tile.aisle2 = cid
				}
			case aisle2:
				if car.tile.Y+1 < lot.height {
					next := &lot.tiles[car.tile.X+(car.tile.Y+1)*lot.width]
					if next.aisle1 < cid {
						next.aisle1 = cid
					}
				}
			}
		}
	}
}

func (lot *Lot) parkingCarsClaimTiles() {
	for i, car := range lot.cars {
		cid := i + 1
		if car.motive == parking {
			tile := &lot.tiles[car.tile.X+car.tile.Y*lot.width]
			switch car.pos {
			case aisle1:
				if tile.aisle2 < cid {
					tile.aisle2 = cid
				}
			case aisle2:
				switch car.motive {
				case parking:
					if tile.left == 0 && tile.right == 0 {
						switch rand.Intn(2) {
						case 0:
							tile.left = cid
						case 1:
							tile.right = cid
						}
					} else if tile.left == 0 {
						tile.left = cid
					} else if tile.right == 0 {
						tile.right = cid
					} else if car.tile.Y+1 < lot.height {
						next := &lot.tiles[car.tile.X+(car.tile.Y+1)*lot.width]
						if next.aisle1 < cid {
							next.aisle1 = cid
						}
					}
				}
			}
		}
	}
}

func (lot *Lot) enteringCarsClaimTiles() {
	for i := range lot.cars {
		car := &lot.cars[i]
		cid := i + 1
		if car.motive == entering {
			car.tile = image.Pt(rand.Intn(lot.width), 0)
			tile := &lot.tiles[car.tile.X+car.tile.Y*lot.width]
			if tile.aisle1 < cid {
				tile.aisle1 = cid
			}
		}
	}
}

func (lot *Lot) moveCars() {
	for i := range lot.cars {
		car := &lot.cars[i]
		cid := i + 1
		tile := &lot.tiles[car.tile.X+car.tile.Y*lot.width]

		if car.motive == entering && tile.aisle1 == cid {
			car.motive = parking
			car.color = color(rand.Intn(numColors))
			car.pos = aisle1
			car.wait = rand.Intn(50) + rand.Intn(50) + 100
			continue
		}

		switch car.pos {
		case aisle1:
			if tile.aisle2 == cid {
				car.pos = aisle2
			}
		case aisle2:
			if tile.left == cid {
				car.pos = left1
			} else if tile.right == cid {
				car.pos = right1
			} else if car.tile.Y+1 < lot.height {
				next := &lot.tiles[car.tile.X+(car.tile.Y+1)*lot.width]
				if next.aisle1 == cid {
					car.tile.Y++
					car.pos = aisle1
				}
			} else {
				car.motive = entering
				car.color = 0
				car.tile = image.ZP
				car.pos = aisle1
			}
		default:
			switch car.motive {
			case parking:
				switch car.pos {
				case left1, left2:
					car.pos--
				case right1, right2:
					car.pos++
				case left3, right3:
					car.motive = waiting
				}
			case waiting:
				car.wait--
				if car.wait == 0 {
					car.motive = exiting
				}
			case exiting:
				switch car.pos {
				case left1, left2, left3:
					if tile.aisle2 == cid {
						car.pos++
					}
				case right1, right2, right3:
					if tile.aisle2 == cid {
						car.pos--
					}
				case aisle1:
					if tile.aisle2 == cid {
						car.pos = aisle2
					}
				}
			}
		}
	}
}

// Draw draws the parking lot onto the given display, at its origin.
func (lot *Lot) Draw(dis *display.Display, box image.Rectangle) {
	// dis.Clear(dis.Bounds()) // FIXME should suffice
	dis.Fill(box, " ", display.Colors[7], display.Colors[0])

	// Compute offset position by centering the lot.
	center := box.Dx() / 2
	width := (lot.width / 2) * tileWidth
	offset := image.Pt(center-5-width, 2)

	y := 0
	for x := 0; x < lot.width*tileWidth; x += tileWidth {
		dis.Draw(tileBox.Add(image.Pt(x, y)).Add(box.Min).Add(offset), firstTileTemplate, image.ZP, draw.Over)
	}

	for y := tileHeight; y < tileHeight*lot.height; y += tileHeight {
		for x := 0; x < lot.width*tileWidth; x += tileWidth {
			dis.Draw(tileBox.Add(image.Pt(x, y)).Add(box.Min).Add(offset), tileTemplate, image.ZP, draw.Over)
		}
	}

	for _, car := range lot.cars {
		if car.motive != entering {
			pt := point.MulRespective(car.tile, tileBox.Size()).Add(offsets[int(car.pos)]).Add(box.Min).Add(offset)
			r := carGlyph(car)
			dis.Set(pt.X, pt.Y, r, display.Colors[7], display.Colors[0])
		}
	}
}

func carGlyph(car car) string {
	if car.motive == entering {
		return "👻"
	}
	switch car.pos {
	case aisle1, aisle2:
		switch car.color {
		case blueCar, redCar:
			return "🚘"
		case yellowCar:
			return "🚖"
		}
	default:
		switch car.color {
		case blueCar:
			return "🚙"
		case redCar:
			return "🚗"
		case yellowCar:
			return "🚕"
		}
	}
	return "🦑"
}
