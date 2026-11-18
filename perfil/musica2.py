# Jukebox - fase 2: vigila quien escribe en las variables candidatas al pausar.
# Salida: perfil/musica/pausas.txt
import json, os, socket, time
DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "musica")
os.makedirs(DIR, exist_ok=True)
def q(cmd, timeout=30):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(timeout)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
RANGOS = [("0x001D7D00", "0x001D8C00"), ("0x0010EA0C", "0x0010EA10"),
          ("0x00110DE4", "0x00110DE8"), ("0x0004189C", "0x000418A0"),
          ("0x0004198C", "0x00041990")]
out = open(os.path.join(DIR, "pausas.txt"), "a", encoding="utf-8")
t0 = time.time(); armado = False; fallos = 0
while True:
    try:
        if not armado:
            q({"cmd": "wtrace_disarm_all"})
            q({"cmd": "wtrace_range", "lo": RANGOS[0][0], "hi": RANGOS[0][1]})
            for lo, hi in RANGOS[1:]:
                q({"cmd": "wtrace_add", "lo": lo, "hi": hi})
            q({"cmd": "wtrace_reset"})
            armado = True
            out.write("armado %d s\n" % int(time.time() - t0)); out.flush()
        r = q({"cmd": "wtrace_dump"})
        h = q({"cmd": "cdrom_command_history"})
        out.write("=== %d s\n%s\n%s\n" % (int(time.time() - t0), json.dumps(r), json.dumps(h)))
        out.flush(); fallos = 0
    except Exception as e:
        fallos += 1
        if (armado and fallos >= 5) or (not armado and time.time() - t0 > 900):
            out.write("fin: %s\n" % e); out.flush(); break
    time.sleep(5)
