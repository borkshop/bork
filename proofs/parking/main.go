package main

import (
	"fmt"
	"image"
	"image/draw"
	"math/rand"
	"os"
	"time"

	"github.com/borkshop/bork/internal/borkmark"
	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/text"
	"github.com/borkshop/bork/internal/point"
)

// Behaviors:
// * cars enter, park, wait, and exit the parking lot
// * cars move forward when there is space in front of them
// * cars enter available parking spaces
// * cars wait a random amount of time for their passengers to shop
// * cars wait for the space behind them to be safe to enter
// * cars drive down the aisle
// * cars project their right of way onto tiles in their projected path
// * cars yield when they lose the right of way in their path
// Components:
//   Color.
//   Parking:
//   * yield to exiting cars
//   Waiting:
//   * time remaining
//   Exiting:
//   * yield to exiting cars
// * time waiting
// * cars are either entering or exiting the parking lot
// * direction a car is facing
// * direction a car is moving
// * car velocity (turns between moves)

// Parking lot tile:
//
// +-- A. --
// |LllarrR.

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

var tileTemplate *display.Display

func init() {
	tileTemplate = display.New(tileBox)
	text.Write(tileTemplate, tileBox, "┼──    ──\n│", borkmark.Yellow)
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
	blue = iota
	red
	yellow
	numColors
)

func newLot(w, h, numCars int) *lot {
	numTiles := w * h
	return &lot{
		width:  w,
		height: h,
		tiles:  make([]tile, numTiles),
		zero:   make([]tile, numTiles),
		cars:   make([]car, numCars),
	}
}

type lot struct {
	width, height int
	tiles         []tile
	zero          []tile
	cars          []car
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() error {
	rand.Seed(time.Now().UnixNano())

	lot := newLot(10, 10, 100)
	prep(lot)

	// prepare to draw
	dis := display.New(image.Rect(0, 0, lot.width*tileWidth, lot.height*tileHeight))

	var buf []byte
	cur := display.Reset
	buf, cur = cur.Clear(buf)

	for n := 0; n < 100; n++ {

		render(dis, lot, time.Now())
		buf, cur = display.Render(buf, cur, dis, display.Model8)
		// buf = append(buf, "\r\n"...)
		// for i, car := range lot.cars {
		// 	buf, cur = cur.ClearLine(buf)
		// 	buf = append(buf, fmt.Sprintf("%d: %s %s %s  %s %d\r\n", i, car.tile, car.pos.String(), carGlyph(car), car.motive.String(), car.wait)...)
		// 	if i > 10 {
		// 		break
		// 	}
		// }

		tick(lot)

		_, err := os.Stdout.Write(buf)
		if err != nil {
			return err
		}
		buf, cur = cur.Home(buf)

		time.Sleep(200 * time.Millisecond)
	}

	_ = lot
	return nil
}

func prep(lot *lot) {
	car := &lot.cars[0]
	car.color = color(rand.Intn(numColors))
	car.motive = parking
	car.tile = image.ZP
	car.pos = aisle1
	car.wait = 1

	tile := &lot.tiles[0]
	tile.aisle1 = 1
}

func tick(lot *lot) {

	// cars claim where they want to be next
	copy(lot.tiles, lot.zero)

	// cars mark where they already are, unconditionally claiming their current
	// location.
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

	// exiting cars try to exit, parking cars wait for them
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

	// parking cars try to get a spot or move on
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

	// entering cars try to find an aisle
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

	// cars mark where they already are, unconditionally claiming their current
	// location, again, because sigh.
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

	// cars move if they won their claim
	for i := range lot.cars {
		car := &lot.cars[i]
		cid := i + 1
		tile := &lot.tiles[car.tile.X+car.tile.Y*lot.width]

		if car.motive == entering && tile.aisle1 == cid {
			car.motive = parking
			car.color = color(rand.Intn(numColors))
			car.pos = aisle1
			car.wait = rand.Intn(20) + rand.Intn(20) + 10
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

	// cars enter on top
}

func render(dis *display.Display, lot *lot, now time.Time) {
	// dis.Clear(dis.Bounds()) // FIXME should suffice
	dis.Fill(dis.Bounds(), " ", display.Colors[7], display.Colors[0])

	for x := 0; x < lot.width*tileWidth; x += tileWidth {
		for y := 0; y < tileHeight*lot.height; y += tileHeight {
			dis.Draw(tileBox.Add(image.Pt(x, y)), tileTemplate, image.ZP, draw.Over)
		}
	}

	for _, car := range lot.cars {
		if car.motive != entering {
			pt := point.MulRespective(car.tile, tileBox.Size()).Add(offsets[int(car.pos)])
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
		case blue, red:
			return "🚘"
		case yellow:
			return "🚖"
		}
	default:
		switch car.color {
		case blue:
			return "🚙"
		case red:
			return "🚗"
		case yellow:
			return "🚕"
		}
	}
	return ""
}
