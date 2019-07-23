` generate a rendering of the Mandelbrot set `

std := load('std')
bmp := load('bmp').bmp

log := std.log
f := std.format
map := std.map
reduce := std.reduce
range := std.range
wf := std.writeFile

` graph position `
CENTERX := ~0.540015
CENTERY := 0.59468

` rendering configurations `
WIDTH := 120
HEIGHT := 120
SCALE := 50000 ` pixels for 1 unit `
MAXITER := 500

` set the correct "escape" threshold for sequence `
ESCAPE := (WIDTH > HEIGHT :: {
	true -> WIDTH / SCALE
	false -> HEIGHT / SCALE
})
ESCAPE := (ESCAPE < 2 :: {
	 true -> 2
	 false -> ESCAPE
})

` complex arithmetic functions `
cpxAbsSq := z => z.0 * z.0 + z.1 * z.1
cpxAdd := (z, w) => [z.0 + w.0, z.1 + w.1]
cpxMul := (z, w) => [
	z.0 * w.0 - z.1 * w.1
	z.0 * w.1 + z.1 * w.0
]

` given a number [0, WIDTH * HEIGHT), return a, (a, b) pair in a + bi
	whose mandelbrot set value is to be computed `
idxToCpx := i => (
	x := (i % WIDTH) - WIDTH / 2
	y := floor(i / WIDTH) - HEIGHT / 2

	[x / SCALE + CENTERX, y / SCALE + CENTERY]
)

` compute divergence speed for a given a + bi,
	returns a number [0, 1)`
lb := 1 / ln(ESCAPE) ` log base`
lhb := ln(0.5) * lb ` log half-base `
mcompute := coord => (
	thresholdSq := ESCAPE * ESCAPE
	(sub := (last, iter) => cpxAbsSq(last) > thresholdSq :: {
		` provably diverges `
		true -> (
			` smoothed rendering from https://csl.name/post/mandelbrot-rendering/ `
			fractional := 5 + iter - lhb - ln(ln(cpxAbsSq(last))) * lb
			fractional / MAXITER
		)
		` maybe converges? `
		false -> iter :: {
			` give up `
			MAXITER -> 1
			` try again `
			_ -> sub(cpxAdd(cpxMul(last, last), coord), iter + 1)
		}
	})([0, 0], 0)
)

` hsl to rgb color converter, for rendering the exterior of the set `
hsl := (h, s, l) => (
	` ported from https://stackoverflow.com/questions/2353211/hsl-to-rgb-color-conversion `
	h2rgb := (p, q, t) => (
		` wrap to [0, 1) `
		t := (t < 1 :: {
			true -> t + 1
			false -> t
		})
		t := (t > 1 :: {
			true -> t - 1
			false -> t
		})

		[t < 1/6, t < 1/2, t < 2/3] :: {
			[true, _, _] -> p + (q - p) * 6 * t
			[_, true, _] -> q
			[_, _, true] -> p + (q - p) * (2/3 - t) * 6
			_ -> p
		}
	)

	q := (l < 0.5 :: {
		true -> l * (1 + s)
		false -> l + s - l * s
	})
	p := 2 * l - q

	[
		255 * h2rgb(p, q, h + 1/3)
		255 * h2rgb(p, q, h)
		255 * h2rgb(p, q, h - 1/3)
	]
)

` map [0, 1) output from mcompute to RGB values `
mrgb := n => n :: {
	1 -> [20, 20, 20]
	_ -> hsl(n, 0.8, 0.6)
}

` generate image `
startTime := time()
total := WIDTH * HEIGHT
file := bmp(WIDTH, HEIGHT, map(range(0, WIDTH * HEIGHT, 1), x => (
	x % floor(total / 50) :: {
		0 -> log(f('{{ progress }}% rendered	{{ rate }} pixels/sec', {
			progress: floor(x / total * 100)
			rate: floor(x / (time() - startTime))
		}))
	}

	mrgb(mcompute(idxToCpx(x)))))
)

` save file `
wf('mandelbrot.bmp', file, result => result :: {
	true -> log('done!')
	() -> log('error writing file!')
})
