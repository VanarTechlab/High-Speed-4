# Utilidades: inicio de funcion y llamadores (jal) en una foto de RAM
import struct,sys
d=open('fotos/c6.bin','rb').read()
W=lambda a: struct.unpack_from('<I',d,a-0x80000000)[0]
def start(a):
    while True:
        w=W(a)
        if (w>>16)==0x27BD and (w&0x8000): return a
        a-=4
def callers(t):
    w=0x0C000000|((t>>2)&0x3FFFFFF)
    return [0x80000000+i for i in range(0,len(d)-3,4) if struct.unpack_from('<I',d,i)[0]==w]
