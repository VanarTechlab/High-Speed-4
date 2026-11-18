# Muestra en directo los valores candidatos del menu de pausa.
import json, socket, time
ADDR = [("pausa 0x801E75B5", 0x801E75B5), ("cursor 0x80152A28", 0x80152A28),
        ("contador 0x80148C5D", 0x80148C5D)]
def rd(a):
    with socket.create_connection(("127.0.0.1", 4370), 2) as s:
        s.settimeout(5)
        s.sendall((json.dumps({"id": 1, "cmd": "read_ram", "addr": "0x%08X" % a, "len": 1}) + "\n").encode())
        r = json.loads(s.makefile("rb").readline())
        return int(r["hex"][:2], 16)
print("Esperando al juego... (deja esta ventana visible junto al juego)")
while True:
    try:
        vals = ["%s = %d" % (n, rd(a)) for n, a in ADDR]
        print("\r" + "   |   ".join(vals) + "      ", end="", flush=True)
    except Exception:
        print("\rEsperando al juego...                                              ", end="", flush=True)
    time.sleep(0.3)
