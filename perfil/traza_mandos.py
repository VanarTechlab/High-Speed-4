# Traza que mando lee cada jugador (build-perfil): fuerza 4 humanos (como forzar_ia.py humano) y registra las
# llamadas a las funciones de lectura de mando con su argumento (a0 = indice de mando). -> perfil\traza_mandos.json
import json, os, socket, time
BASE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(BASE, "traza_mandos.json"); LOG = open(os.path.join(BASE, "traza_mandos.log"), "w", buffering=1)
FOTO = open(os.path.join(BASE, "fotos", "c6.bin"), "rb").read()
DI = 0x80114878; SZ = 0xB4
fichas = {}
for s in (2, 3):
    b = bytearray(FOTO[DI + s * SZ - 0x80000000: DI + (s + 1) * SZ - 0x80000000]); b[4] = 1; b[0x54] = s + 1; fichas[s] = bytes(b)
TARGETS = ["0x800BE42C", "0x800BE494", "0x800E4DA4", "0x800BE1A8"]
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(60)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
def rd(a, n=1): return bytes.fromhex(q({"cmd": "read_ram", "addr": "0x%08X" % a, "len": n})["hex"])
def wr(a, v): q({"cmd": "write_ram", "addr": "0x%08X" % a, "val": "%x" % v})
armado = False; inicio = None
while True:
    try:
        if not armado:
            for t in TARGETS: LOG.write("arm %s\n" % q({"cmd": "fntrace_arm", "target": t}))
            armado = True
        st = rd(0x8010FB90)[0]
        if st == 3:
            inicio = None
            if rd(0x8011486C)[0] == 2:
                for base in (0x80114868, 0x801187E8):
                    if rd(base + 4)[0] == 2: wr(base, 4); wr(base + 8, 2)
                for s, data in fichas.items():
                    a = DI + s * SZ; cur = rd(a, SZ)
                    for k in range(SZ):
                        if cur[k] != data[k]: wr(a + k, data[k])
        elif st == 2:
            if inicio is None: inicio = time.time(); LOG.write("carrera\n"); q({"cmd": "fntrace_clear"})
            elif time.time() - inicio > 8:
                d = q({"cmd": "fntrace_dump", "count": 20000})
                json.dump(d, open(OUT, "w")); LOG.write("guardado\n"); break
    except Exception as e:
        LOG.write("err %s\n" % e); time.sleep(1)
    time.sleep(0.3)
