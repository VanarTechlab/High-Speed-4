# Vigila el hilo de dibujo (render thread) cada 0,5 s y anota por que se detiene. Log en perfil\hilo.log
import json, os, socket, time, collections
LOG = os.path.join(os.path.dirname(os.path.abspath(__file__)), "hilo.log")
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(10)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
out = open(LOG, "w", buffering=1)
t0 = None; prev = None; vistos = set(); fallos = 0; hprev = None; tick = 0
while True:
    try:
        r = q({"cmd": "render_thread"})
        if t0 is None: t0 = time.time()
        fallos = 0
        razones = collections.Counter()
        for e in r.get("recent_acquires", []):
            k = (e["frame"], e["reason"])
            if k not in vistos:
                vistos.add(k); razones[e["reason"]] += 1
        if prev:
            out.write("t=%6.1f active=%s held=%s acquires=+%d acquire_ms=+%.1f busy_ms=+%.1f frames=+%d razones=%s\n" % (
                time.time() - t0, r.get("active"), r.get("held"),
                r["acquires"] - prev["acquires"], r["acquire_ms"] - prev["acquire_ms"],
                r["render_busy_ms"] - prev["render_busy_ms"],
                r["frames_produced"] - prev["frames_produced"], dict(razones)))
        prev = r
        tick += 1
        if tick % 4 == 0:
            h = q({"cmd": "hd_textures"})
            ks = ["draw_queries","matched_draws","ready_draws","applied_draws","composition_hits","composition_builds",
                  "composition_pixels","gl_uploads","gl_evictions","decoded_images","decode_evictions"]
            if hprev:
                out.write("   HD 2s: " + " ".join("%s=+%d" % (k, h.get(k,0) - hprev.get(k,0)) for k in ks) +
                          " decoded_bytes=%d\n" % h.get("decoded_bytes",0))
            hprev = h
    except Exception as ex:
        fallos += 1
        if t0 is not None and fallos > 10: break
        if t0 is None and fallos > 600: break
    time.sleep(0.5)
