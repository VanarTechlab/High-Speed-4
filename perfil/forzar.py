# Prueba: carrera de 2 jugadores con rivales IA. Mientras se esta en menus (0x8010FB90 = 3) escribe en la
# configuracion de carrera: total de coches = 2 + R, humanos = 2 (lo pone el menu), rivales IA = R.
# Estructuras (MEDIDO con fotos c1/c2/c6/j2): 0x80114868 y su copia 0x801187E8: +0 total, +4 humanos, +8 rivales.
# Uso: forzar.py R   (R = rivales IA, por defecto 2)
import json, socket, sys, time
R = int(sys.argv[1]) if len(sys.argv) > 1 else 2
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(10)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
def rd(a): return q({"cmd": "read_ram", "addr": "0x%08X" % a, "len": 1})["hex"]
print("Forzando %d rivales IA (total %d) en los menus. Cierra esta ventana al terminar." % (R, 2 + R))
ultimo = None
inicio = None   # momento en que empezo la carrera (estado 02)
CORREGIR = len(sys.argv) > 2 and sys.argv[2] == "corregir"
while True:
    try:
        st = q({"cmd": "read_ram", "addr": "0x8010FB90", "len": 4})["hex"][:2]
        if st == "03":
            for base in (0x80114868, 0x801187E8):
                humanos = int(rd(base + 4), 16)
                if humanos == 2:
                    q({"cmd": "write_ram", "addr": "0x%08X" % base, "val": "%x" % (2 + R)})
                    q({"cmd": "write_ram", "addr": "0x%08X" % (base + 8), "val": "%x" % R})
        if st == "02" and inicio is None: inicio = time.time()
        if st != "02": inicio = None
        if CORREGIR and inicio is not None and time.time() - inicio < 5:
            # estado de la carrera como en 1J con rivales: [total, total, rivales, humanos, rivales] y +0x40 = humanos
            for a, v in ((0x8013DAEC, 2 + R), (0x8013DAF0, R), (0x8013DAF8, R), (0x8013DB28, 2)):
                q({"cmd": "write_ram", "addr": "0x%08X" % a, "val": "%x" % v})
        info = (st, rd(0x80114868), rd(0x8011486C), rd(0x80114870), rd(0x8013DAE8), rd(0x8013DAF0), rd(0x8013DAEC), rd(0x8013DB28))
        if info != ultimo:
            print(time.strftime("%H:%M:%S"), "estado %s | config total %s humanos %s rivales %s | carrera total %s rivales %s corriendo %s humanos(+0x40) %s" % info)
            ultimo = info
    except Exception:
        pass
    time.sleep(0.2)
