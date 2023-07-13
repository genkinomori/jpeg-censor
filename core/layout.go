package core

import (
	"errors"
	"image"
	"image/color"
)

func extendImageHeight(im image.Image, desiredHeight int) *image.RGBA {
	switch im := im.(type) {
	case *image.RGBA:
		currentHeight := im.Bounds().Dy()
		heightToExtend := desiredHeight - currentHeight
		if heightToExtend <= 0 {
			return im
		}
		im.Rect.Max.Y += heightToExtend
		pixLength := cap(im.Pix)
		pixLengthNeeded := im.Stride * desiredHeight
		if pixLength >= pixLengthNeeded {
			im.Pix = im.Pix[:pixLengthNeeded]
		} else {
			newPix := make([]uint8, pixLengthNeeded)
			copy(newPix, im.Pix)
			im.Pix = newPix
		}
		return im

	default:
		b := im.Bounds()
		width := b.Dx()
		newim := image.NewRGBA(image.Rect(0, 0, width, desiredHeight))
		for y := 0; y < desiredHeight; y++ {
			yy := y + b.Min.Y
			if yy >= b.Max.Y {
				break
			}
			for x := 0; x < width; x++ {
				xx := x + b.Min.X
				newim.Set(x, y, im.At(xx, yy))
			}
		}
		return newim
	}
}

func invColor(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return color.RGBA{uint8(255 - b), uint8(255 - g), uint8(255 - r), uint8(a)}
}

// `cr.CensoredImage` would be occupied and should not be used again.
func LayoutCensoredImage(cr *CensorResult) *image.RGBA {
	bounds := cr.CensoredImage.Bounds()
	imHeight := ceildiv(bounds.Dy(), 16) * 16
	imWidth := bounds.Dx()
	blocksInOneRow := imWidth / 16
	rowsNeeded := ceildiv(len(cr.Blocks), blocksInOneRow)
	blocksHeight := rowsNeeded * 16
	header := MakeHeader(cr)
	headerLength := header.Len()
	headerHeight := ceildiv(headerLength, imWidth)
	totalHeight := imHeight + blocksHeight + headerHeight
	im := extendImageHeight(cr.CensoredImage, totalHeight)

	// write blocks
	var (
		bx = bounds.Min.X
		by = bounds.Min.Y
		yy = imHeight
		xx = 0
	)
	for _, b := range cr.Blocks {
		bm := b.Image
		bb := bm.Bounds()
		for y := bb.Min.Y; y < bb.Max.Y; y++ {
			for x := bb.Min.X; x < bb.Max.X; x++ {
				ix := xx + (x - bb.Min.X) + bx
				iy := yy + (y - bb.Min.Y) + by
				im.SetRGBA(ix, iy, invColor(bm.RGBAAt(x, y)))
			}
		}
		xx += 16
		if xx >= imWidth {
			yy, xx = yy+16, 0
		}
	}

	// write header
	var (
		y = totalHeight - 1
		x = imWidth - 1
	)
	for i := 0; i < headerLength; i++ {
		b := header.Get(i)
		if b == 0 {
			im.SetRGBA(x+bx, y+by, color.RGBA{0, 0, 0, 255})
		} else {
			im.SetRGBA(x+bx, y+by, color.RGBA{255, 255, 255, 255})
		}
		if x == 0 {
			y, x = y-1, imWidth-1
		} else {
			x--
		}
	}

	return im
}

var (
	ErrImageIsTooSmall   = errors.New("image is too small")
	ErrHeaderIsIncorrect = errors.New("header is incorrect")
)

func UnlayoutCensoredImage(im image.Image) (*CensorResult, error) {
	bounds := im.Bounds()
	height := bounds.Dy()
	width := bounds.Dx()
	area := width * height
	headerPixels := (len(MAGIC) + 8) * 8
	if area < headerPixels {
		return nil, ErrImageIsTooSmall
	}

	var (
		by = bounds.Min.Y
		bx = bounds.Min.X
		yy = height - 1
		xx = width - 1
	)
	header := NewHeader()
	advance := func() {
		v, _, _, _ := im.At(xx+bx, yy+by).RGBA()
		v >>= 8
		if v < 128 {
			header.PushBit(0)
		} else {
			header.PushBit(1)
		}
		if xx == 0 {
			yy, xx = yy-1, width-1
		} else {
			xx--
		}
	}
	for i := 0; i < headerPixels; i++ {
		advance()
	}
	headerLength, err := ParseMagicAndLength(header)
	if err != nil {
		return nil, err
	}
	headerPixels += int(headerLength)
	if area < headerPixels {
		return nil, ErrImageIsTooSmall
	}
	header = NewHeader()
	for i := 0; i < int(headerLength); i++ {
		advance()
	}
	ph := ParseHeader(header)

	pHeight := int(ph.Height)
	if pHeight > height {
		return nil, ErrHeaderIsIncorrect
	}
	ret := &CensorResult{
		CensoredImage: subImage(im, image.Rect(bx, by, bx+width, by+pHeight)),
	}
	blocks := make([]*Block, 0)
	xx, yy = 0, ceildiv(pHeight, 16)*16
	for i := 0; i < int(ph.Blocks); i++ {
		b := &Block{
			Index: int(ph.BlockIndex[i]),
		}
		bim := image.NewRGBA(image.Rect(0, 0, 16, 16))
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if x+xx >= width {
					break
				}
				bim.Set(x, y, invColor(im.At(x+xx+bx, y+yy+by)))
			}
		}
		b.Image = bim
		blocks = append(blocks, b)
		xx += 16
		if xx >= width {
			xx, yy = 0, yy+16
		}
	}

	ret.Blocks = blocks
	return ret, nil
}
