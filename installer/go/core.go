package main

// Instalacion de High Speed 4 (VanarTechlab, licencia MIT).
// Copia el juego, deja elegido el disco del jugador, descarga el compilador clang (opcional) y
// prepara el juego: traduce y compila el codigo DESDE EL DISCO DEL JUGADOR (hs4_aot\hs4_preparar.py).
// No incluye ningun archivo de Electronic Arts ni BIOS de Sony.

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	Version     = "0.8"
	DiscSHA256  = "f0436c8dc7cf13d73974d75c2cc20bc35543726a5ca5bf51f88f961f6a28fe5d"
	ClangZipURL = "https://github.com/RetroPortingToolKit/RetroPorting-Toolchains/releases/latest/download/cmake-clang-v1-windows-x64.zip"
	GameExe     = "HighSpeed4.exe"
)

type Options struct {
	Base     string // carpeta del instalador (con juego\ y herramientas\)
	Cue      string // .cue del disco del jugador
	Dest     string // carpeta donde se instala el juego
	Preparar bool   // descargar clang y preparar el juego ahora
	Accesos  bool   // crear accesos directos (escritorio y menu Inicio)
	Log      func(string)
	Progress func(int) // 0..100
}

func (o *Options) logf(f string, a ...any) { o.Log(fmt.Sprintf(f, a...)) }
func (o *Options) prog(p int) {
	if o.Progress != nil {
		o.Progress(p)
	}
}

func fileExists(p string) bool { st, err := os.Stat(p); return err == nil && !st.IsDir() }
func dirExists(p string) bool  { st, err := os.Stat(p); return err == nil && st.IsDir() }

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
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

// copyTree copia src en dst. No toca lo que el jugador ya tiene y el juego crea solo
// (ajustes, partidas, codigo compilado), salvo que venga en la descarga.
func copyTree(src, dst string, every func(n int)) (int, error) {
	n := 0
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		t := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(t, 0o755)
		}
		n++
		if every != nil && n%50 == 0 {
			every(n)
		}
		return copyFile(p, t)
	})
	return n, err
}

// binOfCue devuelve el .bin de la primera pista (FILE "..." BINARY) del .cue
func binOfCue(cue string) (string, error) {
	f, err := os.Open(cue)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if len(l) > 5 && strings.EqualFold(l[:5], "FILE ") {
			rest := strings.TrimSpace(l[5:])
			name := rest
			if strings.HasPrefix(rest, "\"") {
				if i := strings.Index(rest[1:], "\""); i >= 0 {
					name = rest[1 : 1+i]
				}
			} else if i := strings.LastIndex(rest, " "); i > 0 {
				name = rest[:i]
			}
			if !filepath.IsAbs(name) {
				name = filepath.Join(filepath.Dir(cue), name)
			}
			return name, nil
		}
	}
	return "", errors.New("el .cue no tiene ninguna linea FILE")
}

func sha256File(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func tomlPath(s string) string { return strings.ReplaceAll(s, `\`, "/") }

// writeDiscSetting escribe [disc] path en settings.toml (lo lee el lanzador); conserva el resto.
func writeDiscSetting(dest, cue string) error {
	f := filepath.Join(dest, "settings.toml")
	line := `path = "` + tomlPath(cue) + `"`
	old, err := os.ReadFile(f)
	if err != nil {
		return os.WriteFile(f, []byte("# psxrecomp user settings - escrito por el instalador de High Speed 4.\r\n[disc]\r\n"+line+"\r\n"), 0o644)
	}
	lines := strings.Split(strings.ReplaceAll(string(old), "\r\n", "\n"), "\n")
	inDisc, done := false, false
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "[") {
			inDisc = strings.EqualFold(t, "[disc]")
		}
		if inDisc && !done && strings.HasPrefix(t, "path") {
			lines[i] = line
			done = true
		}
	}
	txt := strings.Join(lines, "\r\n")
	if !done {
		txt = strings.TrimRight(txt, "\r\n") + "\r\n\r\n[disc]\r\n" + line + "\r\n"
	}
	return os.WriteFile(f, []byte(txt), 0o644)
}

func clangCacheDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base, _ = os.UserCacheDir()
	}
	return filepath.Join(base, "retcomm", "toolchains", "cmake-clang-v1", "latest")
}

func findClang(dir string) string {
	found := ""
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && found == "" && !info.IsDir() && strings.EqualFold(info.Name(), "clang"+exeSuffix) {
			found = p
		}
		return nil
	})
	return found
}

type progReader struct {
	r      io.Reader
	n, tot int64
	cb     func(n, tot int64)
	last   time.Time
}

func (p *progReader) Read(b []byte) (int, error) {
	k, err := p.r.Read(b)
	p.n += int64(k)
	if time.Since(p.last) > 500*time.Millisecond || err != nil {
		p.last = time.Now()
		p.cb(p.n, p.tot)
	}
	return k, err
}

// ensureClang devuelve la ruta de clang.exe: el ya instalado o el que descarga.
func ensureClang(o *Options) (string, error) {
	dir := clangCacheDir()
	if c := findClang(dir); c != "" {
		o.logf("Compilador clang ya instalado: %s", c)
		return c, nil
	}
	o.logf("Descargando el compilador clang (libre, de RetroPortingToolKit)...")
	resp, err := http.Get(ClangZipURL)
	if err != nil {
		return "", fmt.Errorf("no se pudo descargar clang: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("no se pudo descargar clang (HTTP %d)", resp.StatusCode)
	}
	tmp, err := os.CreateTemp("", "hs4_clang_*.zip")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	lastMB := int64(-1)
	pr := &progReader{r: resp.Body, tot: resp.ContentLength, cb: func(n, tot int64) {
		if n/(25<<20) != lastMB {
			lastMB = n / (25 << 20)
			if tot > 0 {
				o.logf("  %d de %d MB", n>>20, tot>>20)
				o.prog(25 + int(25*n/tot))
			} else {
				o.logf("  %d MB", n>>20)
			}
		}
	}}
	if _, err := io.Copy(tmp, pr); err != nil {
		tmp.Close()
		return "", fmt.Errorf("descarga de clang cortada: %v", err)
	}
	tmp.Close()
	o.logf("Descomprimiendo clang...")
	zr, err := zip.OpenReader(tmp.Name())
	if err != nil {
		return "", err
	}
	defer zr.Close()
	os.RemoveAll(dir)
	for _, f := range zr.File {
		t := filepath.Join(dir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(t, filepath.Clean(dir)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(t, 0o755)
			continue
		}
		os.MkdirAll(filepath.Dir(t), 0o755)
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out, err := os.Create(t)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return "", err
		}
	}
	c := findClang(dir)
	if c == "" {
		return "", errors.New("el paquete de clang descargado no trae clang.exe")
	}
	return c, nil
}

// prepare ejecuta hs4_aot\hs4_preparar.py con el python que trae el juego y pasa su salida al registro.
func prepare(o *Options, clang string) error {
	py := filepath.Join(o.Dest, "overlay_toolchain", "python", "python"+exeSuffix)
	script := filepath.Join(o.Dest, "hs4_aot", "hs4_preparar.py")
	c := exec.Command(py, "-u", script, "--disc", o.Cue, "--clang", clang)
	c.Dir = o.Dest
	hideWindow(c)
	out, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	c.Stderr = c.Stdout
	if err := c.Start(); err != nil {
		return fmt.Errorf("no se pudo ejecutar %s: %v", py, err)
	}
	sc := bufio.NewScanner(out)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	p := 55
	for sc.Scan() {
		l := sc.Text()
		if strings.HasPrefix(l, "{") { // recibo JSON de la herramienta: largo y poco util para el jugador
			continue
		}
		o.Log("  " + l)
		if strings.Contains(l, "Verified") {
			p = 65
		} else if strings.Contains(l, "Compiled") && p < 90 {
			p += 12
		}
		o.prog(p)
	}
	if err := c.Wait(); err != nil {
		return fmt.Errorf("la preparacion termino con error (%v); el juego funcionara igual, compilando poco a poco mientras juegas", err)
	}
	return nil
}

func shortcuts(o *Options) error {
	exe := filepath.Join(o.Dest, GameExe)
	ps := `$s=New-Object -ComObject WScript.Shell;` +
		`foreach($d in @([Environment]::GetFolderPath('Desktop'),[Environment]::GetFolderPath('Programs'))){` +
		`$l=$s.CreateShortcut((Join-Path $d 'High Speed 4.lnk'));$l.TargetPath='` + exe + `';$l.WorkingDirectory='` + o.Dest + `';$l.Save()}`
	c := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	hideWindow(c)
	return c.Run()
}

func Run(o *Options) error {
	src := filepath.Join(o.Base, "juego")
	if !fileExists(filepath.Join(src, GameExe)) {
		return fmt.Errorf("no encuentro %s en %s: descomprime el zip entero y abre el instalador desde ahi", GameExe, src)
	}
	o.prog(1)
	// 1. Disco
	o.logf("1/5 Comprobando tu disco: %s", o.Cue)
	bin, err := binOfCue(o.Cue)
	if err != nil {
		return err
	}
	if !fileExists(bin) {
		return fmt.Errorf("el .cue apunta a %s, pero ese archivo no existe", bin)
	}
	sum, err := sha256File(bin)
	if err != nil {
		return err
	}
	if sum != DiscSHA256 {
		return errors.New("este disco no es el original de Need for Speed: High Stakes USA (SLUS-00826) que necesita High Speed 4, o la imagen esta modificada/incompleta. Vuelca tu disco original en .cue/.bin (una pista de datos)")
	}
	o.logf("  Disco original USA (SLUS-00826): correcto.")
	o.prog(10)
	// 2. Copiar el juego
	o.logf("2/5 Copiando el juego a %s ...", o.Dest)
	if err := os.MkdirAll(o.Dest, 0o755); err != nil {
		return err
	}
	n, err := copyTree(src, o.Dest, func(n int) { o.prog(10 + min(14, n/100)) })
	if err != nil {
		return err
	}
	o.logf("  %d archivos copiados.", n)
	if err := writeDiscSetting(o.Dest, o.Cue); err != nil {
		return err
	}
	o.prog(25)
	// 3-4. Compilador y preparacion
	if o.Preparar {
		o.logf("3/5 Compilador clang")
		clang, err := ensureClang(o)
		if err != nil {
			o.logf("  AVISO: %v", err)
			o.logf("  Sigo sin preparar: el juego funcionara, pero ira compilando mientras juegas.")
		} else {
			os.WriteFile(filepath.Join(o.Dest, "overlay_toolchain", "compiler.txt"), []byte(tomlPath(clang)+"\r\n"), 0o644)
			o.prog(52)
			o.logf("4/5 Preparando el juego desde tu disco (unos minutos)...")
			if err := prepare(o, clang); err != nil {
				o.logf("  AVISO: %v", err)
			}
		}
	} else {
		o.logf("3/5 y 4/5 Sin preparar (casilla desmarcada): el juego ira compilando mientras juegas.")
	}
	o.prog(95)
	// 5. Accesos directos
	if o.Accesos {
		o.logf("5/5 Creando accesos directos (escritorio y menu Inicio)...")
		if err := shortcuts(o); err != nil {
			o.logf("  AVISO: no se pudieron crear los accesos directos (%v)", err)
		}
	}
	o.prog(100)
	o.logf("")
	o.logf("Listo. Abre High Speed 4 desde el acceso directo o con %s.", filepath.Join(o.Dest, GameExe))
	return nil
}
