# Perfil para el port a Vita: en que se va el tiempo del hilo del juego (PC).
# Uso: lo lanza perfil\25_perfil_vita.bat (E = normal, L = ligera). Espera a una carrera de 1 jugador
# (estado 0x8010FB90 = 1), deja 20 s de margen y mide 30 s: host_profile (funciones del .exe),
# phase_profile (interprete / nativo / estatico / gpu) y gcp_ring (funciones del juego PS1).
# Salida: perfil\vita_<modo>_*.json y perfil\vita_<modo>.log
import json, os, socket, sys, time
MODO = sys.argv[1] if len(sys.argv) > 1 else "E"
D = os.path.dirname(os.path.abspath(__file__))
LOG = open(os.path.join(D, "vita_%s.log" % MODO), "w", buffering=1)
def q(cmd, timeout=60):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(timeout)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
def guarda(nombre, obj):
    json.dump(obj, open(os.path.join(D, "vita_%s_%s.json" % (MODO, nombre)), "w"), indent=1)
def estado():
    return int(q({"cmd": "read_ram", "addr": "0x8010FB90", "len": 1})["hex"][:2], 16)
fase = "espera"
t0 = None
while True:
    try:
        if fase == "espera":
            if MODO == "L":
                try: LOG.write("hd off: %s\n" % q({"cmd": "hd_textures", "replacements": 0}))
                except Exception as e: LOG.write("hd err %s\n" % e)
            fase = "menus"
        st = estado()
        if fase == "menus":
            if st == 1:
                fase = "carrera"; t0 = time.time(); LOG.write("carrera 1J detectada\n")
        elif fase == "carrera":
            if st != 1:
                fase = "menus"; LOG.write("salio de carrera, espero otra\n"); continue
            if time.time() - t0 > 20:
                seq = q({"cmd": "host_profile", "top": 1}).get("seq", -1)
                LOG.write("inicio medicion seq %s\n" % seq)
                time.sleep(30)
                if estado() != 1:
                    fase = "menus"; LOG.write("salio durante la medicion, repetir\n"); continue
                guarda("host", q({"cmd": "host_profile", "since": seq, "top": 400}))
                guarda("phase", q({"cmd": "phase_profile", "window": 25}))
                g = q({"cmd": "gcp_ring"}); guarda("gcp_stats", g)
                try:
                    f = int(g.get("last_frame", g.get("frame", 0)))
                    if f > 0:
                        guarda("gcp_fn", q({"cmd": "gcp_ring", "frame_lo": max(0, f - 1500), "frame_hi": f, "key": "fn", "top": 60}))
                        guarda("gcp_native", q({"cmd": "gcp_ring", "frame_lo": max(0, f - 1500), "frame_hi": f, "key": "native", "top": 40}))
                except Exception as e:
                    LOG.write("gcp err %s\n" % e)
                LOG.write("LISTO: ya puedes cerrar el juego\n")
                print("\n*** PERFIL LISTO: ya puedes cerrar el juego ***\n")
                break
    except Exception as e:
        LOG.write("err %s\n" % e); time.sleep(1)
    time.sleep(0.5)
