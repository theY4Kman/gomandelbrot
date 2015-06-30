package gomandelbrot

import (
	"image"
	"image/color"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type tile struct {
	x1, x2, y1, y2 int
}

// return a imange containing a rendered mandelbrot set
func Mandelbrot(w, h, i int, z float32, seed int64) *image.RGBA {
	// if seed is set to 0 use current time to seed PRNG
	if seed == 0 {
		rand.Seed(time.Now().UTC().UnixNano())
	}

	// make a butt load of colors!
	colors := make([]color.RGBA, i)
	for index := range colors {
		colors[index] = randomColor()
	}

	var wg = new(sync.WaitGroup)
	work := make(chan tile)

	zoom := 1 / z

	m := image.NewRGBA(image.Rect(0, 0, w, h))

	for t := 0; t < runtime.NumCPU(); t++ {
		go func() {
			for tile := range work {
				for x := tile.x1; x < tile.x2; x++ {
					for y := tile.y1; y < tile.y2; y++ {
						setColor(m, colors, x, y, i, zoom)
					}
				}
				wg.Done()
			}
		}()
	}

	tx := 1
	ty := h

	wg.Add(tx * ty)

	go func() {
		for x := 0; x < w; x += w / tx {
			for y := 0; y < h; y += h / ty {
				work <- tile{x, x + w/tx, y, y + h/ty}
			}
		}

		close(work)
	}()

	wg.Wait()
	return m

}

func randomColor() color.RGBA {

	r := uint8(rand.Float32() * 255)
	g := uint8(rand.Float32() * 255)
	b := uint8(rand.Float32() * 255)

	return color.RGBA{r, g, b, 255}

}

func setColor(m *image.RGBA, colors []color.RGBA, px, py, maxi int, zoom float32) {

	x0 := zoom * (3.5*float32(px)/float32(m.Bounds().Size().X) - 2.5)
	y0 := zoom * (2*float32(py)/float32(m.Bounds().Size().Y) - 1.0)
	x := float32(0)
	y := float32(0)

	i := 0

	for x*x+y*y < 2*2 && i < maxi {
		xtemp := x*x - y*y + x0
		y = 2*x*y + y0
		x = xtemp

		i++
	}

	m.Set(px, py, colors[i-1])
}
