import struct,sys
f=open(sys.argv[1],'rb'); base=int(sys.argv[2])
def sec(l):
    f.seek(l*2352); d=f.read(2352); return d[16:2064] if d[15]==1 else d[24:2072]
h=b''.join(sec(base+i) for i in range(8))
print(h[:16])
magic=h[:4]; tot=struct.unpack('>I',h[4:8])[0]; n=struct.unpack('>I',h[8:12])[0]; hs=struct.unpack('>I',h[12:16])[0]
print(magic, 'total',tot,'n',n,'hdr',hs)
p=16
for i in range(n):
    off,sz=struct.unpack('>II',h[p:p+8]); p+=8
    e=h.index(b'\0',p); name=h[p:e].decode('latin1'); p=e+1
    print(i, name, 'off',off,'size',sz,'lba',base+off//2048, 'sec', sz//2048)
