` a basic image encoder for bitmap files `

std := load('std')

each := std.each

` utility function for splitting a large number > 16^2 into 4-byte list`
hexsplit := n => (
	` accumulate list, growing to right `
	acc := [0, 0, 0, 0]

	` max 4 bytes `
	(sub := (p, i) => p < 256 | i > 3 :: {
		true -> (
			acc.(i) := p
			acc
		)
		false -> (
			acc.(i) := p % 256
			sub(floor(p / 256), i + 1)
		)
	})(n, 0)
)

` core bitmap image encoder function `
bmp := (width, height, pixels) => (
	` file buffer in which we build the image data `
	buf := []

	` last written byte offset `
	last := {off: 0}
	` append byte values to buf `
	appd := list => each(list, b => (
		buf.(last.off) := b
		last.off := last.off + 1
	))

	` bmp requires that we pad out each pixel row to 4-byte chunks `
	padding := ((3 * width) % 4 :: {
		0 -> []
		1 -> [0, 0, 0]
		2 -> [0, 0]
		3 -> [0]
	})

	` write the nth row of pixels to buf `
	wrow := y => (
		offset := width * y
		(sub := x => x :: {
			width -> ()
			_ -> (
				appd(pixels.(offset + x))
				sub(x + 1)
			)
		})(0)
		appd(padding)
	)

	` write the pixel array to buf `
	wpixels := () => (
		(sub := y => y :: {
			height -> ()
			_ -> (
				wrow(y)
				sub(y + 1)
			)
		})(0)
	)

	` - - bmp header: BITMAPINFOHEADER format `

	` bmp format identifier `
	appd([point('B'), point('M')])
	` file size: 54 is the header bytes, plus 3 bytes per px + row-padding bytes `
	appd(hexsplit(54 + (width * 3 + len(padding)) * height))
	` unused 4 bytes in this format `
	appd([0, 0, 0, 0])
	` pixel array data offset: always 54 if following our simple format `
	appd([54, 0, 0, 0])

	` - - DIB header `

	` num bytes in the DIB header from here `
	appd([40, 0, 0, 0])
	` bitmap width in pixels `
	appd(hexsplit(width))
	` bitmap height in pixels, bottom to top`
	appd(hexsplit(height))
	` number of color planes used: 1 `
	appd([1, 0])
	` number of bits per pixel: 24 (8-bit rgb) `
	appd([24, 00])
	` pixel array compression format: none used `
	appd([0, 0, 0, 0])
	` size of raw bitmap data: 16 bits `
	appd([16, 0, 0, 0])
	` horizontal print resolution of the image: 72 dpi = 2835 pixels / meter `
	appd(hexsplit(2835))
	` vertical print resolution of the image: 72 dpi = 2835 pixels / meter `
	appd(hexsplit(2835))
	` number of colors in palette: 0 `
	appd([0, 0, 0, 0])
	` number of "important" colors: 0 `
	appd([0, 0, 0, 0])

	` - - write pixel array `
	wpixels()

	` return image file buffer `
	buf
)
