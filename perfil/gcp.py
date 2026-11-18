# Perfil del codigo del JUEGO (ciclos de la PS1) de los ultimos 600 fotogramas.
import json, socket, sys
PORT = 4370
def q(**kw):
    with socket.create_connection(("127.0.0.1", PORT), 5) as s:
        s.settimeout(30)
        s.sendall((json.dumps(dict(id=1, cmd="gcp_ring", **kw)) + "\n").encode())
        with s.makefile("rb") as f:
            return json.loads(f.readline())
try:
    h = q()
    hi = h["newest_frame"]; lo = max(h["oldest_frame"], hi - 600)
    for key, nom in (("fn", "funcion"), ("block", "bloque"), ("ra", "llamador")):
        r = q(key=key, frame_lo=lo, frame_hi=hi, top=20)
        print("-- por %s (frames %d-%d, peso %s)" % (nom, lo, hi, r.get("weight")))
        for e in r.get("top", []):
            print("  %s %5.1f%%" % (e["addr"], 100 * e["share"]))
    r = None
    with socket.create_connection(("127.0.0.1", PORT), 5) as s:
        s.sendall(b'{"id":2,"cmd":"idle_skip"}\n')
        print("-- idle_skip:", s.makefile("rb").readline().decode().strip())
except Exception as e:
    print("gcp error:", e)
