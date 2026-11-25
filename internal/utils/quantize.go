package utils

import (
	"image"
	"image/color"
	"image/draw"
)

// pngColors memetakan quality (1 - 99) ke jumlah warna PNG.
// 10 -> 16, 40 -> 64, 70 -> 128, 90 ke atas -> 256.
func pngColors(quality int) int {
	points := []struct{ quality, colors int }{{10, 16}, {40, 64}, {70, 128}, {90, 256}}

	if quality <= points[0].quality {
		return points[0].colors
	}

	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		if quality <= b.quality {
			return a.colors + (quality-a.quality)*(b.colors-a.colors)/(b.quality-a.quality)
		}
	}

	return 256
}

// quantizePNG mengurangi jumlah warna gambar menjadi maksimal maxColors (<= 256).
func quantizePNG(src image.Image, maxColors int) *image.Paletted {
	img, ok := src.(*image.NRGBA)
	if !ok {
		img = image.NewNRGBA(src.Bounds())
		draw.Draw(img, img.Bounds(), src, src.Bounds().Min, draw.Src)
	}

	// Warnanya memang sudah sedikit (logo, ikon): tidak ada warna yang diubah.
	if out := exactPalette(img, maxColors); out != nil {
		return out
	}

	palette := buildPalette(img, maxColors)
	return dither(img, palette)
}

func exactPalette(img *image.NRGBA, maxColors int) *image.Paletted {
	out := image.NewPaletted(img.Bounds(), nil)
	index := make(map[color.NRGBA]uint8, maxColors)
	b := img.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.NRGBAAt(x, y)
			if c.A == 0 {
				c = color.NRGBA{}
			}

			i, ok := index[c]
			if !ok {
				if len(out.Palette) == maxColors {
					return nil
				}
				i = uint8(len(out.Palette))
				index[c] = i
				out.Palette = append(out.Palette, c)
			}
			out.Pix[out.PixOffset(x, y)] = i
		}
	}

	return out
}

// Warna dikelompokkan ke sel: 5 bit per channel RGB dan 4 bit alpha.
// Pixel solid (alpha 255) punya kelompok alpha sendiri supaya tidak
// tercampur dengan pixel semi transparan.
const (
	opaqueGroup = 16
	totalCells  = (opaqueGroup + 1) << 15

	// Alpha tidak ikut di-dither, jadi selisihnya dibuat lebih berat
	// supaya gradasi transparansi tidak terlihat bergaris.
	alphaWeight = 16
)

func cellIndex(r, g, b, a uint8) int {
	group := int(a >> 4)
	if a == 255 {
		group = opaqueGroup
	}
	return group<<15 | int(r>>3)<<10 | int(g>>3)<<5 | int(b>>3)
}

type colorCell struct {
	mean  [4]float64 // R, G, B, A
	count float64
}

type colorBox struct {
	cells   []colorCell
	mean    [4]float64
	channel int     // channel dengan penyebaran terbesar
	spread  float64 // total kuadrat selisih pada channel tersebut
}

func newColorBox(cells []colorCell) colorBox {
	box := colorBox{cells: cells}
	var total float64

	for _, c := range cells {
		total += c.count
		for ch := range c.mean {
			box.mean[ch] += c.mean[ch] * c.count
		}
	}
	for ch := range box.mean {
		box.mean[ch] /= total
	}

	var spread [4]float64
	for _, c := range cells {
		for ch := range c.mean {
			d := c.mean[ch] - box.mean[ch]
			spread[ch] += d * d * c.count
		}
	}
	spread[3] *= alphaWeight
	for ch, s := range spread {
		if s > box.spread {
			box.channel, box.spread = ch, s
		}
	}

	return box
}

// split membagi box di nilai rata-rata channel dengan penyebaran terbesar.
func (box colorBox) split() (low, high []colorCell) {
	for _, c := range box.cells {
		if c.mean[box.channel] < box.mean[box.channel] {
			low = append(low, c)
		} else {
			high = append(high, c)
		}
	}
	return low, high
}

// buildPalette berulang kali membagi kelompok warna yang paling menyebar
// sampai jumlahnya mencapai maxColors.
func buildPalette(img *image.NRGBA, maxColors int) color.Palette {
	var (
		sums        = make([]colorCell, totalCells)
		transparent bool
		palette     color.Palette
	)

	for i := 0; i < len(img.Pix); i += 4 {
		p := img.Pix[i : i+4 : i+4]
		if p[3] == 0 {
			transparent = true
			continue
		}
		cell := &sums[cellIndex(p[0], p[1], p[2], p[3])]
		cell.count++
		for ch := range cell.mean {
			cell.mean[ch] += float64(p[ch])
		}
	}

	if transparent {
		palette = append(palette, color.NRGBA{})
	}

	var opaque, translucent []colorCell
	for i, cell := range sums {
		if cell.count == 0 {
			continue
		}
		for ch := range cell.mean {
			cell.mean[ch] /= cell.count
		}
		if i>>15 == opaqueGroup {
			opaque = append(opaque, cell)
		} else {
			translucent = append(translucent, cell)
		}
	}

	var boxes []colorBox
	for _, cells := range [][]colorCell{opaque, translucent} {
		if len(cells) > 0 {
			boxes = append(boxes, newColorBox(cells))
		}
	}

	for len(palette)+len(boxes) < maxColors {
		widest := 0
		for i := range boxes {
			if boxes[i].spread > boxes[widest].spread {
				widest = i
			}
		}
		if boxes[widest].spread == 0 {
			break
		}

		low, high := boxes[widest].split()
		if len(low) == 0 || len(high) == 0 {
			boxes[widest].spread = 0 // hanya selisih pembulatan
			continue
		}
		boxes[widest] = newColorBox(low)
		boxes = append(boxes, newColorBox(high))
	}

	for _, box := range boxes {
		palette = append(palette, color.NRGBA{
			R: uint8(box.mean[0] + 0.5),
			G: uint8(box.mean[1] + 0.5),
			B: uint8(box.mean[2] + 0.5),
			A: uint8(box.mean[3] + 0.5),
		})
	}

	return palette
}

// dither menerapkan palette dengan Floyd-Steinberg supaya gradasi tetap halus.
// Pencarian warna terdekat di-cache per sel agar gambar besar tetap cepat.
func dither(img *image.NRGBA, palette color.Palette) *image.Paletted {
	colors := make([]color.NRGBA, len(palette))
	transparentIndex := uint8(0)
	for i, c := range palette {
		colors[i] = c.(color.NRGBA)
		if colors[i].A == 0 {
			transparentIndex = uint8(i)
		}
	}

	cache := make([]int16, totalCells)
	for i := range cache {
		cache[i] = -1
	}

	nearest := func(r, g, b, a uint8) uint8 {
		cell := cellIndex(r, g, b, a)
		if cache[cell] >= 0 {
			return uint8(cache[cell])
		}

		best, bestDistance := 0, int(^uint(0)>>1)
		for i, c := range colors {
			// Solid hanya boleh dipasangkan dengan solid, begitu juga sebaliknya.
			if c.A == 0 || (c.A == 255) != (a == 255) {
				continue
			}
			dr, dg, db, da := int(c.R)-int(r), int(c.G)-int(g), int(c.B)-int(b), int(c.A)-int(a)
			if d := dr*dr + dg*dg + db*db + alphaWeight*da*da; d < bestDistance {
				best, bestDistance = i, d
			}
		}
		cache[cell] = int16(best)
		return uint8(best)
	}

	var (
		bounds = img.Bounds()
		width  = bounds.Dx()
		out    = image.NewPaletted(bounds, palette)
		// Satu kolom tambahan di kiri dan kanan supaya tidak perlu cek batas.
		current = make([][3]int, width+2)
		next    = make([][3]int, width+2)
	)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < width; x++ {
			p := img.Pix[y*img.Stride+x*4:]
			if p[3] == 0 {
				out.Pix[y*out.Stride+x] = transparentIndex
				continue
			}

			var want [3]uint8
			for ch := range want {
				want[ch] = uint8(min(255, max(0, int(p[ch])+current[x+1][ch]/16)))
			}

			i := nearest(want[0], want[1], want[2], p[3])
			out.Pix[y*out.Stride+x] = i

			got := [3]uint8{colors[i].R, colors[i].G, colors[i].B}
			for ch := range want {
				diff := int(want[ch]) - int(got[ch])
				current[x+2][ch] += diff * 7
				next[x][ch] += diff * 3
				next[x+1][ch] += diff * 5
				next[x+2][ch] += diff
			}
		}
		current, next = next, current
		clear(next)
	}

	return out
}
