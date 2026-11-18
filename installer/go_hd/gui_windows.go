//go:build gui && windows

package main

// "Texturas HD.exe": kit HD de High Speed 4 (mismo estilo que el instalador). VanarTechlab, licencia MIT.
// Solo usa la API de Windows (user32, gdi32, comdlg32, shell32, comctl32) a traves de syscall.

import (
	"fmt"
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

	idRecoger = 101
	idCrear   = 102
	idOnOff   = 103
	idPack    = 104

	wClient = 640
	hClient = 850
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
	hInst                                          uintptr
	gWin, gEstado, gExpl, gRec, gCrear, gOn, gPack uintptr
	gProg, gLog                                    uintptr
	gFondo, gNegro, gEditBr, gLegalBr              uintptr
	gFont, gFontB, gFontS                          uintptr
	gKit                                           *Kit
	gWorking                                       bool
	gMu                                            sync.Mutex
	gQueue                                         []string
	gProgVal                                       int
	gErr                                           error
	wndProcCB                                      = syscall.NewCallback(wndProc)
)

func setText(h uintptr, s string) { call(pSetWindowTextW, h, uintptr(unsafe.Pointer(u16(s)))) }

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

func refresh() {
	c := gKit.Count()
	en, du := gKit.HDState()
	est := "Texturas HD: DESACTIVADAS"
	if en {
		est = "Texturas HD: ACTIVADAS"
	}
	rec := "Recogida: apagada"
	if du {
		rec = "Recogida: ENCENDIDA (juega y cierra el juego)"
	}
	setText(gEstado, fmt.Sprintf("%s   |   %s\r\nRecogidas: %d   Instaladas en HD: %d   Nuevas por convertir: %d", est, rec, c.Recogidas, c.Instaladas, c.Pendientes))
	if du {
		setText(gRec, "Parar la recogida")
	} else {
		setText(gRec, "B1.  Recoger texturas")
	}
	if en {
		setText(gOn, "Desactivar HD")
	} else {
		setText(gOn, "Activar HD")
	}
}

func lock(on bool) {
	en := uintptr(1)
	if on {
		en = 0
	}
	for _, c := range []uintptr{gPack, gRec, gCrear, gOn} {
		call(pEnableWindow, c, en)
	}
}

const explText = "Dos formas de tener texturas HD (pulsa los botones con el juego CERRADO):\r\n" +
	"A) Recomendado: descarga el pack de texturas HD de High Speed 4 (zip aparte, en la web de VanarTechlab) y pulsa \"Instalar pack descargado\". Listo.\r\n" +
	"B) Crearlas tú desde tu partida (sin descargar el pack; necesita gráfica con Vulkan): 1. \"Recoger texturas\", juega y CIERRA el juego (se guardan al salir). " +
	"2. \"Crear e instalar texturas HD\" (IA Real-ESRGAN). Repítelo cuando juegues algo nuevo: solo convierte lo nuevo.\r\n" +
	"La próxima vez que abras el juego ya se ven en HD. \"Desactivar HD\" vuelve a las originales sin borrar nada."

func trabajar(f func() error) {
	setText(gLog, "")
	call(pSendMessageW, gProg, pbmSetPos, 1, 0)
	gWorking = true
	lock(true)
	go func() {
		err := f()
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

func pickZip() string {
	buf := make([]uint16, 1024)
	o := openFileName{}
	o.lStructSize = uint32(unsafe.Sizeof(o))
	o.hwndOwner = gWin
	o.lpstrFilter = u16multi("Pack de texturas HD (.zip)", "*.zip", "Todos los archivos", "*.*")
	o.lpstrFile = &buf[0]
	o.nMaxFile = uint32(len(buf))
	o.lpstrTitle = u16("Elige el zip del pack de texturas HD de High Speed 4")
	o.flags = 0x1000 | 0x800 | 0x8
	if call(pGetOpenFileNameW, uintptr(unsafe.Pointer(&o))) != 0 {
		return syscall.UTF16ToString(buf)
	}
	return ""
}

func wndProc(hw, msg, wp, lp uintptr) uintptr {
	switch msg {
	case wmCreate:
		gWin = hw
		gEstado = create("STATIC", "", 0, 20, 306, 600, 40, 0, gFontB)
		gExpl = create("EDIT", explText, wsBorder|esMultiline|esReadOnly, 20, 352, 600, 150, 0, gFontS)
		gPack = create("BUTTON", "A)  Instalar pack descargado (.zip)  -  recomendado", wsTabStop, 20, 512, 600, 38, idPack, gFontB)
		gRec = create("BUTTON", "B1.  Recoger texturas", wsTabStop, 20, 558, 190, 34, idRecoger, gFont)
		gCrear = create("BUTTON", "B2.  Crear e instalar texturas HD", wsTabStop, 220, 558, 270, 34, idCrear, gFont)
		gOn = create("BUTTON", "Activar HD", wsTabStop, 500, 558, 120, 34, idOnOff, gFont)
		gProg = create("msctls_progress32", "", 0, 20, 602, 600, 16, 0, gFont)
		call(pSendMessageW, gProg, pbmSetRange, 0, 100<<16)
		gLog = create("EDIT", "", wsBorder|wsVScroll|esMultiline|esAutoVScroll|esReadOnly, 20, 626, 600, 208, 0, gFont)
		logAdd("Kit HD de High Speed 4, por Vanar. Cierra el juego antes de usar estos botones.")
		refresh()
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
		if c == gExpl {
			call(pSetTextColor, wp, rgb(255, 214, 170))
			call(pSetBkColor, wp, rgb(46, 20, 52))
			return gLegalBr
		}
		if c == gLog {
			call(pSetTextColor, wp, rgb(235, 228, 245))
			call(pSetBkColor, wp, rgb(24, 18, 36))
			return gEditBr
		}
		call(pSetTextColor, wp, rgb(255, 170, 90))
		call(pSetBkColor, wp, rgb(14, 10, 24))
		call(pSetBkMode, wp, 1)
		return gNegro
	case wmCommand:
		if gWorking {
			return 0
		}
		en, du := gKit.HDState()
		switch wp & 0xFFFF {
		case idRecoger:
			if du {
				gKit.SetHD(en, false)
				logAdd("Recogida apagada.")
			} else {
				gKit.SetHD(true, true)
				logAdd("Recogida ENCENDIDA. Abre High Speed 4, juega y luego CIERRA el juego. Después pulsa B2.")
			}
		case idCrear:
			trabajar(gKit.Crear)
			return 0
		case idPack:
			if z := pickZip(); z != "" {
				trabajar(func() error { return gKit.InstalarPack(z) })
			}
			return 0
		case idOnOff:
			gKit.SetHD(!en, false)
			if en {
				logAdd("Texturas HD desactivadas (no se borra nada).")
			} else {
				logAdd("Texturas HD activadas.")
			}
		}
		refresh()
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
		refresh()
		gMu.Lock()
		err := gErr
		gMu.Unlock()
		if err == nil {
			call(pSendMessageW, gProg, pbmSetPos, 100, 0)
			msgBox("¡Listo! Las texturas HD están instaladas y activadas.\nAbre High Speed 4 y ya se verán.", "High Speed 4 - Texturas HD", mbIconInfo)
		} else {
			call(pSendMessageW, gProg, pbmSetPos, 0, 0)
			msgBox("Algo falló. Lee el mensaje de ERROR en el cuadro de abajo.", "High Speed 4 - Texturas HD", mbIconError)
		}
		return 0
	case wmClose:
		if gWorking && msgBox("La conversión sigue en marcha. ¿Cerrar de todas formas?", "High Speed 4", mbYesNo|mbIconWarning) != idYes {
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
	icc := initCC{8, 0x20}
	call(pInitCommonControlsEx, uintptr(unsafe.Pointer(&icc)))
	exe, _ := os.Executable()
	base := filepath.Dir(exe)
	gKit = &Kit{Base: base,
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
	hInst = call(pGetModuleHandleW, 0)
	gFondo = call(pLoadImageW, 0, uintptr(unsafe.Pointer(u16(filepath.Join(base, "hd_kit", "fondo.bmp")))), 0, 0, 0, 0x10)
	gNegro = call(pCreateSolidBrush, rgb(14, 10, 24))
	gEditBr = call(pCreateSolidBrush, rgb(24, 18, 36))
	gLegalBr = call(pCreateSolidBrush, rgb(46, 20, 52))
	seg := uintptr(unsafe.Pointer(u16("Segoe UI")))
	gFont = call(pCreateFontW, uintptr(0xFFFFFFF1), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, seg)
	gFontB = call(pCreateFontW, uintptr(0xFFFFFFF0), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 5, 0, seg)
	gFontS = call(pCreateFontW, uintptr(0xFFFFFFF2), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, seg)
	cls := u16("HighSpeed4TexturasHD")
	wc := wndClassEx{lpfnWndProc: wndProcCB, hInstance: hInst, hIcon: call(pLoadIconW, 0, 32512),
		hCursor: call(pLoadCursorW, 0, 32512), hbrBackground: gNegro, lpszClassName: cls}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	call(pRegisterClassExW, uintptr(unsafe.Pointer(&wc)))
	style := uintptr(wsCaption | wsSysMenu | wsMinimizeBox)
	r := rect{0, 0, wClient, hClient}
	call(pAdjustWindowRect, uintptr(unsafe.Pointer(&r)), style, 0)
	w := call(pCreateWindowExW, 0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16("High Speed 4 - Texturas HD"))),
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
