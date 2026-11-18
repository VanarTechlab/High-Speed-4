# Desensambla codigo MIPS desde una foto de RAM: disram.py foto.bin desde hasta
import sys, struct, importlib.util
spec = importlib.util.spec_from_file_location("d2", __file__.replace("disram.py", "dis2.py"))
src = open(__file__.replace("disram.py", "dis2.py")).read().split("lo=int(sys.argv[2]")[0]
exec(src)
ram = open(sys.argv[1], "rb").read()
lo = int(sys.argv[2], 16); hi = int(sys.argv[3], 16)
for a in range(lo, hi + 4, 4):
    i = struct.unpack_from("<I", ram, a - 0x80000000)[0]
    print("%08x: %08x  %s" % (a, i, dis(a, i)))
