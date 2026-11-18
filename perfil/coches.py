# Mapa de coches (multijugador): guarda la RAM cada segundo durante 60 s desde que empieza la carrera.
# Salida: perfil\coches\ram_NNNN.bin
import json, os, socket, time
DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "coches")
os.makedirs(DIR, exist_ok=True)
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(30)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
t0 = time.time(); n = 0; ok = False; fallos = 0; carrera = False
while n < 60:
    try:
        if not carrera:
            v = q({"cmd": "read_ram", "addr": "0x8010FB90", "len": 4})
            open(os.path.join(DIR, "estado.txt"), "a").write("%.0f %s\n" % (time.time(), v.get("hex")))
            if v["hex"][:2] in ("00", "03"):
                time.sleep(1); continue
            carrera = True; time.sleep(8); t0 = time.time()
        r = q({"cmd": "dump_ram", "addr": "0x80000000", "len": 2097152})
        open(os.path.join(DIR, "ram_%04d.bin" % n), "wb").write(bytes.fromhex(r["hex"]))
        ok = True; n += 1; fallos = 0
    except Exception:
        fallos += 1
        if (ok and fallos > 5) or time.time() - t0 > 900: break
    time.sleep(1)
