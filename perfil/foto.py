# Foto de la RAM con nombre: foto.py NOMBRE -> perfil\fotos\NOMBRE.bin (el juego debe estar abierto con servidor de depuracion)
import json, os, socket, sys
DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "fotos"); os.makedirs(DIR, exist_ok=True)
name = sys.argv[1] if len(sys.argv) > 1 else "foto"
with socket.create_connection(("127.0.0.1", 4370), 5) as s:
    s.settimeout(30)
    s.sendall((json.dumps({"id": 1, "cmd": "dump_ram", "addr": "0x80000000", "len": 2097152}) + "\n").encode())
    r = json.loads(s.makefile("rb").readline())
open(os.path.join(DIR, name + ".bin"), "wb").write(bytes.fromhex(r["hex"]))
print("Guardada", name)
