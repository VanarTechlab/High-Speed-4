# Menu de pausa: guarda la RAM cada 2 s para encontrar el estado del menu.
# Salida: perfil/musica/ram_<seg>.bin y perfil/musica/cd.txt
import json, os, socket, sys, time
DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "menu")
os.makedirs(DIR, exist_ok=True)
def q(cmd, timeout=30):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(timeout)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        f = s.makefile("rb")
        return json.loads(f.readline())
cd = open(os.path.join(DIR, "cd.txt"), "a", encoding="utf-8")
t0 = time.time(); conectado = False; fallos = 0; n = 0
while n < 75:
    try:
        r = q({"cmd": "dump_ram", "addr": "0x80000000", "len": 2097152})
        seg = int(time.time() - t0)
        open(os.path.join(DIR, "ram_%04d.bin" % seg), "wb").write(bytes.fromhex(r["hex"]))
        h = q({"cmd": "cdrom_command_history"})
        cd.write("=== %d s\n%s\n" % (seg, json.dumps(h))); cd.flush()
        conectado = True; fallos = 0; n += 1
    except Exception as e:
        fallos += 1
        if (conectado and fallos >= 5) or (not conectado and time.time() - t0 > 900):
            cd.write("fin: %s\n" % e); cd.flush(); break
    time.sleep(2)
