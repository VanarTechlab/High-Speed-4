import base64, glob, hashlib, json, sys, time, collections
t=time.time()
raw=open(sys.argv[1],'rb').read()
n=len(raw)//2352
ud=bytearray(n*2048)
for i in range(n):
    o=i*2352+24; ud[i*2048:(i+1)*2048]=raw[o:o+2048]
ud=bytes(ud)
idx=collections.defaultdict(list)
for s in range(n):
    idx[ud[s*2048:s*2048+16]].append(s)
print('indice', round(time.time()-t,1), flush=True)
fs=sorted(glob.glob(sys.argv[2]+'/*.json'))+[sys.argv[3]]
uniq={}
for f in fs:
    x=json.load(open(f)); x=x if isinstance(x,list) else [x]
    for c in x:
        b=base64.b64decode(c['bytes_b64'])
        h=hashlib.sha256(b).hexdigest()
        uniq.setdefault(h,(c['load_addr'],len(b),b))
print('capturas distintas', len(uniq), flush=True)
ok=0; ok_guard=0; nof=[]; tot=0; totok=0
for h,(la,L,b) in uniq.items():
    tot+=L
    found=False
    for s in idx.get(b[:16],[]):
        if ud[s*2048:s*2048+L]==b: found=True; break
    if not found:
        # sin los 4 bytes de guarda del final
        for s in idx.get(b[:16],[]):
            if ud[s*2048:s*2048+L-4]==b[:-4]: found='guard'; break
    if found is True: ok+=1; totok+=L
    elif found=='guard': ok_guard+=1; totok+=L
    else: nof.append((la,L))
print('encontradas enteras', ok, 'sin guarda', ok_guard, 'no encontradas', len(nof), 'bytes ok %.1f%%'%(100*totok/tot), flush=True)
print('ejemplos no encontradas', nof[:15])
