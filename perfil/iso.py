import struct,sys
f=open(sys.argv[1],'rb')
SS=2352; OFF=24  # mode2 form1: 24 bytes header (sync12+hdr4+subhdr8)
def sec(l):
    f.seek(l*SS); d=f.read(SS)
    mode=d[15]
    return d[16:16+2048] if mode==1 else d[24:24+2048]
pvd=sec(16); assert pvd[1:6]==b'CD001', pvd[:8]
root=pvd[156:156+34]
def rec(b,p):
    L=b[p]; lba=struct.unpack('<I',b[p+2:p+6])[0]; size=struct.unpack('<I',b[p+10:p+14])[0]
    fl=b[p+25]; nl=b[p+32]; name=b[p+33:p+33+nl]
    return L,lba,size,fl,name
def walk(lba,size,path,out):
    n=(size+2047)//2048
    for i in range(n):
        b=sec(lba+i); p=0
        while p<2048 and b[p]:
            L,l,s,fl,name=rec(b,p)
            if name not in (b'\x00',b'\x01'):
                nm=name.decode('latin1')
                if fl&2: walk(l,s,path+'/'+nm,out)
                else: out.append((l,s,path+'/'+nm))
            p+=L
out=[]
L,l,s,fl,nm=rec(root,0); walk(l,s,'',out)
for l,s,n in sorted(out): print(l, (s+2047)//2048, s, n)
