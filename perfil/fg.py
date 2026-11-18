# Vigila Smooth motion (frame generation) cada 2 s: motivos de rechazo, breaker, flips. Log en perfil\fg.log
import json, os, socket, time
LOG = os.path.join(os.path.dirname(os.path.abspath(__file__)), "fg.log")
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(10)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
out = open(LOG, "w", buffering=1); t0 = None; fallos = 0; tick = 0
while True:
    try:
        r = q({"cmd": "frame_gen"})
        if t0 is None: t0 = time.time()
        fallos = 0; tick += 1
        out.write("t=%6.1f %s\n" % (time.time() - t0, json.dumps(r)))
        if tick % 5 == 0:
            rt = q({"cmd": "render_thread"})
            out.write("   RT: " + json.dumps(rt.get("recent_acquires", [])[-6:]) + "\n")
    except Exception:
        fallos += 1
        if (t0 is not None and fallos > 10) or (t0 is None and fallos > 600): break
    time.sleep(2)
