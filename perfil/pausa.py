# Busca el indicador de pausa: guarda la RAM cada segundo durante ~90 s, empezando al entrar en carrera.
import json, os, socket, time
DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pausa2")
os.makedirs(DIR, exist_ok=True)
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(30)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
t0 = time.time(); n = 0; ok = False; fallos = 0; carrera = False
while n < 90:
    try:
        if not carrera:
            v = q({"cmd": "read_ram", "addr": "0x8010FB90", "len": 4})
            if v["hex"][:2] != "01":
                time.sleep(1); continue
            carrera = True; t0 = time.time()
        r = q({"cmd": "dump_ram", "addr": "0x80000000", "len": 2097152})
        open(os.path.join(DIR, "ram_%04d.bin" % int(time.time() - t0)), "wb").write(bytes.fromhex(r["hex"]))
        ok = True; n += 1; fallos = 0
    except Exception:
        fallos += 1
        if (ok and fallos > 5) or time.time() - t0 > 900: break
    time.sleep(1)
