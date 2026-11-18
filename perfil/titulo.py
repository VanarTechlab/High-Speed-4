# Al detectar la pausa, guarda 3 fotos de la RAM (a los 2, 6 y 10 s) en perfil/titulo/.
import json, os, socket, time
DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "titulo")
os.makedirs(DIR, exist_ok=True)
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(30)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
t0 = time.time()
while time.time() - t0 < 1800:
    try:
        v = q({"cmd": "read_ram", "addr": "0x80118234", "len": 1})
        if v["hex"][:2] != "00":
            for k, espera in enumerate((2, 4, 4)):
                time.sleep(espera)
                r = q({"cmd": "dump_ram", "addr": "0x80000000", "len": 2097152})
                open(os.path.join(DIR, "pausa_%d.bin" % k), "wb").write(bytes.fromhex(r["hex"]))
            break
    except Exception:
        pass
    time.sleep(0.5)
