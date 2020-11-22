package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

func genGammaToLinear() (out [256]uint16) {
	for i := range out {
		out[i] = uint16(math.Floor(math.Pow(float64(i)/255, 2.2)*65535 + .5))
	}

	// I'm fudging this to make gamma corrected -> linear -> gamma corrected
	// a lossless operation.
	// Otherwise both 0 and 1 would map to 0, and I'd be losing information.
	out[1] = 1

	return
}

func genLinearToGamma() (out [65536]uint8) {
	for i := range out {
		out[i] = uint8(math.Floor(math.Pow(float64(i)/65535, 0.4545)*255 + .5))
	}

	// Reverse the fudging I did in gammaToLinear, otherwise 1 would map to 2.
	out[1] = 1

	return
}

var (
	gammaToLinear = genGammaToLinear()
	linearToGamma = genLinearToGamma()
)

// rasterize creates a w by h image by invoking cb in multiple goroutines with x and y corresponding to pixel coordinates.
//
// Due to the way the image filtering, these coordinates may fall outside of the image.
//
// cb will be invoked w*h*s*s times.
func rasterize(w, h, s int, cb func(x, y float64) color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	progress := make(chan struct{})

	work := make(chan int)
	numWorkers := runtime.NumCPU()+1

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			// give each worker its own rng to avoid locking.
			rng := rand.New(rand.NewSource(rand.Int63()))

			for y := range work {
				scanline := img.Pix[y*img.Stride:y*img.Stride+w*4:y*img.Stride+w*4]

				fY := float64(y)

				for x := 0; x < w; x++ {
					fX := float64(x)

					var r, g, b uint64

					for i := 0; i < s*s; i++ {
						sample := cb(
							fX+rng.NormFloat64()*0.44+0.5,
							fY+rng.NormFloat64()*0.44+0.5,
						)

						r += uint64(gammaToLinear[sample.R])
						g += uint64(gammaToLinear[sample.G])
						b += uint64(gammaToLinear[sample.B])
					}

					scanline[x*4+0] = linearToGamma[r/uint64(s*s)]
					scanline[x*4+1] = linearToGamma[g/uint64(s*s)]
					scanline[x*4+2] = linearToGamma[b/uint64(s*s)]
					scanline[x*4+3] = 255
				}

				if showProgress {
					progress <- struct{}{}
				}
			}
		}()
	}

	go func() {
		for y := 0; y < h; y++ {
			work <- y
		}
		close(work)
		wg.Wait()
		close(progress)
	}()

	for i := 0; ; i++ {
		if _, k := <- progress; !k {
			break
		}

		fmt.Printf("\r%d/%d (%d%%)", i, h, 100*i/h)
	}

	if showProgress {
		fmt.Println()
	}

	return img
}
