import os, shutil, time
# Copia a hd\entrada_nuevas las texturas de dumps que aun no estan escaladas (sin las semitransparentes ST*). Lo ejecuta Claude.
t0=time.time()
d='../build-windows/mods/texture-packs/SLUS-00826/dumps'
have=set(os.listdir('salida_anime_todo')); done=set(os.listdir('entrada_nuevas'))
fs=[f for f in sorted(os.listdir(d)) if f.endswith('.png') and f not in have and not f.startswith('texupload-ST')]
pend=[f for f in fs if f not in done]
print('total nuevas',len(fs),'pendientes',len(pend),'listado %.0fs'%(time.time()-t0))
for f in pend:
    if time.time()-t0>150: break
    shutil.copyfile(os.path.join(d,f),'entrada_nuevas/'+f)
print('copiadas ahora; en carpeta', len(os.listdir('entrada_nuevas')))
