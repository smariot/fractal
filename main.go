package main

import (
	"image/color"
	"image/png"
	"log"
	"os"
	"time"
)

const (
	// Position and size
	px   = -0.5557506
	py   = -0.55560
	size = 0.000000001
	//px   = -2
	//py   = -1.2
	//size = 2.5

	// Quality
	imgWidth = 1024
	maxIter  = 1000
	samples  = 50

	showProgress = true
)

func main() {
	log.Println("Rendering...")
	start := time.Now()
	scale := size / float64(imgWidth)
	img := rasterize(imgWidth, imgWidth, samples, func(x, y float64) color.RGBA {
		return paint(mandelbrotIter(x*scale+px, y*scale+py, maxIter))
	})
	end := time.Now()

	log.Println("Done rendering in", end.Sub(start))

	log.Println("Encoding image...")
	f, err := os.Create("result.png")
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
	log.Println("Done!")
}

func paint(r float64, n int) color.RGBA {
	var insideSet = color.RGBA{ R: 255, G: 255, B: 255, A: 255 }

	if r > 4 {
		c := hslToRGB(float64(n) / 800 * r, 1, 0.5)
		return c
	} else {
		return insideSet
	}
}

func mandelbrotIter(px, py float64, maxIter int) (float64, int) {
	var x, y, xx, yy, xy float64

	for i := 0; i < maxIter; i++ {
		xx, yy, xy = x * x, y * y, x * y
		if xx + yy > 4 {
			return xx + yy, i
		}
		x = xx - yy + px
		y = 2 * xy + py
	}

	return xx + yy, maxIter
}

// by u/Boraini
//func mandelbrotIterComplex(px, py float64, maxIter int) (float64, int) {
//	var current complex128
//	pxpy := complex(px, py)
//
//	for i := 0; i < maxIter; i++ {
//		magnitude := cmplx.Abs(current)
//		if magnitude > 2 {
//			return magnitude * magnitude, i
//		}
//		current = current * current + pxpy
//	}
//
//	magnitude := cmplx.Abs(current)
//	return magnitude * magnitude, maxIter
//}