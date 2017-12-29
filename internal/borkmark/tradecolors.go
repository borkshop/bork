package borkmark

import "image/color"

var (
	// Blue is the primary BØRK brand color, used for voids and fills, and
	// regular copy on BØRK White.
	Blue = color.RGBA{2, 50, 145, 255}
	// Yellow is the BØRK logotype color, used for titles, headings, and trim
	// lines, on a field of BØRK Blue.
	Yellow = color.RGBA{213, 179, 42, 255}
	// White is the BØRK showroom floor color, used for copy on BØRK Blue.
	White = color.RGBA{192, 198, 187, 255}

	// Asphalt is a shade of sticky bitumen that tars the BØRK parking lot.
	Asphalt = color.RGBA{29, 33, 48, 255}
	// Smog is the color of the sky over BØRK, an accumulating poisonous vapour
	// herding the masses onto the showroom floor.
	Smog = color.RGBA{20, 185, 255, 255}
)
