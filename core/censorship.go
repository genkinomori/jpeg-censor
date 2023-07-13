package core

import (
	"image"
	"image/color"
)

type Block struct {
	Index int
	Image *image.RGBA
}

type CensorResult struct {
	CensoredImage image.Image
	Blocks        []*Block
}

func ceildiv(a, b int) int {
	return (a + b - 1) / b
}

func averageColor(p *image.RGBA) color.RGBA {
	var r_sum, g_sum, b_sum, count int32
	bounds := p.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := p.RGBAAt(x, y)
			r_sum += int32(c.R)
			g_sum += int32(c.G)
			b_sum += int32(c.B)
			count += 1
		}
	}
	if count == 0 {
		return color.RGBA{}
	} else {
		return color.RGBA{uint8(r_sum / count), uint8(g_sum / count), uint8(b_sum / count), 255}
	}
}

func Censor(original image.Image, mask *image.Gray, gridExtend int) *CensorResult {
	bounds := original.Bounds()
	censored := image.NewRGBA(bounds)
	blocks := make([]*Block, 0)

	k := 1
	for yy := bounds.Min.Y; yy < bounds.Max.Y; yy += 16 {
		var (
			lastColor      color.RGBA
			lastColorCount int = 0
		)
		for xx := bounds.Min.X; xx < bounds.Max.X; xx += 16 {
			my := yy + 16
			if my > bounds.Max.Y {
				my = bounds.Max.Y
			}
			mx := xx + 16
			if mx > bounds.Max.X {
				mx = bounds.Max.X
			}
			to_mask := false
			for y := yy; !to_mask && y < my; y++ {
				for x := xx; !to_mask && x < mx; x++ {
					if mask.GrayAt(x, y).Y >= 128 {
						to_mask = true
					}
				}
			}
			if to_mask {
				sb := image.NewRGBA(image.Rect(0, 0, mx-xx, my-yy))
				for y := yy; y < my; y++ {
					for x := xx; x < mx; x++ {
						sb.Set(x-xx, y-yy, original.At(x, y))
					}
				}
				block := &Block{
					Index: k,
					Image: sb,
				}
				blocks = append(blocks, block)

				var paintColor color.RGBA
				if lastColorCount > 0 {
					paintColor = lastColor
					lastColorCount--
				} else {
					paintColor = averageColor(sb)
					lastColor = paintColor
					lastColorCount = gridExtend
				}
				// paint
				for y := yy; y < my; y++ {
					for x := xx; x < mx; x++ {
						censored.SetRGBA(x, y, paintColor)
					}
				}
			} else {
				// copy
				for y := yy; y < my; y++ {
					for x := xx; x < mx; x++ {
						censored.Set(x, y, original.At(x, y))
					}
				}
			}
			k++
		}
	}
	return &CensorResult{
		CensoredImage: censored,
		Blocks:        blocks,
	}
}

func subImage(im image.Image, rect image.Rectangle) *image.RGBA {
	switch im := im.(type) {
	case *image.RGBA:
		rect = rect.Intersect(im.Rect)
		if rect.Empty() {
			return &image.RGBA{}
		}
		i := im.PixOffset(rect.Min.X, rect.Min.Y)
		return &image.RGBA{
			Pix:    im.Pix[i:],
			Stride: im.Stride,
			Rect:   rect,
		}

	default:
		newim := image.NewRGBA(rect)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				newim.Set(x, y, im.At(x, y))
			}
		}
		return newim
	}
}

// `cr.CensoredImage` would be occupied and should not be used again.
func Uncensor(cr *CensorResult) *image.RGBA {
	c := cr.CensoredImage
	blocks := make(map[int]*Block)
	for _, b := range cr.Blocks {
		blocks[b.Index] = b
	}
	bounds := c.Bounds()
	var im *image.RGBA
	switch c := c.(type) {
	case *image.RGBA:
		im = c
	default:
		im = subImage(c, bounds)
	}

	k := 1
	for yy := bounds.Min.Y; yy < bounds.Max.Y; yy += 16 {
		for xx := bounds.Min.X; xx < bounds.Max.X; xx += 16 {
			my := yy + 16
			if my > bounds.Max.Y {
				my = bounds.Max.Y
			}
			mx := xx + 16
			if mx > bounds.Max.X {
				mx = bounds.Max.X
			}
			block, haveBlock := blocks[k]
			if haveBlock {
				for y := yy; y < my; y++ {
					for x := xx; x < mx; x++ {
						im.Set(x, y, block.Image.At(x-xx, y-yy))
					}
				}
			}
			k++
		}
	}
	return im
}
