import base64, glob, hashlib, json, struct, sys, collections
raw=open(sys.argv[1],'rb').read()
n=len(raw)//2352
def sec(i): o=i*2352+24; return raw[o:o+2048]
pvd=sec(16); root=pvd[156:156+34]
files=[]
def walk(lba,size,path):
    data=b''.join(sec(lba+k) for k in range((size+2047)//2048)); p=0
    while p<len(data):
        L=data[p]
        if L==0: p=((p//2048)+1)*2048; continue
        rec=data[p:p+L]; elba=struct.unpack('<I',rec[2:6])[0]; esz=struct.unpack('<I',rec[10:14])[0]
        flags=rec[25]; nl=rec[32]; name=rec[33:33+nl]
        if name not in (b'\x00',b'\x01'):
            nm=name.decode('latin1').split(';')[0]
            if flags&2: walk(elba,esz,path+nm+'/')
            else: files.append((elba,esz,path+nm))
        p+=L
walk(struct.unpack('<I',root[2:6])[0],struct.unpack('<I',root[10:14])[0],'/')
files.sort()
print(len(files),'archivos'); 
for f in files[:60]: print(f)
