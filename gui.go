// Copyright 2021 The img-diff Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/app/headless"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/image/tiff"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

const (
	width  = 800
	height = 800
)

var (
	defaultMargin = unit.Dp(10)
)

type UI struct {
	img1 image.Image
	img2 image.Image
	diff image.Image

	dmin float64
	dmax float64
	size image.Point

	ctx   layout.Context
	theme *material.Theme
}

func NewUI(img1, img2 image.Image) *UI {
	diff, dmin, dmax := imageDiff(img1, img2)

	return &UI{
		img1:  img1,
		img2:  img2,
		diff:  diff,
		dmin:  dmin,
		dmax:  dmax,
		size:  image.Pt(width, height),
		theme: material.NewTheme(gofont.Collection()),
	}
}

func (ui *UI) run() {
	win := app.NewWindow(
		app.Title("img-diff"),
		app.Size(unit.Px(width), unit.Px(height)),
	)
	defer win.Close()

	for e := range win.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(new(op.Ops), e)
			ui.size = e.Size
			ui.Layout(gtx)
			e.Frame(gtx.Ops)
		case key.Event:
			switch e.Name {
			case "Q", key.NameEscape:
				win.Close()

			case "R":
				// TODO: rescale/resize

			case "F11":
				err := ui.screenshot()
				if err != nil {
					log.Fatalf("could not take screenshot: %+v", err)
				}
			}
		case system.DestroyEvent:
			os.Exit(0)
		}
	}
}

func (ui *UI) Layout(gtx C) D {
	widgets := []layout.Widget{
		func(gtx C) D {
			return layout.Center.Layout(
				gtx,
				func(gtx C) D {
					imgs := []image.Image{ui.img1, ui.img2}
					list := &layout.List{Axis: layout.Horizontal}
					return list.Layout(gtx, len(imgs),
						func(gtx C, i int) D {
							img := imgs[i]
							scale := ui.scale(img)
							return widget.Border{
								Color: color.NRGBA{A: 255},
								Width: unit.Dp(2),
							}.Layout(gtx, func(gtx C) D {
								return layout.UniformInset(defaultMargin).Layout(
									gtx,
									Image{
										Src:   paint.NewImageOp(img),
										Scale: scale,
									}.Layout,
								)
							})
						},
					)
				},
			)
		},

		func(gtx C) D {
			label := material.H6(
				ui.theme,
				fmt.Sprintf("Diff:\n - min= %g\n - max= %g", ui.dmin, ui.dmax),
			)
			label.Font.Variant = text.Variant("Mono")
			return layout.Center.Layout(
				gtx,
				label.Layout,
			)
		},

		func(gtx C) D {
			return layout.UniformInset(defaultMargin).Layout(gtx, func(gtx C) D {
				return layout.Center.Layout(gtx, func(gtx C) D {
					return widget.Border{
						Color: color.NRGBA{A: 255},
						Width: unit.Dp(2),
					}.Layout(gtx, func(gtx C) D {
						img := ui.diff
						scale := ui.scale(img)
						return Image{
							Src:   paint.NewImageOp(img),
							Scale: scale,
						}.Layout(gtx)
					})
				})
			})
		},
	}

	list := layout.List{
		Axis: layout.Vertical,
	}
	return list.Layout(gtx, len(widgets), func(gtx C, i int) D {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func (ui *UI) scale(img image.Image) float32 {
	sz := 0.5 * float32(ui.size.X-100)
	dx := float32(img.Bounds().Dx())
	scale := dx / sz
	return 1 / scale
}

func (ui *UI) screenshot() error {
	head, err := headless.NewWindow(ui.size.X, ui.size.Y)
	if err != nil {
		return err
	}

	gtx := layout.Context{
		Ops:         new(op.Ops),
		Constraints: layout.Exact(ui.size),
	}
	ui.Layout(gtx)

	err = head.Frame(gtx.Ops)
	if err != nil {
		return err
	}

	img, err := head.Screenshot()
	if err != nil {
		return err
	}

	f, err := os.Create("out.png")
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return err
	}

	return f.Close()
}

type Image struct {
	Src   paint.ImageOp
	Scale float32
}

func (img Image) Layout(gtx layout.Context) layout.Dimensions {
	scale := img.Scale
	if scale == 0 {
		scale = 160.0 / 72.0
	}
	size := img.Src.Size()
	x := float32(size.X)
	y := float32(size.Y)

	w, h := gtx.Px(unit.Dp(x*scale)), gtx.Px(unit.Dp(y*scale))
	cs := gtx.Constraints
	d := cs.Constrain(image.Pt(w, h))
	stack := op.Push(gtx.Ops)
	clip.Rect(image.Rectangle{Max: d}).Add(gtx.Ops)

	aff := f32.Affine2D{}.Scale(
		f32.Pt(0, 0),
		f32.Pt(scale, scale),
	)
	op.Affine(aff).Add(gtx.Ops)

	img.Src.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Pop()
	return layout.Dimensions{Size: d}
}

func loadImage(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("could not open image file %q: %w", name, err)
	}
	defer f.Close()

	switch ext := strings.ToLower(filepath.Ext(name)); ext {
	case ".png":
		img, err := png.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("could not decode PNG image file %q: %w", name, err)
		}
		return img, nil

	case ".jpeg", ".jpg":
		img, err := jpeg.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("could not decode JPEG image file %q: %w", name, err)
		}
		return img, nil

	case ".gif":
		img, err := gif.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("could not decode GIF image file %q: %w", name, err)
		}
		return img, nil

	case ".tif", ".tiff":
		img, err := tiff.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("could not decode TIFF image file %q: %w", name, err)
		}
		return img, nil

	default:
		return nil, fmt.Errorf("unknown image file extension %q", ext)
	}
}

func imageDiff(v1, v2 image.Image) (image.Image, float64, float64) {
	img1, ok := v1.(*image.RGBA)
	if !ok {
		img1 = newRGBAFrom(v1)
	}

	img2, ok := v2.(*image.RGBA)
	if !ok {
		img2 = newRGBAFrom(v2)
	}

	r1 := img1.Bounds()
	r2 := img2.Bounds()
	diff := image.NewGray16(r1.Union(r2))
	draw.Draw(
		diff, diff.Bounds(),
		&image.Uniform{C: color.RGBA{A: 255}},
		image.Point{}, draw.Src,
	)

	bnd := r1.Intersect(r2)
	dmin := +math.MaxFloat64
	dmax := -math.MaxFloat64
	for x := bnd.Min.X; x < bnd.Max.X; x++ {
		for y := bnd.Min.Y; y < bnd.Max.Y; y++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)
			vd := yiqDiff(c1, c2)
			if vd > 0 {
				dmin = math.Min(vd, dmin)
			}
			dmax = math.Max(vd, dmax)
			diff.SetGray16(x, y, color.Gray16{Y: uint16(vd * math.MaxUint16)})
		}
	}
	return diff, dmin, dmax
}

// yiqDiff returns the normalized difference between the colors of 2 pixels,
// in the NTSC YIQ color space, as described in:
//
//   Measuring perceived color difference using YIQ NTSC
//   transmission color space in mobile applications.
//   Yuriy Kotsarenko, Fernando Ramos.
//
// An electronic version is available at:
//
// - http://www.progmat.uaem.mx:8080/artVol2Num2/Articulo3Vol2Num2.pdf
func yiqDiff(c1, c2 color.RGBA) float64 {
	const max = 35215.0 // difference between 2 maximally different pixels.

	var (
		r1 = float64(c1.R)
		g1 = float64(c1.G)
		b1 = float64(c1.B)

		r2 = float64(c2.R)
		g2 = float64(c2.G)
		b2 = float64(c2.B)

		y1 = r1*0.29889531 + g1*0.58662247 + b1*0.11448223
		i1 = r1*0.59597799 - g1*0.27417610 - b1*0.32180189
		q1 = r1*0.21147017 - g1*0.52261711 + b1*0.31114694

		y2 = r2*0.29889531 + g2*0.58662247 + b2*0.11448223
		i2 = r2*0.59597799 - g2*0.27417610 - b2*0.32180189
		q2 = r2*0.21147017 - g2*0.52261711 + b2*0.31114694

		y = y1 - y2
		i = i1 - i2
		q = q1 - q2

		diff = 0.5053*y*y + 0.299*i*i + 0.1957*q*q
	)
	return diff / max
}

func newRGBAFrom(src image.Image) *image.RGBA {
	var (
		bnds = src.Bounds()
		dst  = image.NewRGBA(bnds)
	)
	draw.Draw(dst, bnds, src, image.Point{}, draw.Src)
	return dst
}
