# Prueba 3: carrera de 2 jugadores + 2 rivales IA "de verdad".
# En menus (0x8010FB90 = 3), con la config en modo 2 jugadores, escribe:
#  - config de carrera 0x80114868 (+copia 0x801187E8): total 4, rivales 2
#  - fichas de piloto 2 y 3 (0x80114878 + slot*0xB4, 0xB4 bytes): copiadas de una carrera 1J con rivales
#    (foto c6: +4 = 2 = rival IA). En 2J el menu las deja a 0 y el juego no los registra como corredores.
import json, os, socket, time
BASE = os.path.dirname(os.path.abspath(__file__))
FOTO = open(os.path.join(BASE, "fotos", "c6.bin"), "rb").read()
DI = 0x80114878; SZ = 0xB4
import sys
TIPO = 1 if (len(sys.argv) > 1 and sys.argv[1] in ("humano", "controles")) else 2   # 1 = humano, 2 = rival IA
CONTROLES = len(sys.argv) > 1 and sys.argv[1] == "controles"
HUM4 = len(sys.argv) > 1 and sys.argv[1] == "hum4"     # config con 4 humanos (y 0 rivales)
if HUM4: TIPO = 1
# Controles por jugador (MEDIDO en 0x8009CF3C): 4 bytes en 0x8013E4FC + 4*jugador: [direccion, acelerador, freno, ?]
# Prueba: jugador 3 acelera recto, jugador 4 acelera girando.
PRUEBA = {0x8013E504: 0x00, 0x8013E505: 0xC0, 0x8013E508: 0x40, 0x8013E509: 0xC0}
fichas = {}
for s in (2, 3):
    b = bytearray(FOTO[DI + s * SZ - 0x80000000: DI + (s + 1) * SZ - 0x80000000])
    b[4] = TIPO
    if TIPO == 1:
        b[0x54] = s + 1          # numero de jugador (en 2J: ficha0 = 1, ficha1 = 2) -> 3 y 4
    fichas[s] = bytes(b)
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(10)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
def rd(a, n=1): return bytes.fromhex(q({"cmd": "read_ram", "addr": "0x%08X" % a, "len": n})["hex"])
def wr(a, v): q({"cmd": "write_ram", "addr": "0x%08X" % a, "val": "%x" % v})
print("Forzando 2 coches extra de tipo %s. Deja esta ventana abierta." % ("HUMANO" if TIPO == 1 else "rival IA"))
ultimo = None
while True:
    try:
        st = rd(0x8010FB90)[0]
        if st == 3 and rd(0x8011486C)[0] in (2, 4):
            for base in (0x80114868, 0x801187E8):
                if rd(base + 4)[0] in (2, 4):
                    wr(base, 4)
                    if HUM4: wr(base + 4, 4); wr(base + 8, 0)
                    else: wr(base + 8, 2)
            for s, data in fichas.items():
                a = DI + s * SZ
                cur = rd(a, SZ)
                for k in range(SZ):
                    if cur[k] != data[k]: wr(a + k, data[k])
        if CONTROLES and st == 2:
            # cada 3 s alterna: acelera girando a la izquierda / suelta todo y gira a la derecha
            fase = int(time.time() / 3) % 2
            for j in (2, 3):                       # jugadores 3 y 4 (base 0)
                a = 0x8013E4FC + 4 * j
                wr(a, 0xC0 if fase == 0 else 0x40)  # direccion
                wr(a + 1, 0xC0 if fase == 0 else 0) # acelerador
                wr(a + 2, 0)                        # freno
                wr(0x8013E510 + 4 * j, 1)           # "hay entrada"
        info = (st, rd(0x80114868)[0], rd(0x80114870)[0], rd(DI + 2 * SZ + 4)[0], rd(DI + 3 * SZ + 4)[0], rd(0x8013DAE8)[0], rd(0x8013DAEC)[0], rd(0x8013DAF0)[0])
        if info != ultimo:
            print(time.strftime("%H:%M:%S"), "estado %d | config total %d rivales %d | ficha2 %d ficha3 %d | carrera total %d corriendo %d rivales %d" % info)
            ultimo = info
    except Exception:
        pass
    time.sleep(0.3)
