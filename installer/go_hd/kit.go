package main

// Kit HD de High Speed 4 (VanarTechlab, licencia MIT).
// Convierte en HD las texturas que el propio jugador recogio jugando (mods\texture-packs\SLUS-00826\dumps)
// con Real-ESRGAN (modelo realesrgan-x4plus-anime) y las instala en replacements, aplicando los mismos
// criterios que el pack de Vanar: fuera las semitransparentes (texupload-ST*), las que la IA estropea
// (excluidas.txt o color cambiado) y x4 / x2 segun decisiones.txt (o el detalle, si la textura es nueva).
// No incluye ninguna imagen del juego.

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	GameID      = "SLUS-00826"
	Modelo      = "realesrgan-x4plus-anime"
	UmbralColor = 20.0 // como hd\verificar.py
	UmbralX4    = 3.0  // como hd\preparar_pack.py
	HDPackage   = "psx.enhancement.hd-textures"
	HDFeature   = "hd-textures"
)

type Kit struct {
	Base     string // carpeta del juego
	Log      func(string)
	Progress func(int)
}

func (k *Kit) logf(f string, a ...any) { k.Log(fmt.Sprintf(f, a...)) }
func (k *Kit) prog(p int) {
	if k.Progress != nil {
		k.Progress(p)
	}
}

func (k *Kit) texDir() string    { return filepath.Join(k.Base, "mods", "texture-packs", GameID) }
func (k *Kit) dumps() string     { return filepath.Join(k.texDir(), "dumps") }
func (k *Kit) repl() string      { return filepath.Join(k.texDir(), "replacements") }
func (k *Kit) kitDir() string    { return filepath.Join(k.Base, "hd_kit") }
func (k *Kit) stateFile() string { return filepath.Join(k.Base, "mods", "state.toml") }

func fileExists(p string) bool { st, err := os.Stat(p); return err == nil && !st.IsDir() }

func listPNG(dir string) []string {
	ents, _ := os.ReadDir(dir)
	var r []string
	for _, e := range ents {
		if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".png") {
			r = append(r, e.Name())
		}
	}
	sort.Strings(r)
	return r
}

func readList(p string) map[string]string {
	m := map[string]string{}
	f, err := os.Open(p)
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		parts := strings.Fields(l)
		v := ""
		if len(parts) > 1 {
			v = parts[1]
		}
		m[parts[0]] = v
	}
	return m
}

// Estado del lanzador (mods\state.toml): activar las texturas HD y la recogida.
// Reescribe solo el bloque [[feature]] de hd-textures y anade el [[package]] si falta.
func (k *Kit) SetHD(enabled, dump bool) error {
	f := k.stateFile()
	old, _ := os.ReadFile(f)
	txt := strings.ReplaceAll(string(old), "\r\n", "\n")
	if strings.TrimSpace(txt) == "" {
		txt = "format_version = 2\n"
	}
	// separar en bloques que empiezan por [[...]]
	lines := strings.Split(txt, "\n")
	var blocks [][]string
	cur := []string{}
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "[[") && len(cur) > 0 {
			blocks = append(blocks, cur)
			cur = nil
		}
		cur = append(cur, l)
	}
	blocks = append(blocks, cur)
	hasPkg := false
	featIdx := -1
	for i, b := range blocks {
		joined := strings.Join(b, "\n")
		head := strings.TrimSpace(b[0])
		if head == "[[package]]" && strings.Contains(joined, `id = "`+HDPackage+`"`) {
			hasPkg = true
		}
		if head == "[[feature]]" && strings.Contains(joined, `package_id = "`+HDPackage+`"`) {
			featIdx = i
		}
	}
	feat := []string{"[[feature]]", `package_id = "` + HDPackage + `"`, `id = "` + HDFeature + `"`,
		fmt.Sprintf("enabled = %v", enabled), "[feature.values]", fmt.Sprintf(`dump = "%v"`, dump), `replacements = "true"`, ""}
	if featIdx >= 0 {
		blocks[featIdx] = feat
	} else {
		blocks = append(blocks, feat)
	}
	if !hasPkg {
		// el [[package]] va antes de los [[feature]]
		ins := len(blocks)
		for i, b := range blocks {
			if strings.TrimSpace(b[0]) == "[[feature]]" {
				ins = i
				break
			}
		}
		pkg := []string{"[[package]]", `id = "` + HDPackage + `"`, ""}
		blocks = append(blocks[:ins], append([][]string{pkg}, blocks[ins:]...)...)
	}
	var out []string
	for _, b := range blocks {
		out = append(out, b...)
	}
	res := strings.Join(out, "\n")
	for strings.Contains(res, "\n\n\n") {
		res = strings.ReplaceAll(res, "\n\n\n", "\n\n")
	}
	os.MkdirAll(filepath.Dir(f), 0o755)
	return os.WriteFile(f, []byte(strings.ReplaceAll(res, "\n", "\r\n")), 0o644)
}

// HDState lee si las texturas HD y la recogida estan activas.
func (k *Kit) HDState() (enabled, dump bool) {
	b, err := os.ReadFile(k.stateFile())
	if err != nil {
		return false, false
	}
	txt := strings.ReplaceAll(string(b), "\r\n", "\n")
	i := strings.Index(txt, `package_id = "`+HDPackage+`"`)
	if i < 0 {
		return false, false
	}
	blk := txt[i:]
	if j := strings.Index(blk, "[["); j > 0 {
		blk = blk[:j]
	}
	return strings.Contains(blk, "enabled = true"), strings.Contains(blk, `dump = "true"`)
}

type Counts struct{ Recogidas, Instaladas, Pendientes int }

func (k *Kit) Count() Counts {
	excl := readList(filepath.Join(k.kitDir(), "excluidas.txt"))
	have := map[string]bool{}
	for _, f := range listPNG(k.repl()) {
		have[f] = true
	}
	var c Counts
	c.Instaladas = len(have)
	for _, f := range listPNG(k.dumps()) {
		if strings.HasPrefix(f, "texupload-ST") {
			continue
		}
		c.Recogidas++
		if _, ex := excl[f]; !ex && !have[f] {
			c.Pendientes++
		}
	}
	return c
}

func loadPNG(p string) (*image.NRGBA, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	b := im.Bounds()
	n := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			n.Set(x, y, color.NRGBAModel.Convert(im.At(b.Min.X+x, b.Min.Y+y)))
		}
	}
	return n, nil
}

// boxDown reduce por un factor entero promediando (alfa premultiplicado, como el "RGBa" de PIL)
func boxDown(src *image.NRGBA, s int) *image.NRGBA {
	w, h := src.Rect.Dx()/s, src.Rect.Dy()/s
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b, a, n float64
			for yy := 0; yy < s; yy++ {
				for xx := 0; xx < s; xx++ {
					px, py := x*s+xx, y*s+yy
					if px >= src.Rect.Dx() || py >= src.Rect.Dy() {
						continue
					}
					c := src.NRGBAAt(px, py)
					af := float64(c.A) / 255
					r += float64(c.R) * af
					g += float64(c.G) * af
					b += float64(c.B) * af
					a += float64(c.A)
					n++
				}
			}
			if n == 0 {
				continue
			}
			A := a / n
			col := color.NRGBA{A: uint8(A + 0.5)}
			if A > 0 {
				k := 255 / A / n
				col.R = clamp8(r * k)
				col.G = clamp8(g * k)
				col.B = clamp8(b * k)
			}
			dst.SetNRGBA(x, y, col)
		}
	}
	return dst
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

func gray(r, g, b float64) float64 { return 0.299*r + 0.587*g + 0.114*b }

// colorDiff: diferencia media (gris) entre el original y la HD reducida, solo donde el original es visible
func colorDiff(orig, small *image.NRGBA) float64 {
	w, h := orig.Rect.Dx(), orig.Rect.Dy()
	var sum float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			a := orig.NRGBAAt(x, y)
			if a.A == 0 {
				continue
			}
			b := small.NRGBAAt(x, y)
			m := float64(a.A) / 255
			dr := abs(float64(a.R)*m - float64(b.R)*m)
			dg := abs(float64(a.G)*m - float64(b.G)*m)
			db := abs(float64(a.B)*m - float64(b.B)*m)
			sum += gray(dr, dg, db)
		}
	}
	return sum / float64(w*h)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func savePNG(p string, im image.Image) error {
	tmp := p + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(f, im); err != nil {
		f.Close()
		return err
	}
	f.Close()
	os.Remove(p)
	return os.Rename(tmp, p)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// Crear: escala las texturas recogidas que faltan y las instala.
func (k *Kit) Crear() error {
	esr := filepath.Join(k.kitDir(), "realesrgan", "realesrgan-ncnn-vulkan"+exeSuffix)
	if _, err := os.Stat(esr); err != nil {
		return fmt.Errorf("falta %s (reinstala High Speed 4)", esr)
	}
	excl := readList(filepath.Join(k.kitDir(), "excluidas.txt"))
	dec := readList(filepath.Join(k.kitDir(), "decisiones.txt"))
	os.MkdirAll(k.repl(), 0o755)
	have := map[string]bool{}
	for _, f := range listPNG(k.repl()) {
		have[f] = true
	}
	var pend []string
	nST, nEx := 0, 0
	for _, f := range listPNG(k.dumps()) {
		switch {
		case strings.HasPrefix(f, "texupload-ST"):
			nST++
		case excl[f] != "" || func() bool { _, ok := excl[f]; return ok }():
			nEx++
		case !have[f]:
			pend = append(pend, f)
		}
	}
	k.logf("Texturas recogidas: %d nuevas para convertir (%d ya instaladas, %d semitransparentes y %d de la lista de excluidas se quedan como en el original).",
		len(pend), len(have), nST, nEx)
	if len(pend) == 0 {
		if len(have) == 0 {
			return errors.New("no hay texturas recogidas. Pulsa 1, juega unas carreras y CIERRA el juego (las texturas se guardan al salir)")
		}
		k.logf("No hay texturas nuevas: ya estaba todo convertido.")
		return k.SetHD(true, false)
	}
	work := filepath.Join(k.kitDir(), "trabajo")
	in, outd := filepath.Join(work, "entrada"), filepath.Join(work, "salida")
	os.RemoveAll(work)
	os.MkdirAll(in, 0o755)
	os.MkdirAll(outd, 0o755)
	defer os.RemoveAll(work)
	for i, f := range pend {
		if err := copyFile(filepath.Join(k.dumps(), f), filepath.Join(in, f)); err != nil {
			return err
		}
		if i%500 == 0 {
			k.prog(2 + 3*i/len(pend))
		}
	}
	// El modelo viaja como .bin.modelo (el empaquetado de PSXRecomp no admite archivos .bin): se prepara aqui.
	mdir := filepath.Join(work, "models")
	os.MkdirAll(mdir, 0o755)
	msrc := filepath.Join(filepath.Dir(esr), "models")
	for _, par := range [][2]string{{Modelo + ".param", Modelo + ".param"}, {Modelo + ".bin.modelo", Modelo + ".bin"}, {Modelo + ".bin", Modelo + ".bin"}} {
		if p := filepath.Join(msrc, par[0]); fileExists(p) {
			if err := copyFile(p, filepath.Join(mdir, par[1])); err != nil {
				return err
			}
		}
	}
	if !fileExists(filepath.Join(mdir, Modelo+".bin")) || !fileExists(filepath.Join(mdir, Modelo+".param")) {
		return fmt.Errorf("falta el modelo %s en %s (reinstala High Speed 4)", Modelo, msrc)
	}
	k.logf("Escalando con Real-ESRGAN (usa la tarjeta grafica; puede tardar varios minutos)...")
	t0 := time.Now()
	cmd := exec.Command(esr, "-i", in, "-o", outd, "-m", mdir, "-n", Modelo, "-s", "4", "-f", "png")
	cmd.Dir = filepath.Dir(esr)
	hideWindow(cmd)
	var errOut strings.Builder
	cmd.Stderr = &errOut
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("no se pudo ejecutar Real-ESRGAN: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	tick := time.NewTicker(2 * time.Second)
	last := -1
loop:
	for {
		select {
		case err := <-done:
			tick.Stop()
			if err != nil && len(listPNG(outd)) == 0 {
				msg := strings.TrimSpace(errOut.String())
				if len(msg) > 600 {
					msg = msg[len(msg)-600:]
				}
				return fmt.Errorf("Real-ESRGAN fallo (%v). Necesita una tarjeta grafica con Vulkan. Detalle: %s", err, msg)
			}
			break loop
		case <-tick.C:
			n := len(listPNG(outd))
			if n != last && (n-last >= 200 || n == len(pend)) {
				last = n
				k.logf("  %d de %d (%s)", n, len(pend), time.Since(t0).Round(time.Second))
			}
			k.prog(5 + 75*n/len(pend))
		}
	}
	k.logf("Escalado en %s. Revisando e instalando...", time.Since(t0).Round(time.Second))
	var nX4, nX2, nMal, nErr int64
	var idx int64 = -1
	var wg sync.WaitGroup
	var mu sync.Mutex
	workers := runtime.NumCPU()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := atomic.AddInt64(&idx, 1)
				if int(i) >= len(pend) {
					return
				}
				f := pend[i]
				orig, err1 := loadPNG(filepath.Join(k.dumps(), f))
				hd, err2 := loadPNG(filepath.Join(outd, f))
				if err1 != nil || err2 != nil {
					atomic.AddInt64(&nErr, 1)
					continue
				}
				ow, oh := orig.Rect.Dx(), orig.Rect.Dy()
				if hd.Rect.Dx() != 4*ow || hd.Rect.Dy() != 4*oh {
					atomic.AddInt64(&nErr, 1)
					continue
				}
				if colorDiff(orig, boxDown(hd, 4)) > UmbralColor {
					atomic.AddInt64(&nMal, 1)
					continue
				}
				d := dec[f]
				if d == "" {
					if detallePIL(hd) >= UmbralX4 {
						d = "x4"
					} else {
						d = "x2"
					}
				}
				dst := filepath.Join(k.repl(), f)
				var err error
				if d == "x4" {
					err = copyFile(filepath.Join(outd, f), dst)
					atomic.AddInt64(&nX4, 1)
				} else {
					small := mitadPIL(hd)
					err = savePNG(dst, small)
					atomic.AddInt64(&nX2, 1)
				}
				if err != nil {
					atomic.AddInt64(&nErr, 1)
				}
				if i%500 == 0 {
					mu.Lock()
					k.prog(80 + int(18*i)/len(pend))
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	k.logf("Instaladas %d texturas HD (%d a x4, %d a x2). %d se quedan como el original (la IA cambiaba el color). %d con error.",
		nX4+nX2, nX4, nX2, nMal, nErr)
	if err := k.SetHD(true, false); err != nil {
		return err
	}
	k.logf("Texturas HD ACTIVADAS (y la recogida apagada). Abre el juego y ya se veran.")
	k.prog(100)
	return nil
}

// InstalarPack instala un pack de texturas HD ya hecho (zip descargado aparte): copia todos los .png
// del zip (esten en la carpeta que esten) a replacements y activa las texturas HD.
func (k *Kit) InstalarPack(zipPath string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("no se pudo abrir %s: %v", zipPath, err)
	}
	defer zr.Close()
	var pngs []*zip.File
	for _, f := range zr.File {
		if !f.FileInfo().IsDir() && strings.EqualFold(filepath.Ext(f.Name), ".png") {
			pngs = append(pngs, f)
		}
	}
	if len(pngs) == 0 {
		return errors.New("ese zip no trae texturas (.png). Elige el zip del pack de texturas HD de High Speed 4")
	}
	k.logf("Instalando %d texturas HD desde %s ...", len(pngs), filepath.Base(zipPath))
	if err := os.MkdirAll(k.repl(), 0o755); err != nil {
		return err
	}
	for i, f := range pngs {
		name := filepath.Base(filepath.FromSlash(f.Name))
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(filepath.Join(k.repl(), name))
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
		if i%1000 == 0 {
			k.prog(100 * i / len(pngs))
			if i > 0 {
				k.logf("  %d de %d", i, len(pngs))
			}
		}
	}
	if err := k.SetHD(true, false); err != nil {
		return err
	}
	k.logf("Instaladas %d texturas HD. Texturas HD ACTIVADAS: abre el juego y ya se veran.", len(pngs))
	k.prog(100)
	return nil
}
