# Captura los comandos GPU (GP0) de varios frames de una carrera de 2 jugadores, con la funcion del juego
# que emitio cada uno (pc/ra/func). -> perfil\vistas_frameN.json
import json, os, socket, time
BASE = os.path.dirname(os.path.abspath(__file__))
LOG = open(os.path.join(BASE, "vistas.log"), "w", buffering=1)
def q(cmd, timeout=60):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(timeout)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        f = s.makefile("rb"); return json.loads(f.readline())
inicio = None; n = 0
while n < 1:
    try:
        st = bytes.fromhex(q({"cmd": "read_ram", "addr": "0x8010FB90", "len": 1})["hex"])[0]
        if st != 2: inicio = None; time.sleep(0.5); continue
        if inicio is None: inicio = time.time(); LOG.write("carrera 2J\n")
        if time.time() - inicio < 8 + 4 * n: time.sleep(0.5); continue
        rs = q({"cmd": "gpu_ring_stats"}); LOG.write("ring %s\n" % rs)
        todos = {}
        for k in range(1, 9):
            fr = rs["newest_frame"] - k
            d = q({"cmd": "gpu_frame_dump", "frame": fr, "count": 65536}, 120)
            todos[fr] = d; LOG.write("frame %d: %d comandos\n" % (fr, len(d.get("entries", []))))
        json.dump(todos, open(os.path.join(BASE, "vistas_frame%d.json" % n), "w")); LOG.write("guardado %d\n" % n)
        n += 1
    except Exception as e:
        LOG.write("err %s\n" % e); time.sleep(1)
