package main

// Remuestreo separable como el de PIL (Image.resize): Lanczos (a=3) y bicubico (a=-0.5),
// con el soporte del filtro escalado al reducir. Sirve para reproducir hd\preparar_pack.py.

import (
	"image"
	"math"
)

type filtro struct {
	f       func(float64) float64
	soporte float64
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1
	}
	x *= math.Pi
	return math.Sin(x) / x
}

var lanczos = filtro{func(x float64) float64 {
	if x < 0 {
		x = -x
	}
	if x < 3 {
		return sinc(x) * sinc(x/3)
	}
	return 0
}, 3}

var bicubic = filtro{func(x float64) float64 {
	const a = -0.5
	if x < 0 {
		x = -x
	}
	if x < 1 {
		return ((a+2)*x-(a+3))*x*x + 1
	}
	if x < 2 {
		return (((x-5)*x+8)*x - 4) * a
	}
	return 0
}, 2}

// pesos de PIL (precompute_coeffs)
func coefs(in, out int, fl filtro) (starts []int, ws [][]float64) {
	scale := float64(in) / float64(out)
	fs := scale
	if fs < 1 {
		fs = 1
	}
	sup := fl.soporte * fs
	for i := 0; i < out; i++ {
		center := (float64(i) + 0.5) * scale
		xmin := int(math.Max(0, math.Floor(center-sup+0.5)))
		xmax := int(math.Min(float64(in), math.Floor(center+sup+0.5)))
		w := make([]float64, xmax-xmin)
		tot := 0.0
		for x := xmin; x < xmax; x++ {
			v := fl.f((float64(x) - center + 0.5) / fs)
			w[x-xmin] = v
			tot += v
		}
		if tot != 0 {
			for j := range w {
				w[j] /= tot
			}
		}
		starts = append(starts, xmin)
		ws = append(ws, w)
	}
	return
}

// imagen en float de 4 canales (r,g,b,a)
type fimg struct {
	w, h int
	p    []float64
}

func toF(im *image.NRGBA, premult bool) *fimg {
	w, h := im.Rect.Dx(), im.Rect.Dy()
	f := &fimg{w, h, make([]float64, w*h*4)}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := im.NRGBAAt(x, y)
			m := 1.0
			if premult {
				m = float64(c.A) / 255
			}
			i := (y*w + x) * 4
			f.p[i], f.p[i+1], f.p[i+2], f.p[i+3] = float64(c.R)*m, float64(c.G)*m, float64(c.B)*m, float64(c.A)
		}
	}
	return f
}

func resizeF(s *fimg, nw, nh int, fl filtro) *fimg {
	// horizontal
	st, ws := coefs(s.w, nw, fl)
	t := &fimg{nw, s.h, make([]float64, nw*s.h*4)}
	for y := 0; y < s.h; y++ {
		for x := 0; x < nw; x++ {
			for c := 0; c < 4; c++ {
				v := 0.0
				for j, w := range ws[x] {
					v += w * s.p[(y*s.w+st[x]+j)*4+c]
				}
				t.p[(y*nw+x)*4+c] = clampF(v)
			}
		}
	}
	st, ws = coefs(s.h, nh, fl)
	r := &fimg{nw, nh, make([]float64, nw*nh*4)}
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			for c := 0; c < 4; c++ {
				v := 0.0
				for j, w := range ws[y] {
					v += w * t.p[((st[y]+j)*nw+x)*4+c]
				}
				r.p[(y*nw+x)*4+c] = clampF(v)
			}
		}
	}
	return r
}

func clampF(v float64) float64 {
	v = math.Round(v)
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// detallePIL reproduce preparar_pack.py: im (RGBa) -> mitad LANCZOS -> RGBA -> RGB -> doble BICUBIC; diferencia gris media
func detallePIL(hd *image.NRGBA) float64 {
	w, h := hd.Rect.Dx(), hd.Rect.Dy()
	hw, hh := w/2, h/2
	if hw < 1 {
		hw = 1
	}
	if hh < 1 {
		hh = 1
	}
	small := resizeF(toF(hd, true), hw, hh, lanczos)
	// RGBa -> RGBA (des-premultiplicar) -> RGB
	for i := 0; i < len(small.p); i += 4 {
		a := small.p[i+3]
		for c := 0; c < 3; c++ {
			if a > 0 {
				small.p[i+c] = clampF(small.p[i+c] * 255 / a)
			} else {
				small.p[i+c] = 0
			}
		}
		small.p[i+3] = 255
	}
	back := resizeF(small, w, h, bicubic)
	var sum float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := hd.NRGBAAt(x, y)
			i := (y*w + x) * 4
			g := 0.299*abs(float64(o.R)-back.p[i]) + 0.587*abs(float64(o.G)-back.p[i+1]) + 0.114*abs(float64(o.B)-back.p[i+2])
			sum += math.Floor(g + 0.5)
		}
	}
	return sum / float64(w*h)
}

// mitadPIL: la version x2 como preparar_pack.py (RGBa LANCZOS a la mitad, alfa a 0/255 con umbral 128)
func mitadPIL(hd *image.NRGBA) *image.NRGBA {
	w, h := hd.Rect.Dx()/2, hd.Rect.Dy()/2
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	s := resizeF(toF(hd, true), w, h, lanczos)
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			a := s.p[i+3]
			var c [3]uint8
			for k := 0; k < 3; k++ {
				if a > 0 {
					c[k] = uint8(clampF(s.p[i+k] * 255 / a))
				}
			}
			al := uint8(0)
			if a >= 128 {
				al = 255
			}
			out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = c[0], c[1], c[2], al
		}
	}
	return out
}
