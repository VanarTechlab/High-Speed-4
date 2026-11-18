//go:build gui && windows

package main

// "Instalar High Speed 4.exe": ventana de instalacion (mismo estilo que el instalador de Belial 64).
// VanarTechlab, licencia MIT.
// Solo usa la API de Windows (user32, gdi32, comdlg32, shell32, comctl32) a traves de syscall.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

func init() { runtime.LockOSThread() }

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")

	pRegisterClassExW  = user32.NewProc("RegisterClassExW")
	pCreateWindowExW   = user32.NewProc("CreateWindowExW")
	pDefWindowProcW    = user32.NewProc("DefWindowProcW")
	pGetMessageW       = user32.NewProc("GetMessageW")
	pTranslateMessage  = user32.NewProc("TranslateMessage")
	pDispatchMessageW  = user32.NewProc("DispatchMessageW")
	pIsDialogMessageW  = user32.NewProc("IsDialogMessageW")
	pPostQuitMessage   = user32.NewProc("PostQuitMessage")
	pPostMessageW      = user32.NewProc("PostMessageW")
	pSendMessageW      = user32.NewProc("SendMessageW")
	pShowWindow        = user32.NewProc("ShowWindow")
	pUpdateWindow      = user32.NewProc("UpdateWindow")
	pDestroyWindow     = user32.NewProc("DestroyWindow")
	pMessageBoxW       = user32.NewProc("MessageBoxW")
	pLoadCursorW       = user32.NewProc("LoadCursorW")
	pLoadIconW         = user32.NewProc("LoadIconW")
	pLoadImageW        = user32.NewProc("LoadImageW")
	pGetClientRect     = user32.NewProc("GetClientRect")
	pFillRect          = user32.NewProc("FillRect")
	pAdjustWindowRect  = user32.NewProc("AdjustWindowRect")
	pSetWindowTextW    = user32.NewProc("SetWindowTextW")
	pGetWindowTextW    = user32.NewProc("GetWindowTextW")
	pGetWindowTextLenW = user32.NewProc("GetWindowTextLengthW")
	pEnableWindow      = user32.NewProc("EnableWindow")

	pCreateSolidBrush   = gdi32.NewProc("CreateSolidBrush")
	pCreateFontW        = gdi32.NewProc("CreateFontW")
	pCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	pSelectObject       = gdi32.NewProc("SelectObject")
	pBitBlt             = gdi32.NewProc("BitBlt")
	pDeleteDC           = gdi32.NewProc("DeleteDC")
	pSetTextColor       = gdi32.NewProc("SetTextColor")
	pSetBkColor         = gdi32.NewProc("SetBkColor")
	pSetBkMode          = gdi32.NewProc("SetBkMode")

	pGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	pSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	pSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	pShellExecuteW        = shell32.NewProc("ShellExecuteW")
	pInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	pGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	pGetLogicalDrives     = kernel32.NewProc("GetLogicalDrives")
	pGetDriveTypeW        = kernel32.NewProc("GetDriveTypeW")
	pCoInitialize         = ole32.NewProc("CoInitialize")
	pCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
)

const (
	wsChild       = 0x40000000
	wsVisible     = 0x10000000
	wsBorder      = 0x00800000
	wsVScroll     = 0x00200000
	wsTabStop     = 0x00010000
	wsCaption     = 0x00C00000
	wsSysMenu     = 0x00080000
	wsMinimizeBox = 0x00020000
	esAutoHScroll = 0x80
	esMultiline   = 0x4
	esAutoVScroll = 0x40
	esReadOnly    = 0x800
	bsAutoCheck   = 0x3

	wmCreate         = 0x0001
	wmDestroy        = 0x0002
	wmClose          = 0x0010
	wmEraseBkgnd     = 0x0014
	wmSetFont        = 0x0030
	wmCommand        = 0x0111
	wmCtlColorEdit   = 0x0133
	wmCtlColorStatic = 0x0138
	wmApp            = 0x8000
	wmAppLog         = wmApp + 1
	wmAppProg        = wmApp + 2
	wmAppFin         = wmApp + 3

	emSetSel     = 0x00B1
	emReplaceSel = 0x00C2
	bmGetCheck   = 0x00F0
	bmSetCheck   = 0x00F1
	pbmSetRange  = 0x0401
	pbmSetPos    = 0x0402

	mbIconError   = 0x10
	mbIconWarning = 0x30
	mbIconInfo    = 0x40
	mbYesNo       = 0x4
	idYes         = 6

	idMpqBtn  = 101
	idDestBtn = 102
	idInstall = 103
	idCine    = 104
	idAccesos = 105

	wClient = 640
	hClient = 880
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type msgT struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	ptX, ptY int32
	lPrivate uint32
}

type rect struct{ left, top, right, bottom int32 }

type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

type browseInfo struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

type initCC struct{ size, icc uint32 }

func call(p *syscall.LazyProc, a ...uintptr) uintptr { r, _, _ := p.Call(a...); return r }

func u16(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		p, _ = syscall.UTF16PtrFromString("?")
	}
	return p
}

// lista de cadenas separadas por NUL y terminada en doble NUL (filtros del dialogo)
func u16multi(parts ...string) *uint16 {
	var b []uint16
	for _, s := range parts {
		b = append(b, utf16.Encode([]rune(s))...)
		b = append(b, 0)
	}
	b = append(b, 0)
	return &b[0]
}

func rgb(r, g, b uint32) uintptr { return uintptr(r | g<<8 | b<<16) }

var (
	hInst                                  uintptr
	gWin, gMpq, gDest, gCine, gInst, gProg uintptr
	gLog, gMpqBtn, gDestBtn, gLegal, gAcc  uintptr
	gFondo, gNegro, gEditBr, gLegalBr      uintptr
	gFont, gFontB, gFontS                  uintptr
	gBase                                  string
	gWorking                               bool
	gMu                                    sync.Mutex
	gQueue                                 []string
	gProgVal                               int
	gErr                                   error
	gDestFinal                             string
	wndProcCB                              = syscall.NewCallback(wndProc)
)

func setText(h uintptr, s string) { call(pSetWindowTextW, h, uintptr(unsafe.Pointer(u16(s)))) }

func getText(h uintptr) string {
	n := int(call(pGetWindowTextLenW, h))
	buf := make([]uint16, n+2)
	call(pGetWindowTextW, h, uintptr(unsafe.Pointer(&buf[0])), uintptr(n+1))
	return strings.TrimSpace(syscall.UTF16ToString(buf))
}

func msgBox(text, title string, flags uintptr) uintptr {
	return call(pMessageBoxW, gWin, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), flags)
}

func logAdd(s string) {
	n := call(pGetWindowTextLenW, gLog)
	if n > 60000 {
		setText(gLog, "")
		n = 0
	}
	call(pSendMessageW, gLog, emSetSel, n, n)
	t := strings.ReplaceAll(s, "\n", "\r\n") + "\r\n"
	call(pSendMessageW, gLog, emReplaceSel, 0, uintptr(unsafe.Pointer(u16(t))))
}

func create(class, text string, style uintptr, x, y, w, h int, id int, font uintptr) uintptr {
	c := call(pCreateWindowExW, 0, uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(text))),
		wsChild|wsVisible|style, uintptr(x), uintptr(y), uintptr(w), uintptr(h), gWin, uintptr(id), hInst, 0)
	call(pSendMessageW, c, wmSetFont, font, 1)
	return c
}

func title(t string, y int) { create("STATIC", t, 0, 20, y, 600, 20, 0, gFontB) }

func findCue() string {
	cands := []string{
		filepath.Join(gBase, "Need for Speed - High Stakes (USA).cue"),
		filepath.Join(filepath.Dir(gBase), "Need for Speed - High Stakes (USA).cue"),
	}
	for _, c := range cands {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

func defaultDest() string {
	if b := os.Getenv("LOCALAPPDATA"); b != "" {
		return filepath.Join(b, "Programs", "High Speed 4")
	}
	return filepath.Join(gBase, "High Speed 4")
}

func pickMpq() {
	buf := make([]uint16, 1024)
	o := openFileName{}
	o.lStructSize = uint32(unsafe.Sizeof(o))
	o.hwndOwner = gWin
	o.lpstrFilter = u16multi("Imagen de disco (.cue)", "*.cue", "Todos los archivos", "*.*")
	o.lpstrFile = &buf[0]
	o.nMaxFile = uint32(len(buf))
	o.lpstrTitle = u16("Busca el .cue de tu disco original de Need for Speed: High Stakes (USA)")
	o.flags = 0x1000 | 0x800 | 0x8 // FILEMUSTEXIST | PATHMUSTEXIST | NOCHANGEDIR
	if call(pGetOpenFileNameW, uintptr(unsafe.Pointer(&o))) != 0 {
		setText(gMpq, syscall.UTF16ToString(buf))
	}
}

func pickDest() {
	name := make([]uint16, 520)
	b := browseInfo{hwndOwner: gWin, pszDisplayName: &name[0],
		lpszTitle: u16("Elige la carpeta donde instalar High Speed 4"),
		ulFlags:   0x1 | 0x40} // RETURNONLYFSDIRS | NEWDIALOGSTYLE
	id := call(pSHBrowseForFolderW, uintptr(unsafe.Pointer(&b)))
	if id == 0 {
		return
	}
	path := make([]uint16, 1024)
	if call(pSHGetPathFromIDListW, id, uintptr(unsafe.Pointer(&path[0]))) != 0 {
		setText(gDest, syscall.UTF16ToString(path))
	}
	call(pCoTaskMemFree, id)
}

func lock(on bool) {
	en := uintptr(1)
	if on {
		en = 0
	}
	for _, c := range []uintptr{gMpq, gMpqBtn, gDest, gDestBtn, gCine, gAcc, gInst} {
		call(pEnableWindow, c, en)
	}
}

func install() {
	cue := getText(gMpq)
	dest := getText(gDest)
	if !fileExists(cue) || !strings.EqualFold(filepath.Ext(cue), ".cue") {
		msgBox("Primero busca el archivo .cue de tu disco con el botón \"Buscar...\".\n\nHigh Speed 4 no incluye el juego: necesitas tu disco original de Need for Speed: High Stakes (PlayStation, USA, SLUS-00826) volcado a .cue + .bin.", "Falta el disco", mbIconWarning)
		return
	}
	if dest == "" {
		msgBox("Elige dónde instalar con el botón \"Elegir...\".", "Falta el destino", mbIconWarning)
		return
	}
	prep := call(pSendMessageW, gCine, bmGetCheck, 0, 0) == 1
	acc := call(pSendMessageW, gAcc, bmGetCheck, 0, 0) == 1
	setText(gLog, "")
	call(pSendMessageW, gProg, pbmSetPos, 1, 0)
	gWorking = true
	gDestFinal = dest
	lock(true)
	o := &Options{Base: gBase, Cue: cue, Dest: dest, Preparar: prep, Accesos: acc,
		Log: func(s string) {
			gMu.Lock()
			gQueue = append(gQueue, s)
			gMu.Unlock()
			call(pPostMessageW, gWin, wmAppLog, 0, 0)
		},
		Progress: func(p int) {
			gMu.Lock()
			gProgVal = p
			gMu.Unlock()
			call(pPostMessageW, gWin, wmAppProg, 0, 0)
		}}
	go func() {
		err := Run(o)
		gMu.Lock()
		gErr = err
		if err != nil {
			gQueue = append(gQueue, "", "*** ERROR: "+err.Error())
		}
		gMu.Unlock()
		call(pPostMessageW, gWin, wmAppLog, 0, 0)
		call(pPostMessageW, gWin, wmAppFin, 0, 0)
	}()
}

const legalText = "High Speed 4 es GRATUITO y de LIBRE DISTRIBUCIÓN (si pagaste por él, te han estafado). " +
	"Es un proyecto NO OFICIAL hecho por un fan (Vanar), sin afiliación ni respaldo de Electronic Arts ni de Sony. " +
	"Need for Speed y High Stakes son marcas de Electronic Arts; PlayStation es una marca de Sony Interactive Entertainment.\r\n" +
	"No incluye el juego, ni código ni arte de EA, ni BIOS de Sony: usa TU propio disco original, y el código que se " +
	"traduce y compila desde él es solo para tu uso personal (no lo compartas). " +
	"Hecho con PSXRecomp (PolyForm Noncommercial + permiso de donaciones) y recomp-ui; usa OpenBIOS (MIT) y el compilador clang (LLVM). Ver LICENCIAS."

func wndProc(hw, msg, wp, lp uintptr) uintptr {
	switch msg {
	case wmCreate:
		gWin = hw
		y := 306
		title("1.  Tu disco: el .cue de Need for Speed: High Stakes (PlayStation, USA, SLUS-00826)", y)
		gMpq = create("EDIT", findCue(), wsBorder|esAutoHScroll|wsTabStop, 20, y+22, 480, 24, 0, gFont)
		gMpqBtn = create("BUTTON", "Buscar...", wsTabStop, 510, y+21, 110, 26, idMpqBtn, gFont)
		y = 370
		title("2.  ¿Dónde lo instalo? (cualquier carpeta de tu PC)", y)
		gDest = create("EDIT", defaultDest(), wsBorder|esAutoHScroll|wsTabStop, 20, y+22, 480, 24, 0, gFont)
		gDestBtn = create("BUTTON", "Elegir...", wsTabStop, 510, y+21, 110, 26, idDestBtn, gFont)
		gCine = create("BUTTON", "Descargar el compilador y preparar el juego ahora (recomendado; unos minutos)", bsAutoCheck|wsTabStop, 20, y+54, 600, 22, idCine, gFont)
		call(pSendMessageW, gCine, bmSetCheck, 1, 0)
		gAcc = create("BUTTON", "Crear accesos directos (escritorio y menú Inicio)", bsAutoCheck|wsTabStop, 20, y+78, 600, 22, idAccesos, gFont)
		call(pSendMessageW, gAcc, bmSetCheck, 1, 0)
		y = 480
		title("Avisos legales", y)
		gLegal = create("EDIT", legalText, wsBorder|esMultiline|esReadOnly, 20, y+22, 600, 112, 0, gFontS)
		gInst = create("BUTTON", "INSTALAR", wsTabStop, 220, 624, 200, 38, idInstall, gFontB)
		gProg = create("msctls_progress32", "", 0, 20, 674, 600, 16, 0, gFont)
		call(pSendMessageW, gProg, pbmSetRange, 0, 100<<16)
		gLog = create("EDIT", "", wsBorder|wsVScroll|esMultiline|esAutoVScroll|esReadOnly, 20, 698, 600, 166, 0, gFont)
		logAdd("Bienvenido a High Speed 4 v" + Version + " (Need for Speed: High Stakes para PC), por Vanar.")
		logAdd("1) Busca el .cue de tu disco.  2) Elige la carpeta.  3) Pulsa INSTALAR.")
		logAdd("Al preparar, el código del juego se traduce y compila desde TU disco (nada del juego viene en la descarga).")
		return 0
	case wmEraseBkgnd:
		var r rect
		call(pGetClientRect, hw, uintptr(unsafe.Pointer(&r)))
		call(pFillRect, wp, uintptr(unsafe.Pointer(&r)), gNegro)
		if gFondo != 0 {
			m := call(pCreateCompatibleDC, wp)
			old := call(pSelectObject, m, gFondo)
			call(pBitBlt, wp, 0, 0, 640, 300, m, 0, 0, 0x00CC0020)
			call(pSelectObject, m, old)
			call(pDeleteDC, m)
		}
		return 1
	case wmCtlColorStatic, wmCtlColorEdit:
		c := lp
		if c == gLegal {
			call(pSetTextColor, wp, rgb(255, 214, 170))
			call(pSetBkColor, wp, rgb(46, 20, 52))
			return gLegalBr
		}
		if c == gLog || c == gMpq || c == gDest {
			call(pSetTextColor, wp, rgb(235, 228, 245))
			call(pSetBkColor, wp, rgb(24, 18, 36))
			return gEditBr
		}
		call(pSetTextColor, wp, rgb(255, 170, 90))
		call(pSetBkColor, wp, rgb(14, 10, 24))
		call(pSetBkMode, wp, 1) // TRANSPARENT
		return gNegro
	case wmCommand:
		if gWorking {
			return 0
		}
		switch wp & 0xFFFF {
		case idMpqBtn:
			pickMpq()
		case idDestBtn:
			pickDest()
		case idInstall:
			install()
		}
		return 0
	case wmAppLog:
		gMu.Lock()
		q := gQueue
		gQueue = nil
		gMu.Unlock()
		for _, s := range q {
			logAdd(s)
		}
		return 0
	case wmAppProg:
		gMu.Lock()
		p := gProgVal
		gMu.Unlock()
		call(pSendMessageW, gProg, pbmSetPos, uintptr(p), 0)
		return 0
	case wmAppFin:
		gWorking = false
		lock(false)
		gMu.Lock()
		err := gErr
		gMu.Unlock()
		if err == nil {
			call(pSendMessageW, gProg, pbmSetPos, 100, 0)
			if msgBox("¡Listo! High Speed 4 está instalado en:\n"+gDestFinal+"\n\n¿Quieres jugar ahora?", "High Speed 4", mbIconInfo|mbYesNo) == idYes {
				call(pShellExecuteW, hw, uintptr(unsafe.Pointer(u16("open"))), uintptr(unsafe.Pointer(u16(filepath.Join(gDestFinal, GameExe)))), 0, uintptr(unsafe.Pointer(u16(gDestFinal))), 1)
			}
		} else {
			call(pSendMessageW, gProg, pbmSetPos, 0, 0)
			msgBox("Algo falló. Lee el mensaje de ERROR en el cuadro de abajo.", "High Speed 4", mbIconError)
		}
		return 0
	case wmClose:
		if gWorking && msgBox("La instalación sigue en marcha. ¿Cerrar de todas formas?", "High Speed 4", mbYesNo|mbIconWarning) != idYes {
			return 0
		}
		call(pDestroyWindow, hw)
		return 0
	case wmDestroy:
		call(pPostQuitMessage, 0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hw, msg, wp, lp)
	return r
}

func main() {
	icc := initCC{8, 0x20} // ICC_PROGRESS_CLASS
	call(pInitCommonControlsEx, uintptr(unsafe.Pointer(&icc)))
	call(pCoInitialize, 0)
	exe, _ := os.Executable()
	gBase = filepath.Dir(exe)
	hInst = call(pGetModuleHandleW, 0)
	gFondo = call(pLoadImageW, 0, uintptr(unsafe.Pointer(u16(filepath.Join(gBase, "herramientas", "fondo.bmp")))), 0, 0, 0, 0x10)
	gNegro = call(pCreateSolidBrush, rgb(14, 10, 24))
	gEditBr = call(pCreateSolidBrush, rgb(24, 18, 36))
	gLegalBr = call(pCreateSolidBrush, rgb(46, 20, 52))
	seg := uintptr(unsafe.Pointer(u16("Segoe UI")))
	gFont = call(pCreateFontW, uintptr(0xFFFFFFF1), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, seg)  // -15
	gFontB = call(pCreateFontW, uintptr(0xFFFFFFF0), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 5, 0, seg) // -16
	gFontS = call(pCreateFontW, uintptr(0xFFFFFFF3), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, seg) // -13

	cls := u16("HighSpeed4Instalador")
	wc := wndClassEx{
		lpfnWndProc:   wndProcCB,
		hInstance:     hInst,
		hIcon:         call(pLoadIconW, 0, 32512),
		hCursor:       call(pLoadCursorW, 0, 32512),
		hbrBackground: gNegro,
		lpszClassName: cls,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	call(pRegisterClassExW, uintptr(unsafe.Pointer(&wc)))
	style := uintptr(wsCaption | wsSysMenu | wsMinimizeBox)
	r := rect{0, 0, wClient, hClient}
	call(pAdjustWindowRect, uintptr(unsafe.Pointer(&r)), style, 0)
	w := call(pCreateWindowExW, 0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16("High Speed 4 v"+Version+" - Instalador"))),
		style, uintptr(0x80000000), uintptr(0x80000000), uintptr(r.right-r.left), uintptr(r.bottom-r.top), 0, 0, hInst, 0)
	call(pShowWindow, w, 1)
	call(pUpdateWindow, w)
	var m msgT
	for {
		ret := int32(call(pGetMessageW, uintptr(unsafe.Pointer(&m)), 0, 0, 0))
		if ret <= 0 {
			break
		}
		if call(pIsDialogMessageW, w, uintptr(unsafe.Pointer(&m))) == 0 {
			call(pTranslateMessage, uintptr(unsafe.Pointer(&m)))
			call(pDispatchMessageW, uintptr(unsafe.Pointer(&m)))
		}
	}
}
