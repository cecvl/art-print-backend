package processing

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"net/http"
	"strings"
)

// fetchImage downloads and decodes an image from a URL supporting jpg/png
func fetchImage(url string) (image.Image, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	// Limit reading? assume reasonable size
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		// try decoding as jpeg/png explicitly
		if strings.Contains(err.Error(), "unknown format") {
			// try jpeg
			resp2, err2 := http.Get(url)
			if err2 != nil {
				return nil, "", err2
			}
			defer resp2.Body.Close()
			img, err = jpeg.Decode(resp2.Body)
			if err == nil {
				return img, "jpeg", nil
			}
			// try png
			resp3, err3 := http.Get(url)
			if err3 != nil {
				return nil, "", err3
			}
			defer resp3.Body.Close()
			img, err = png.Decode(resp3.Body)
			if err == nil {
				return img, "png", nil
			}
		}
		return nil, "", err
	}
	return img, format, nil
}

// toGray converts any image to grayscale (NRGBA) and returns a float64 slice of luminance
func toGray(img image.Image) ([]float64, int, int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	gray := make([]float64, w*h)
	// convert using draw to NRGBA then compute luminance
	rgba := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
	idx := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := rgba.NRGBAAt(x, y)
			// luminance approximation
			l := 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
			gray[idx] = l
			idx++
		}
	}
	return gray, w, h
}

// laplacianVariance computes variance of the Laplacian response as a blur metric
func laplacianVariance(gray []float64, w, h int) float64 {
	// 3x3 Laplacian kernel
	k := [3][3]float64{{0, 1, 0}, {1, -4, 1}, {0, 1, 0}}
	var sum, sumSq float64
	count := 0
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			var val float64
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					px := x + kx
					py := y + ky
					val += k[ky+1][kx+1] * gray[py*w+px]
				}
			}
			sum += val
			sumSq += val * val
			count++
		}
	}
	if count == 0 {
		return 0
	}
	mean := sum / float64(count)
	variance := (sumSq / float64(count)) - (mean * mean)
	return variance
}

// ComputeBlurScore downloads image and computes Laplacian variance as blur score
func ComputeBlurScore(url string) (float64, error) {
	img, _, err := fetchImage(url)
	if err != nil {
		return 0, err
	}
	return ComputeBlurFromImage(img), nil
}

// ComputeBlurFromImage computes Laplacian variance-based blur score from an image.Image
func ComputeBlurFromImage(img image.Image) float64 {
	gray, w, h := toGray(img)
	// Downsample for speed if very large
	maxDim := w
	if h > maxDim {
		maxDim = h
	}
	if maxDim > 2000 {
		log.Printf("⚠️ large image %dx%d, consider downsizing before analysis", w, h)
	}
	v := laplacianVariance(gray, w, h)
	// normalize by some factor (empirical) to keep numbers reasonable
	norm := v / 1000.0
	if math.IsNaN(norm) || math.IsInf(norm, 0) {
		norm = 0
	}
	return norm
}

// helper: detect approximate color depth by sampling pixel values; returns bits per channel (8 or 16)
func DetectColorDepth(img image.Image) int {
	// if any channel value exceeds 255 then likely 16-bit; but Go decoders generally normalize to 8-bit.
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	samples := 0
	maxVal := 0
	for y := 0; y < h && samples < 1000; y++ {
		for x := 0; x < w && samples < 1000; x++ {
			r, g, b := colorToInt(img.At(x, y))
			if r > maxVal {
				maxVal = r
			}
			if g > maxVal {
				maxVal = g
			}
			if b > maxVal {
				maxVal = b
			}
			samples++
		}
	}
	if maxVal > 255 {
		return 16
	}
	return 8
}

func colorToInt(c color.Color) (int, int, int) {
	r, g, b, _ := c.RGBA()
	// r/g/b are in 0..65535
	return int(r), int(g), int(b)
}
