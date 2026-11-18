# Mide la cobertura de PGXP (vertices precisos vs nativos) cada 2 s en carrera. Log en perfil\pgxp.log
import json, os, socket, time
LOG = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pgxp.log")
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(10)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
def flat(d, pre=""):
    o = {}
    for k, v in d.items():
        if isinstance(v, dict): o.update(flat(v, pre + k + "."))
        elif isinstance(v, (int, float)) and not isinstance(v, bool): o[pre + k] = v
    return o
out = open(LOG, "w", buffering=1)
prev = None; t0 = None; fallos = 0
KEYS = ["pgxp.tri_precise", "pgxp.tri_mixed", "pgxp.tri_native", "pgxp.dataflow_hit", "pgxp.fallback_hit",
        "pgxp.native", "pgxp.value_mismatch", "pgxp.trunc_reject", "pgxp.tolerance_reject", "perspective_triangles"]
while True:
    try:
        r = q({"cmd": "geom_correction"})
        if t0 is None:
            t0 = time.time()
            out.write("estado: " + json.dumps(r) + "\n")
        f = flat(r); fallos = 0
        if prev:
            out.write("t=%6.1f " % (time.time() - t0) + " ".join(
                "%s=+%d" % (k.split(".")[-1], f.get(k, 0) - prev.get(k, 0)) for k in KEYS) + "\n")
        prev = f
    except Exception:
        fallos += 1
        if (t0 is not None and fallos > 10) or (t0 is None and fallos > 600): break
    time.sleep(2)
