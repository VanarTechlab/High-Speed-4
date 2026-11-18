import base64, glob, hashlib, json, sys, collections
raw=open(sys.argv[1],'rb').read()
n=len(raw)//2352
ud=bytearray(n*2048)
for i in range(n):
    o=i*2352+24; ud[i*2048:(i+1)*2048]=raw[o:o+2048]
ud=bytes(ud)
H={}
for s in range(n):
    H.setdefault(hashlib.md5(ud[s*2048:(s+1)*2048]).digest(), s)
fs=sorted(glob.glob(sys.argv[2]+'/*.json'))+[sys.argv[3]]
uniq={}
for f in fs:
    x=json.load(open(f)); x=x if isinstance(x,list) else [x]
    for c in x:
        b=base64.b64decode(c['bytes_b64']); uniq.setdefault(hashlib.sha256(b).hexdigest(),(c['load_addr'],b))
tot=0; okb=0; diffb=0; rep=[]
for h,(la,b) in uniq.items():
    # descartar si entera ya esta
    L=len(b)-4
    chunks=L//2048
    found=sum(1 for k in range(chunks) if hashlib.md5(b[k*2048:(k+1)*2048]).digest() in H)
    rep.append((la,len(b),chunks,found))
bad=[r for r in rep if r[3]<r[2]]
print('capturas con trozos que no estan en el disco:', len(bad), 'de', len(rep))
T=sum(r[2] for r in bad); F=sum(r[3] for r in bad)
print('en esas: trozos de 2 KB', T, 'encontrados', F, '(%.1f%%)'%(100*F/max(T,1)))
for r in sorted(bad, key=lambda r:-r[1])[:12]: print(r)
