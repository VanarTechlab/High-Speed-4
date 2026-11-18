# Traza de escrituras al estado de carrera (0x8013DA00-0x8013DC00) con la version build-perfil (debug tools).
# Fuerza 4 coches (2 humanos + 2 rivales) en menus como forzar.py y, 6 s despues de empezar la carrera,
# guarda quien escribio en perfil\traza_carrera.json
import json, os, socket, time
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "traza_carrera.json")
LOG = open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "traza.log"), "w", buffering=1)
def q(cmd):
    with socket.create_connection(("127.0.0.1", 4370), 5) as s:
        s.settimeout(60)
        s.sendall((json.dumps(dict(id=1, **cmd)) + "\n").encode())
        return json.loads(s.makefile("rb").readline())
def rd(a): return int(q({"cmd": "read_ram", "addr": "0x%08X" % a, "len": 1})["hex"], 16)
armado = False; inicio = None
while True:
    try:
        if not armado:
            r = q({"cmd": "wtrace_range", "lo": "0x8013DA00", "hi": "0x8013DC00"}); LOG.write("arm %s\n" % r); armado = True
        st = q({"cmd": "read_ram", "addr": "0x8010FB90", "len": 4})["hex"][:2]
        if st == "03":
            inicio = None
            for base in (0x80114868, 0x801187E8):
                if rd(base + 4) == 2:
                    q({"cmd": "write_ram", "addr": "0x%08X" % base, "val": "4"})
                    q({"cmd": "write_ram", "addr": "0x%08X" % (base + 8), "val": "2"})
        elif st == "02":
            if inicio is None: inicio = time.time(); LOG.write("carrera\n")
            elif time.time() - inicio > 6:
                # sin 0x8013DB1C (contador de frames que llena el registro)
                a = q({"cmd": "wtrace_dump", "addr_lo": "0x8013DAE0", "addr_hi": "0x8013DB1C", "count": 4000, "newest": 1})
                b = q({"cmd": "wtrace_dump", "addr_lo": "0x8013DB20", "addr_hi": "0x8013DB40", "count": 4000, "newest": 1})
                json.dump({"a": a, "b": b}, open(OUT, "w")); LOG.write("guardado\n"); break
    except Exception as e:
        LOG.write("err %s\n" % e); time.sleep(1)
    time.sleep(0.2)
