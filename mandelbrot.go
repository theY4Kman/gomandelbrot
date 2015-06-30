package gomandelbrot

import (
	"image"
	"image/color"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const tile_width = 48
const tile_height = 48

type tile struct {
	x1, x2, y1, y2 int
}

// return a imange containing a rendered mandelbrot set
func Mandelbrot(width, height, color_count int, zoom float32, seed int64) *image.RGBA {
	seedPRNG(seed)
	colors := initColors(color_count)

	var wg = new(sync.WaitGroup)
	work := make(chan tile)

	zoom = 1 / zoom

	mandel_image := image.NewRGBA(image.Rect(0, 0, width, height))

	// spawn off goroutines to work on tiles
	for t := 0; t < runtime.NumCPU(); t++ {
		go workIt(work, wg, mandel_image, colors, zoom)
	}

	wg.Add(height)
	tile_counter := make(chan int)
	go feedIt(work, tile_counter, width, height)

	go func() {
		tc := 0
		for x := range tile_counter {
			tc += x
		}
		log.Printf("Tile count: %d\n", tc)
	}()
	wg.Wait()

	return mandel_image
}

func workIt(work <-chan tile, wg *sync.WaitGroup, mandel_image *image.RGBA,
	colors []color.RGBA, zoom float32) {
	for tile := range work {
		for x := tile.x1; x < tile.x2; x++ {
			for y := tile.y1; y < tile.y2; y++ {
				setColor(mandel_image, colors, x, y, zoom)
			}
		}
		wg.Done()
	}
}

// I can't belieb this isn't something in golang already
func min(a, b int) (ret int) {
	if a < b {
		ret = a
	} else {
		ret = b
	}
	return
}

func feedIt(work chan<- tile, tile_counter chan<- int, width, height int) {
	for x := 0; x < width; x += tile_width {
		for y := 0; y < height; y += tile_height {
			my_tile := tile{x, min(x+tile_width, width-1), y, min(y+tile_height, height-1)}
			work <- my_tile
			log.Printf("creating tile: %v\n", my_tile)
			tile_counter <- 1
		}
	}
	log.Println("closing channels")
	close(work)
	close(tile_counter)
}

// make a butt load of colors!
func initColors(count int) []color.RGBA {
	colors := make([]color.RGBA, count)
	for index := range colors {
		colors[index] = randomColor()
	}
	return colors
}

// seed PRNG to value given, if it is 0 seed to current time
func seedPRNG(seed int64) {
	if seed == 0 {
		rand.Seed(time.Now().UTC().UnixNano())
	} else {
		rand.Seed(seed)
	}
}

// gimmie a random color
func randomColor() color.RGBA {
	r := uint8(rand.Float32() * 255)
	g := uint8(rand.Float32() * 255)
	b := uint8(rand.Float32() * 255)
	return color.RGBA{r, g, b, 255}
}

// here be mandelbrotwerst
func setColor(m *image.RGBA, colors []color.RGBA, px, py int, zoom float32) {
	x0 := zoom * (3.5*float32(px)/float32(m.Bounds().Size().X) - 2.5)
	y0 := zoom * (2*float32(py)/float32(m.Bounds().Size().Y) - 1.0)
	x := float32(0)
	y := float32(0)

	i := 0

	for x*x+y*y < 2*2 && i < len(colors) {
		xtemp := x*x - y*y + x0
		y = 2*x*y + y0
		x = xtemp
		i++
	}

	m.Set(px, py, colors[i-1])
}
