# NFS4 High Stakes (PS1, SLUS-00826) – PSXRecomp – estado al 2026-11-16

## Funciona (PC/Windows, MSYS2 MINGW64)
- Mando, analógico (lock analog), vibración, sonido, resolución interna 4x.
- Widescreen 16:9 / 21:9 (mod nfshs.widescreen, native_wide=false, gte_game_mode=true).
- Menús en 4:3 y pausa en panorámico: gameplay_state_addr 0x8010FB90 = 1.
- HUD: auto_ui_squash = false (con true movia letreros/paredes cercanos como si fueran HUD). Aguja del velocímetro recortada dentro del dial
  y sombra (LineF) oculta en widescreen (src/mods/nfshs_widescreen.c, override de 0x800D4A68).
- Dial y minimapa completos: parche a PSXRecomp gpu.c (el recorte E3/E4 sigue al HUD;
  líneas sueltas como piezas "enclosed"). Guardado en patches/psxrecomp_hud_widescreen.patch
  (original en analysis/gpu.c.orig_backup).
- Lados sin desaparecer y sin franjas: 64 [[recompiler.patch]] en game.toml que amplían
  el test de esquinas de 90° a x/4>z en 0x8007E71C (bloques), x/2 en 0x800C7288 (sub-bloques; x/4 daba poligonos estirados en Celtic Ruins),
  0x800C9DFC y 0x800CA618 (objetos). Generados con ~/work/genpatch.py (VM). Ya NO se usa cull.keep.
- 30 fps sin ralentizaciones (RESUELTO 2026-11-15): [timing] guest_cycle_scale = 2 +
  guest_cycle_scale_gate 0x8010FB90 = 1 (solo carrera) + guest_cycle_scale_batch = true, Y
  overlays compilados con native CPU spans (-DPSX_NATIVE_CPU_SPANS=1). El .exe ya tenia spans
  ON pero los overlays (donde esta la pista) no: hacian psx_cyc_step/icache/load-timing por
  instruccion. Cambios guardados en patches/psxrecomp_overlay_spans.patch (aplicar en psxrecomp/
  con git apply): compile_overlays.py (spans ON por defecto; OFF con overlay_spans.off junto al
  script o PSX_OVERLAY_SPANS=0; copiar tambien a build-windows\overlay_toolchain) y
  overlay_dispatch_preamble.c.inc (guard NFS4 en los forwarders psx_cyc_load_*/icache). Medido Test Drive Dolphin Cove: 0 underruns, 59.8-60.1 Hz,
  hilo ~720 ms/s (pico 867; antes 900-990 con 15-30 cortes).
  Medir: build-windows\prueba_30fps.bat (E = game.toml, H = lo mas ligero) -> prueba_E/H.log.
  Perfilar: build-perfil (PSX_DEBUG_TOOLS=ON) + perfil\1_overlays.bat y perfil\2_jugar.bat
  (host_profile + gcp_ring cada 10 s -> perfil\resultado.txt). Hot spot de la pista: 0x800C7288.
  OJO: el codigo INTERPRETADO no lleva los [[recompiler.patch]] -> cosas que desaparecen a los
  lados si faltan overlays por compilar (correr compilar_overlays.bat tras cada cambio).
- Mod nfshs.unlocks (pistas, trofeos, coches bonus, dinero).

## Compilar
    cd /d/Proyectos/NFS4/NFS-High-Stakes-Recomp && python psxrecomp/psxrecomp_cli.py rebuild --config game.toml --project-root . --build-dir build-windows --no-pgo
  y si cambió game.toml: doble clic en build-windows\compilar_overlays.bat (esperar "Terminado").
  Los overlays grandes no se compilan solos a tiempo en segundo plano.
- Si se toca algo de psxrecomp que entra en el hash de codegen (p.ej. runtime/include/
  overlay_dispatch_preamble.c.inc; lista en runtime/codegen_hash_sources.cmake), hay que
  recompilar el recompilador y copiarlo, si no compilar_overlays.bat da "STALE RECOMPILER":
    cmake --build build-recompiler --target psxrecomp-game && cp build-recompiler/psxrecomp-game.exe build-windows/overlay_toolchain/
- Debug: build-debug (puerto 4370), lanzar con build-debug\jugar_debug.bat.
- Log del juego: build-windows\jugar_con_log.bat -> run_windows.log.

## Pendiente (siguiente)
0. PLAN PC (2026-11-16), en este orden:
   a) Jukebox HECHO 2026-11-16 (pausa: titulo + L1/R1, nombre y numero del cuadro OK): mod nfshs.jukebox (src/mods/nfshs_jukebox.c +
      mods/preloaded/packages/nfshs.jukebox). Musica = ZZZZYMUS.VIV (BIGF, LBA 126428): 18 temas
      gameNN + lose + win. En carrera la cabecera del VIV esta en RAM (0x801D8B78); el juego
      elige entrada al azar al empezar carrera, al acabar un tema y al quitar la pausa. El mod
      hace que las 18 entradas apunten al tema elegido (opcion "track", cambiable en vivo).
      Herramientas: perfil\4_musica.bat + perfil\musica2.py (historial CD). perfil\iso.py y
      perfil\viv.py listan el disco y el VIV.
      Intento 1 (escribir la cabecera VIV en RAM cada vblank) FALLO: no cambia el tema y
      entrecorta el sonido. Intento 2 FUNCIONA (2026-11-16, desde el launcher): hook en cdrom.c (patches/psxrecomp_jukebox_cdrom.patch,
      psx_cdrom_set_user_remap): el mod sirve los bytes del tema elegido cuando el juego lee el suyo.
      Valores del menu en directo: perfil\6_menu_vivo.bat (ventana con pausa/cursor/contador).
      MENU DE PAUSA (lo quiere ahi): textos en RAM 0x8015261C ("Game Paused", Continue, Restart,
      Options, Quit Race/Forfeit Race; Options > Audio Settings: Mode, Music, Effects, Speech,
      Engine, Ambient). Candidatos (perfil\5_menu.bat, sin confirmar): 0x801E75B5 = en pausa (1),
      0x80152A28 = cursor del menu de pausa (0 Continue, 1 Restart, 2 Options, 3...),
      0x80148C5D = contador. MEDIDO 2026-11-16: 0x80118234 != 0 (vale 5) = menu de pausa abierto;
      0x80152A28 = cursor 1-4 del menu principal (no cambia en submenus). Version rapida (EN PRUEBA):
      en pausa el titulo "Game Paused" muestra "Tema NN/18"/"Aleatorio" y L1/R1 cambian el tema
      (pad transform). MEDIDO: titulo "Game Paused\0Continue" cambia de sitio (se busca en
      0x80140000-0x80170000); el juego redibuja el menu al pulsar izquierda/derecha sin mover
      el cursor -> tras L1/R1 el mod pulsa derecha 4 fotogramas (no tocar el cursor a mano: desincroniza). Cabecera VIV en RAM cambia de sitio (se busca
      "BIGF"): el mod escribe el nombre elegido en los 18 huecos (cuadro "tema que suena").
      Ficha del reproductor (direccion variable, se busca por el puntero +0x38 al nombre):
      +0x18 = 18, +0x1C = indice (0-17) del tema elegido por el juego (diagnostico PSX_JUKEBOX_DIAG=1,
      build-windows\jugar_jukebox_diag.bat): el mod lo cambia al elegido para el numero del cuadro.
   b) Parche HD: HECHO (2026-11-17, "va perfecto"). Estado final:
      - Pack: hd\pack_anime (39.790 texturas, modelo realesrgan-x4plus-anime; x4 si hay detalle, x2 el resto;
        sin semitransparentes ST*; sin las de malas.txt). Instalado en build-windows\mods\texture-packs\
        SLUS-00826\replacements. Reinstalar entero: hd\9_instalar_pack_anime.bat.
      - Escaladas completas en hd\salida_anime_todo (sirven para saber que ya esta hecho).
      - Ampliar: dump ON y jugar -> Claude: hd\preparar_nuevas.py -> hd\10_escalar_nuevas.bat -> Claude:
        HD_SOLO_NUEVAS=1 verificar.py + preparar_pack.py -> hd\11_instalar_nuevas.bat.
      - Motor: patches/psxrecomp_hd_vram_read.patch y psxrecomp_hd_caches.patch; render_thread = true.
      - Limpieza: borradas pruebas x4plus/x2/mixta/pack_final y scripts de pruebas (2026-11-17).
      Historial de la investigacion:
      - Piloto: 1.116 texturas x4 (Real-ESRGAN, carpeta hd\) en mods\texture-packs\SLUS-00826\replacements.
      - MEDIDO (prueba_J con HD vs prueba_E sin HD, Dolphin Cove): con HD el hilo del juego va al 99%
        (work guest ~990 ms/s, 52-58 Hz, underruns cada 2 s); sin HD ~650-780 ms/s, 60 Hz, 0 underruns.
        El precarga va bien (308 MiB en 340 ms): no es la cache.
      - Causa (codigo): con HD el runtime entra en "native authority" y ademas de OpenGL dibuja cada
        poligono por software en VRAM nativa (gpu_gl_renderer.c, glb_draw_*), para poder reconocer texturas.
      - MEDIDO prueba R (PSX_RENDER_THREAD=1): casi siempre 60 Hz y 0 cortes; pero en tramos (salida y
        antes del tunel de Dolphin Cove) el hilo de dibujo se para. perfil\9_hilo.bat + hilo.py
        (comando debug "render_thread") dio el motivo: "vram_read" ~1 vez por frame (el juego lee VRAM),
        que toma el contexto el resto del frame y devuelve todo el dibujo al hilo del juego.
      - Arreglo: patches/psxrecomp_hd_vram_read.patch (rtb_vram_read: con HD y hilo de dibujo, solo vacia
        la cola y lee la copia privada, sin tomar el contexto).
      - MEDIDO tras el arreglo (9_hilo.bat, Dolphin Cove con tunel): 0 underruns, 59.8-60 Hz, hilo del juego
        max ~710 ms/s, ya no aparece "vram_read" como motivo de parada. El usuario confirma: sin cortes.
      - Pendiente: dejar el hilo de dibujo activado por defecto (ahora solo con PSX_RENDER_THREAD=1) y
        decidir x4 o x2 para el pack completo (x4: 1.116 imagenes = 308 MiB; las 21.986 no caben en la precarga de 512 MiB).
      - Decidido pack MIXTO (el usuario vio x4 mejor solo en casas): hd\detalle.py mide cuanto detalle se
        pierde de x4 a x2; >= 3.0 -> x4 (carteles, fachadas, velocimetro), resto -> x2.
      - Escalado completo: hd\5_escalar_todo.bat (Real-ESRGAN x4, ~12 min en el PC del usuario; salieron
        avisos vkQueueSubmit -4 pero las 21.986 salieron bien). hd\verificar.py: 172 donde la IA invento
        otro dibujo (rejillas/cuadros) -> malas.txt, se excluyen (el juego usa la original).
      - hd\preparar_pack.py -> hd\pack_final: 21.814 texturas (7.697 x4 + 14.117 x2), 685 MB.
        Instalar con hd\6_instalar_pack.bat (vacia replacements y copia).
      - MEDIDO 1a prueba del pack completo: 5-7 fps en carrera (hilo de dibujo al 98%). Contadores (hilo.py
        consulta "hd_textures"): ~2.400 "composiciones" reconstruidas por segundo y otras tantas subidas+
        expulsiones de la cache GL (cache de composiciones 256/64 MiB y cache GL 4096 huecos, llenas).
        Arreglo: patches/psxrecomp_hd_caches.patch (composiciones 4096/512 MiB con hash; cache GL 8192 huecos).
      - Calidad: el sol, una elipse negra y efectos salian opacos: son texturas STP (semitransparentes,
        alfa 143 en el dump) y el umbral de alfa de x2 las volvio opacas. Las 656 STP se sacaron del pack
        (hd\pack_excluidas); el juego usa las originales. pack_final = 21.158.
      - MEDIDO 2a prueba (caches grandes, sin STP): se ve bien, ~60 fps, 5 underruns en ~4 min (salida y un
        tramo). Siguen ~50 composiciones nuevas/s todo el rato: la composicion dependia del rango UV de
        cada poligono. Arreglo (en psxrecomp_hd_caches.patch): acquire_draw ensancha la consulta al
        rectangulo completo de las texturas HD encontradas -> una composicion sirve a todos los poligonos.
      - MEDIDO 3a prueba: fluido, 2 underruns en ~4 min (solo a la salida); composiciones en 214 s de carrera:
        218 (antes ~9.000). Texturas se ven bien. FUNCIONA.
      - Cobertura: el pack tiene lo que salio en los dumps (una vuelta por pista, 3.607 texturas base);
        variantes no vistas (noche, clima, sentido inverso, otros coches) quedan originales.
      - Hilo de dibujo por defecto: [video] render_thread = true en game.toml (ambos) y en
        build-windows\settings.toml (lo escribe el launcher y manda sobre game.toml; opcion "render thread").
        Requiere compilar_overlays (cambia el hash de game.toml).
      - Celtic Ruins: tramas/rejillas en el asfalto = la IA invento trama fina en superficies lisas (y esas
        salian como "mucho detalle" -> x4). hd\inventado.py: ratio bordes IA / bordes original ampliado;
        >= 7 = inventada. 666 excluidas (hd\inventadas.txt -> pack_excluidas). pack_final = 20.492.
      - Pantalla de titulo: con DumpC16Textures: true solo salieron tiras 16x127 STC16 de los mapas de pista
        animados del menu; la pantalla de titulo NO se graba -> se copia directa a VRAM (vram-write, no
        soportado por el motor). Se deja como esta.
      - Tras quitar las 666: el usuario ve el asfalto/cielo de Celtic Ruins solo algo mejor. Ningun criterio
        automatico (ratio de bordes, estructura del original) separa bien trama inventada de detalle real:
        el problema es el modelo realesrgan-x4plus (fotos). Probando realesrgan-x4plus-anime y
        realesr-animevideov3 x2 con 28 muestras (hd\7_probar_modelos.bat, hd\entrada_modelos).
      - MEDIDO (analysis\comparacion_modelos.png): anime x4 = sin trama inventada en asfalto/paredes, carteles
        igual de nitidos; animevideo x2 = salida rota (basura de colores). Elegido realesrgan-x4plus-anime.
        Reescalar todo: hd\8_escalar_todo_anime.bat -> hd\salida_anime_todo; luego verificar/preparar_pack con
        HD_SALIDA=salida_anime_todo HD_PACK=pack_anime (excluir texupload-ST*).
      - Hecho: 22.276 escaladas con anime; verificar: 159 con color distinto (excluidas); sin ST*.
        hd\pack_anime = 21.192 texturas (4.741 x4 + 16.451 x2), 416 MB. Instalar: hd\9_instalar_pack_anime.bat.
        PROBADO: se ve mejor que x4plus (usuario). Pack en uso.
    - Error grafico Celtic Ruins (~3:16 de carrera): poligono plano grande de ladera tapando el cielo.
      MEDIDO: sale igual sin HD y en 4:3 -> no es el HD ni el aspecto. SUPUESTO: los 64 parches de culling
      lateral x/4 (siempre compilados, tambien en 4:3) dejan pasar una pieza casi detras de la camara.
      PRUEBA en curso: game.toml sin parches (copias con parches en analysis\game.toml.con_fov y
      analysis\build-windows_game.toml.con_fov; restaurar con perfil\restaurar_fov.bat).
      MEDIDO: sin ningun parche la mancha DESAPARECE -> son los parches de culling lateral.
      PRUEBA B (sin blk): la mancha SIGUE -> no son los bloques.
      PRUEBA C (sin sub): mancha DESAPARECE pero vuelven cosas que desaparecen a los lados -> culpables los
      parches de SUB-BLOQUES (0x800C7450-0x800C7500). Con x/4 (76 grados por lado) entran piezas casi de
      lado/detras cuyos vertices se proyectan mal.
      PRUEBA D: sub-bloques a x/2 (63 grados; sra ...,1 = 0x00031843 / 0x00021043), resto x/4.
      MEDIDO (usuario): sin manchas; suelo y paredes no desaparecen; solo algunos arboles desaparecen un
      poco antes en el borde. QUEDA ASI (copia de referencia: analysis\game.toml.con_fov). Un angulo intermedio (x*3/8) no cabe en los
      huecos de 1 instruccion del parche (haria falta una "code cave").
  Anadir texturas (noche, lluvia, nieve...): dump ON en el launcher y jugar esas variantes ->
      hd\10_escalar_nuevas.bat (solo las nuevas, modelo anime) -> Claude: HD_SOLO_NUEVAS=1 verificar.py y
      preparar_pack.py (HD_SALIDA=salida_anime_todo HD_PACK=pack_anime) -> hd\11_instalar_nuevas.bat.
      2026-11-17: 2a ronda (noche/lluvia/nieve...): 18.691 nuevas (sin ST*), 93 con color malo excluidas,
      18.598 anadidas a pack_anime (x4 ~3.100, x2 ~15.500). Copia a entrada_nuevas: hd\preparar_nuevas.py.
  FPS reales (MEDIDO 2026-11-17 con perfil\11_fg.bat + fg.py, comando "frame_gen", campo flips = cambios de
      imagen mostrada por segundo): menus ~30/s; CARRERA ~56/s (casi 60: guest_cycle_scale 2 deja al juego
      terminar cada imagen en un solo VBlank). Smooth motion (frame generation, PSX_FRAME_GEN=1, opcion G de
      prueba_30fps.bat): en menus genera 30+30; en carrera casi nada (ya hay imagen nueva cada VBlank) y
      rechaza mucho ("few inliers", "camera passes through geometry"). No compensa -> desactivado.
  Resolucion interna: el launcher solo mostraba Game default y 1x..32x aunque el motor acepta 720p, 1080p,
      1440p, 4k, 5k, 8k, display (Match display) y auto (runtime/include/internal_resolution.h). Anadidas
      esas opciones al manifest builtin psx.enhancement.internal-resolution -> patches/psxrecomp_ir_presets.patch
      (se copia a build-windows\mods\bundled al compilar).
  PENDIENTE (para despues): mas detalle en coches lejanos. SUPUESTO: el juego cambia a modelos LOD de menos
      poligonos con la distancia. Plan: 1) medir poligonos de un coche rival cerca/lejos; 2) encontrar la
      comparacion de distancia que elige el modelo; 3) parche [[recompiler.patch]] para usar el detallado mas
      lejos. Riesgo: rendimiento (CPU emulada) y que el modelo detallado no este cargado para todos.
  c) PGXP (poligonos temblones): mod "PGXP Precision" (builtin psx.enhancement.pgxp), desactivado.
      Prueba con build normal (perfil\10_pgxp.bat + pgxp.py, comando "geom_correction"):
      MEDIDO ~4,5 min: triangulos precise 56.636 (0,3%), mixed 2.016.885 (11%), native 15.586.900 (88%);
      vertices dataflow_hit 2,4M + fallback 0,25M frente a native 50M (value_mismatch 24M, tolerance 8M).
      Usuario: lo quieto mejora un poco, coches lejanos en movimiento se deforman/tiritan (= triangulos
      "mixed": unas esquinas corregidas y otras no). Con build normal no sirve -> desactivar.
      Siguiente opcion: build "hook" (psxrecomp_add_game_runtime(... PGXP) en CMakeLists.txt, -DPSX_PGXP=1,
      overlays flavor 2): mucha mas cobertura segun docs (ENHANCEMENTS.md G1.11), coste de rendimiento sin medir.
      - Pendiente opcional: script para anadir texturas nuevas al pack sin reescalar todo.
   c) 4 jugadores local pantalla partida (el juego original solo trae 2: requiere cambiar
      la logica del juego; muy dificil).
      MEDIDO 2026-11-17 (prueba_30fps E): 2 jugadores Single Race ~40% de un nucleo (max 47%, 0 cortes);
      1 jugador con 5 rivales ~70% (picos 99% ~12 s en Landstrasse, 4 cortes -> revisar aparte).
      Modos 2J: Test Drive, Single Race, Hot Pursuit, High Stakes; sin rivales IA, pero con trafico y
      (Hot Pursuit) policia. 1J: max 6 coches (5 rivales).
      Plan elegido: partir de 2 jugadores (camino B) y convertir coches de trafico/policia en J3/J4;
      vistas 3-4 con render passes (psx_mod_render_pass). Paso 1: mapa de coches en RAM
      (perfil\12_coches.bat + coches.py: RAM cada 1 s, 60 s, J1 conduce y J2 quieto).
      MEDIDO (captura 2J Dolphin Cove): estado 0x8010FB90 = 2 en carrera de 2 jugadores (1 en 1J, 3 menus)
        -> en 2J NO se aplica guest_cycle_scale ni la deteccion de carrera del widescreen.
      Coches: array de structs de 0x8F0 bytes; J1 en 0x801C4500.. (pos x,y,z 0x801C45C8, velocidad +0xC,
        4 esquinas 0x801C4538.., matriz 0x801C4568..), J2 = J1 + 0x8F0 (pos 0x801C4EB8). Solo 2 structs en 2J;
        el siguiente hueco es basura. Trafico no encontrado en esa captura.
      Idea siguiente: buscar en el menu el numero de rivales y forzar rivales IA en 2J (array con 4 coches)
        y luego darles los mandos 3/4. Herramientas: perfil\13_juego_debug.bat + 14_foto.bat (foto.py).
      MEDIDO: 0x8013D5F8 NO es el numero de coches (cambia solo varias veces por segundo; coincidio por
        casualidad). Con fotos repetidas (r0a/b, r1a/b, r5a/b) el ajuste estable del menu es 0x80181D20 =
        0 / 1 / 2 (opcion sin rivales / 1 rival / maximo). El numero de coches se calcula al empezar la carrera.
      MEDIDO (fotos en carrera c1/c2/c6/j2): configuracion de carrera en 0x80114868 (copia 0x801187E8):
        +0 total coches (1/2/6, 2J=2), +4 humanos (1, 2J=2), +8 rivales IA (0/1/5, 2J=0).
        Estado de la carrera en 0x8013DAE8: +0 total, +4 total, +8 rivales, +0xC humanos, +0x10 rivales,
        +0x1C humanos, +0x40 humanos.
      PRUEBA 1 (forzar config 4/2/2 en menus, perfil\15_forzar_coches.bat): SALEN 4 COCHES en 2J, pero
        "Race Summary" nada mas empezar y los rivales sin IA (ruedan hacia atras).
        Estado de carrera (foto f4a) incoherente: [4, 2, 0, 2, 0 ...] y +0x40 = 4 (en 1J con rivales seria
        [4, 4, 2, 2, 2] y +0x40 = humanos).
      PRUEBA 2 (corregir estado los 5 primeros s, perfil\16_forzar_corregir.bat): sigue el Race Summary al
        empezar; solo se veian 3 coches separados. +0x40 NO es "humanos" (cambia solo: 0/3/4).
      Siguiente: traza de escrituras (wtrace, solo build-perfil con PSX_DEBUG_TOOLS=ON) a 0x8013DAE0-DB40 con
        4 coches forzados -> perfil\17_traza.bat + traza.py -> perfil\traza_carrera.json (PC/RA de quien escribe).
      MEDIDO (traza + desensamblado desde foto, perfil\disram.py/dis2.py): gp = 0x8013D83C en la carrera;
        gp+684 total, +688 corredores, +692/+700 rivales IA, +696 humanos, +712 otra lista.
        0x80086A78 limpia contadores; 0x80086B08 = registrar coche (a0 = coche): flags = ficha_piloto[+4];
        bit0 = humano (listas 0x80110D0C, corredores 0x80110CC4, flags coche+0x260 |= 0x105),
        bit1 = rival IA (lista 0x80110CE8 ...). Lista de coches 0x80110CA0; coche+0x288 -> ficha de piloto.
        Fichas de piloto: 0x80114878 + slot*0xB4 (+4 tipo 1 humano/2 IA, +8 activo, nombre en +0x5C).
        En 2J forzado las fichas 2 y 3 quedan a 0 -> coches sin tipo (ni humano ni IA).
      PRUEBA 3: copiar fichas 2 y 3 de la foto c6 (tipo IA) en menus (perfil\18_forzar_ia.bat + forzar_ia.py).
        RESULTADO: FUNCIONA. 2 jugadores + 2 rivales IA que corren solos, sin Race Summary; el HUD de posicion
        sigue marcando x/2 (cuenta solo humanos o es el diseno de 2J).
      PRUEBA 4: fichas 2 y 3 con tipo 1 (humano) -> perfil\19_forzar_humanos.bat (forzar_ia.py humano).
        RESULTADO: carrera con 4 "humanos" (corredores 4, rivales 0), el HUD pasa a x/4; uno de los extra
        imita el mando del J1. Ficha de piloto +0x54 = numero de jugador (2J: 1 y 2; copiadas de c6 eran 2 y 3).
      PRUEBA 5: igual con +0x54 = 3 y 4 -> un coche extra acelera solo, el otro quieto.
      MEDIDO (codigo): mandos en 0x8013FBA4: +4 + 8*indice (8 bytes por mando, indices 0-7: 0-3 puerto 1,
        4-7 puerto 2, formato multitap). 0x800E4E08 copia cada frame los buffers de libpad (0x8013FBF8, 34 B)
        y rellena con 0xFF lo que no es el primer mando del puerto. Lectura: 0x800E4DA4 get_buttons(indice),
        0x800BE494 pulsado(indice, mascara), 0x800BE42C (a0 = indice<<16 | mascara).
      Siguiente: traza de llamadas (fntrace, build-perfil) para ver que indice lee cada jugador
        (perfil\20_traza_mandos.bat + traza_mandos.py -> traza_mandos.json).
      MEDIDO (fntrace): en carrera solo se leen los indices de mando 0 y 4. La funcion de entrada por jugador
        0x8009CF3C recorre SOLO 2 jugadores (slti fp,2). Asignacion de controles por jugador en 0x80114ECC +
        0x130*j (cada entrada: dispositivo/boton, a0>>8 = indice de mando; tabla de lectores 0x8011FF08).
        Salida por jugador: 4 bytes en 0x8013E4FC + 4*j = [direccion (der-izq)/2, acelerador, freno, ?];
        0x8013E510 + 4*j = hay entrada. Los jugadores 3 y 4 no se procesan.
      PRUEBA 6: escribir controles fijos en 0x8013E504/E508 (jugadores 3/4) durante la carrera
        (perfil\21_prueba_controles.bat = forzar_ia.py controles).
        RESULTADO: sin efecto (los coches 3/4 no leen ahi; 0x8013E504/E508 son de otras cosas).
      MEDIDO (codigo): 0x8009DCE0 aplica los controles: si humanos >= 2 llama 0x800B80A0(0, &E4FC) y
        0x800B80A0(1, &E500); con 1 humano usa el indice de coche en 0x801144A4+0x1C. Nadie lo hace para 2 y 3.
      PLAN mod nfshs.4p: hook de entrada en 0x8009DCE0 -> psx_mod_call_guest(0x800B80A0, 2, buf3) y (3, buf4),
        con buf3/buf4 (4 bytes [dir, acel, freno, ?]) en RAM libre (p. ej. ficha de piloto 5 sin usar) y
        rellenos desde los mandos 3/4 del PC (requiere MAX_PLAYERS 4 / players = 4; ojo con el multitap del
        motor, que cambia el formato del puerto 1 y el juego no lo entiende).
      MEDIDO (codigo): 0x800B80A0(j, ctrl) mete los controles en un anillo 0x8013F3FC (+0x80*j, 32 ranuras,
        validez en +0x100, contadores +0x200/+0x204+4j: sitio solo para 2 jugadores) y 0x800B81B8(ranura)
        los copia a 0x8013F0D0 + 4*j para los humanos de la config (0x8011486C).
      HECHO: mod nfshs.4p (src/mods/nfshs_4p.c, mods/preloaded/packages/nfshs.4p): en menus convierte 2J en
        4 coches (fichas 2-3 humanas "Player3/4" o rivales IA, opcion "extra"); override de 0x800B81B8 que
        copia J1/J2 como el original y escribe J3/J4 en 0x8013F0D8/DC desde mandos SDL (opciones pad3/pad4)
        o teclado (J3 numerico 8/5/4/6, J4 G/B/F/H). Sin multitap (lee SDL3 directamente).
      PRUEBAS: escribir en 0x8013F0D8/DC CONGELA la presentacion (es otra estructura; 0x8013F0D0 solo tiene
        sitio para 2). Sin escribir nada no se congela.
      MEDIDO (codigo): cada coche humano llama 0x800B5D04 -> j = coche+0x258 (indice de humano) ->
        0x800B54F8(j) / 0x800B580C(j) -> 0x8009DD58(j) copia 0x8013F0D0+4j a 0x8013E504 y lo pasa a un
        historial por jugador en 0x80118324 + 0x80*j (+0 dir/4+64, +32 acel>>3, +64 freno>>3, +96 botones).
        Para j = 2/3 leia basura de 0x8013F0D8 (de ahi el coche que acelera solo).
      CAMBIO: el override ahora es 0x8009DD58: para j = 2/3 escribe los controles de J3/J4 en 0x8013E504.
        Riesgo: el historial de j = 2/3 cae en 0x80118424.. (estructura del juego); en la prueba A no rompio.
      RESULTADO: J4 (G/B/F/H) controla su coche, sin congelarse. J3 sin respuesta: el portatil no tiene teclado
        numerico -> anadidas teclas 8 acel / 0 freno / 7-9 girar.
      RESULTADO (2026-11-17): J3 (8/0/7/9) y J4 (G/B/F/H) controlan sus coches. 4 jugadores conduciendo.
      RESULTADO mandos: con 2 mandos iguales, mando 1 -> J1 y mando 2 -> J3 (opcion pad3 = Mando 2). OK.
      MEDIDO rendimiento 2J + 4 coches (prueba_30fps E, vuelta a Dolphin Cove): hilo del juego ~39% (max 44%),
        0 cortes -> hay margen para dibujar 2 vistas mas.
      Camara: no hay coordenadas de camara persistentes en RAM que sigan al coche (se calcula cada frame).
        Idea: las vistas usan la lista de humanos 0x80110D0C (coche por vista); dibujar otra vez con la lista
        cambiada a los coches 3/4 y en la otra mitad de la pantalla.
      MEDIDO vistas (perfil\22_vistas.bat en build-perfil, gpu_frame_dump + fotos): en 2J el juego va a 60
        imagenes/s con doble buffer en VRAM x=0 / x=320, y=256..495. Lista de "vistas" (capas de dibujo) en
        0x8011FF44, 200 bytes cada una, numero en gp+4696 (0x8013EA94): +0 tamano de OT (2000 en 3D, 1 en HUD),
        +8 DRAWENV buffer 0, +0x64 DRAWENV buffer 1 (92 bytes), +0xC0/+0xC4 OT de cada buffer.
        2J = 10 vistas: 0/1 escena J1 (y 256..375) / J2 (376..495), 2/3 recuadros 45x48 (minimapa?), 4/5 HUD,
        6/7 recuadros 28x28, 8/9 ?. 1J = 6 vistas (escena 320x240). 0x800BEB84(v, buf) = DRAWENV de la vista.
      PRUEBA vistas: config con 4 humanos (no 2) -> perfil\23_forzar_4humanos.bat (forzar_ia.py hum4).
      PENDIENTE 4P: vistas 3 y 4
        (pantalla en 4: render pass con la camara de los coches 3/4 + composicion en cuartos); marcador de
        posicion/HUD de J3/J4; rendimiento (en 2J no se aplica guest_cycle_scale: estado 2 no esta en el gate);
        quitar scripts de prueba de perfil\ cuando ya no hagan falta.
   d) En linea hasta 8 (PSXRecomp trae netplay con rollback para los 2 jugadores del juego;
      mas de 2 depende de c) y es aun mas dificil).
1. Prototipo PS Vita (VitaSDK, 16:9 fijo, 1 jugador, sin netplay/launcher/tcc/debug).
   Viabilidad medida 2026-11-16 (prueba_30fps.bat opcion H: escala 1, spans, res nativa,
   Dolphin Cove): hilo del juego ~450-540 ms/s (emu_busy ~50%, max 56%) de UN nucleo del
   portatil. PSXRecomp no tiene soporte Vita (ni ARMv7, ni GLES2; overlays = DLL compiladas
   en marcha con clang). Estimacion: la Vita (A9 333-444 MHz) quedaria varias veces corta.
   Decisiones (2026-11-17): CPU a 444 MHz (maximo oficial; el homebrew puede pedirlo; 500 MHz solo con
   plugin LOLIcon/PSVshell) si hace falta. Resolucion recomendada 854x480 (2x exacto, panoramico;
   pantalla Vita 960x544). La resolucion no resuelve la CPU: el trabajo principal sera abaratar la CPU emulada.
2. Opcional: texturas HD, modelos .geo, revisar brillo de luces de salida.
- MEDIDO prueba config 4 humanos (23_forzar_4humanos): se CONGELA. Descartado; se hace a mano.
- MEDIDO codigo vistas: 0x800B3EAC crea vistas al empezar (0x800B3D74: si 0x801144B0==1 -> 2 escenas 320x120 en y=256/376, OT 2000;
  0x800BE8A8(x0,y0,x1,y1,[w,h,dfe,isbg,ot]) crea una). Lista de vistas: sitio para 10 (0x80120714 es otra estructura); 2P usa las 10:
  0/1 escenas, 2/3 45x48, 4/5 HUD, 6/7 28x28, 8/9 pantalla completa. OT se reservan en el 1er cuadro (0x800BEFB4 -> 0x800BEC4C, pool prims gp+4720 = 0x228E8 en 2P).
- MEDIDO cuadro de carrera 0x800B4170: si pantalla partida llama 0x800B4024(vista gp+3700, 0) y 0x800B4024(gp+3704, 1).
  0x800B4024: vista actual 0x80118234 (+4 = j) -> 0x800B3EE4 (camara 0x80081128 -> 0x80085A2C(j); camaras 0x80110570 + 0x110*j, +0 = coche; solo 2 ranuras) -> 0x800B3F80 escena.
- HECHO (sin probar) nfshs_4p.c: override 0x800B3EAC (cuadrantes 160x120; vistas 6/7 -> escenas J3/J4; HUD 4/5 a TL/TR) y override 0x800B4024 (tras J2 dibuja J3/J4
  en vistas 6/7 intercambiando la ranura de camara 1 y la entrada 1 de la lista de humanos). NFSHS_4P_PANTALLA=2 = desactivar.
- MEDIDO 1a prueba pantalla 4: funciona, J1/J2 bien; J3/J4 cielo+faro iguales. Registro CAMS: camara +0 y +4 = coche seguido, +0x70/+0x74 modo
  (carrera 0/0x08003284; intro 0x0E/0xA0); la copia de J3/J4 se quedo en modo intro. Arreglo: recopiar de la ranura 1 cuando cambie el modo, +0/+4 = coche.
- MEDIDO 2a prueba: J3/J4 = copia quieta de la camara de J2 (dibujar no mueve camaras) y cierre (salto a 0x00807166 desde 0x800B721C).
  Logica de camaras 0x80083DFC: recorre camaras 0..(bit7 de cam0+0x74), coche = cam+0. Mod: override que tras el original la repite con la camara
  propia de J3/J4 en la ranura 1 (ranura 0 guardada/restaurada). Recopia de J2 solo al cambiar su modo.

## Historial 4P pantalla partida (2026-11-17, detalle de pruebas)
- src/mods/nfshs_4p.c tiene 5 overrides/plugins: fuerza 4 coches (vblank), controles J3/J4 (0x8009DD58),
  vistas en cuadrantes (0x800B3EAC), dibujo extra J3/J4 (0x800B4024), logica de camara extra (0x80083DFC). Diagnostico CAMS en el registro (cuadros 120/600).
- PENDIENTE PROBAR la ultima version (camara extra). Prueba: compilar, build-windows\prueba_4p_teclado.bat, 2 jugadores > Single Race > Dolphin Cove,
  mover J3 (8/7/9) y J4 (G/F/H). Registro: build-windows\run_4p_teclado.log.
- Despues: aguja (tacometro 28x28 de las vistas 6/7 viejas) en J3/J4; HUD J3/J4; vista propia por jugador (ahora J3/J4 copian el modo de J2);
  imagen recortada (cuadrante 160 de ancho: aplastar en horizontal); memoria de primitivas (gp+4720 = 0x228E8) si faltan trozos; opcion en el launcher.
- MEDIDO 3a prueba (camara extra): J3/J4 siguen a su coche (CAMS cuadro 600: posiciones +8..+10 distintas en CAM2/CAM3).
  Un momento hubo 2 vistas repetidas (SUPUESTO: recopia de J2 al cambiar su modo). Cierre en cuadro 6609 (~110 s):
  psx_crash.txt = FAIL-FAST unknown dispatch addr=0x0A000000 ra=0x800EDF00 (a0=0x80138B58 tabla, a1=0xC7AB).
  Distinto del cierre anterior (0x00807166 desde 0x800B721C) -> SUPUESTO: memoria pisada (pool de primitivas
  gp+4720 = 0x228E8 pensado para 2 escenas, o historial j=2/3 en 0x80118424). Prueba A/B: build-windows\prueba_4p_sin_pantalla4.bat
  (NFSHS_4P_PANTALLA=2, log run_4p_sin_pantalla4.log): si NO se cierra en ~3 min -> culpa de las vistas extra.
- MEDIDO 4a prueba (otra vez prueba_4p_teclado, con pantalla en 4): cierre en cuadro 11105 (~3 min): salto a 0x00806F60 desde
  0x800702B4 (jalr de una tabla virtual: objeto 0x8010EA70 = {coche J1, tabla 0x800555EC}). En foto RAM normal (perfil\coches\ram_0009.bin)
  la tabla tiene +8 = 0 y +12 = 0x8005BC78; en el cierre +8 = ..7F6F y +12 = 0x00806F60 -> MEDIDO: algo PISA memoria (0x800555F4..F8).
  Los 3 cierres son en sitios distintos; 2 saltan a valores tipo coordenada (0x0080xxxx), el otro a 0x0A000000. Falta la prueba A/B (prueba_4p_sin_pantalla4.bat, 5 min).
  Si hace falta saber quien escribe: wtrace (solo build-perfil, hay que recompilarlo con el mod) en 0x800555E0-0x80055600.
- MEDIDO prueba A/B (prueba_4p_sin_pantalla4.bat, 4 coches sin pantalla en 4, ~2+ min): NO se cierra -> culpa de las vistas extra.
- MEDIDO codigo (perfil\coches\ram_0009.bin): bloques de primitivas 2J = 0x80010080 y 0x80032968 (gp+4720 = 0x228E8 cada uno),
  final del bloque actual en gp+4716 (0x8013EAA8) = 0x80055250 para el bloque 1. El cierre 4 piso 0x800555F4 = final + 0x3A4.
  0x800BEDAC(vista) prepara el scratchpad: 0x1F800000 OT, +4 cursor de primitivas, +8 limite = final-416 (2J: la vista
  *(0x8013E6B0) (escena J1) se queda con la mitad de lo que queda). Con 4 escenas J3/J4 empiezan al final y se pasan.
- HECHO (sin probar) override 0x800BEDAC en nfshs_4p.c: escenas J1 1/4, J2 1/3, J3 1/2, J4 el resto, reserva 0x2000 para HUD.
  Registro: 'PRIMITIVAS fuera del bloque' (si aun se pasa) y 'LIM cuadro 600/601' (orden de vistas y limites).
  NFSHS_4P_LIMITE=0 = no tocar limites (solo medir). Copia previa: src/mods/nfshs_4p.c.antes_limite.
  Efecto esperado: menos objetos lejanos por escena (cada una tiene ~1/4 del bloque en vez de ~1/2).
- MEDIDO (usuario) con el override de limites: en la seleccion de coche J2 no confirma con K (antes siempre funcionaba).
  A/B: build-windows\prueba_4p_sin_limite.bat (NFSHS_4P_SIN_LIMITE=1 = no se instala el override 0x800BEDAC; log run_4p_sin_limite.log)
  frente a prueba_4p_teclado.bat. El registro dice al empezar 'override de limites: instalado / NO instalado'.
- MEDIDO A/B: sin el override (prueba_4p_sin_limite.bat) K funciona en la seleccion -> el override envolviendo siempre
  0x800BEDAC rompia la seleccion. CAMBIO: fuera de carrera 2J (estado 2) con pantalla en 4 devuelve 0 (rechaza: corre el original).
- MEDIDO 5a prueba (limites + rechazo fuera de carrera): K funciona; no se cerro en la prueba corta. Usuario: escenario aparece de a
  poco (poca memoria por escena), vistas parecidas y efecto 'zoom'. LIM cuadro 600: orden de vistas 4,5,2,3,6,7 (HUD/tacometros),
  0 (J1), 1 (J2), 6 (J3), 7 (J4); usado J1 ~14 KB, J2 ~37 KB (limite ~39 KB), J3 ~33 KB.
- MEDIDO codigo proyeccion: 0x800E2614 (camara de cada vista) -> 0x800E2554(x,y,w,h): GTE H = gp+660 (0x8013DAD0, 190),
  OFX = x+w/2, OFY = y+h/2; en 2J (0,0,320,120) -> centro 160,60 = borde derecho del cuadrante de 160 (se ve solo la mitad izq.).
- MEDIDO runtime: psx_mod_set_main_ram_8mb(1) en la activacion -> RAM de 8 MB (DMA de la GPU la soporta).
- HECHO (sin probar): pedida RAM 8 MB en la activacion (si pantalla en 4); override 0x800BEC4C mueve los bloques de primitivas a
  0x80200000/0x80300000 con 512 KB (0x80120714+20/+44, gp+4720) y 0x800BECE8 restaura los originales antes de liberar;
  override 0x800E2614: centro 80,60 y H/2 (vista de 1J a mitad de tamano). Copia previa: src/mods/nfshs_4p.c.antes_8mb.
  Riesgo: si el juego usa espejos de los 2 MB (direcciones 0x80200000+), la RAM de 8 MB lo rompe.
- MEDIDO 6a prueba (8 MB + proyeccion): no se cierra; RAM 8 MB y bloques en 0x80200000/0x80300000 OK. Uso por escena (LIM 600/601):
  J1 ~40 KB, J2 ~38 KB, J3 ~41 KB, limites ~128 KB -> el escenario que aparece de a poco YA NO es por memoria de primitivas.
  Usuario: leve mejora, aun 'zoom'. Causa (codigo): la camara vuelve a fijar H (0x8008275C -> SetGeomScreen 0x800EB4F8 con gp+660,
  0x80083704 SetGeomScreen(190)). CAMBIO (sin probar): override de SetGeomScreen 0x800EB4F8 -> H/2 en carrera con pantalla en 4;
  0x800E2614 solo centra (80,60). Copia previa: src/mods/nfshs_4p.c.antes_geomscreen.
  Pendiente aparicion de a poco: SUPUESTO estado por indice de camara fuera de la ranura (0x80083DFC llama p. ej. 0x80081330(indice));
  en la foto J2 y J4 se ven casi iguales.
- MEDIDO 7a prueba (video, SetGeomScreen H/2): J1 ya se ve con su coche (sin zoom). La camara se acerca/aleja al acelerar.
  Video: en la vista de J1 solo ~11 imagenes distintas por segundo (45 en 4 s) y el reloj de carrera avanza a tiempo real ->
  el juego va a ~11 fps. SUPUESTO: CPU emulada saturada (4 escenas sin guest_cycle_scale: el gate era solo estado 1) y la camara
  con pasos de tiempo grandes oscila.
- CAMBIO (sin probar): game.toml guest_cycle_scale_gated = true (sin gate de RAM; copia analysis\game.toml.antes_gate4p) y el vblank
  de nfshs.4p abre el gate en estado 1 o 2. Registro: 'guest_cycle_scale ACTIVO/apagado'. OJO: sin el mod 4p no hay escala 2.
  Requiere compilar_overlays.bat (cambia game.toml).
- MEDIDO 8a prueba (guest_cycle_scale en 2J): va mas fluido. Sigue: camara de J1 se acerca/aleja. Usuario: el escenario de
  J3/J4 aparece cerca de J1/J2 y se borra al alejarse -> lo que se dibuja depende de la posicion de J1/J2 (o de sus camaras).
  MEDIDO codigo: el dibujo de coches 0x800B3B90 usa tabla 0x80118064[buffer][j] y 0x800BF6FC usa coches[j] (0x80110CA0+4j, no
  humanos) con j = 0x80118238 (J3/J4 se dibujan con j = 1 -> datos de J2).
- HECHO (diagnostico, sin probar): en la logica de camaras, cuadros 300 y 900: 'DIFF original' y 'DIFF extra' = palabras que
  cambian en 0x80100000-0x8013FFFF y coches 0x801C4000-0x801C6FFF (sin ranuras de camara). Copia: nfshs_4p.c.antes_diffcam.
- MEDIDO diagnostico DIFF (cuadros 300/900): la logica de camaras (original y extra) cambia 0 palabras fuera de las ranuras
  (0x80100000-0x8013FFFF y coches) -> las pasadas extra NO dejan la camara de J1 distinta; el acercar/alejar viene de otra cosa.
  El escenario tampoco lo decide la logica de camaras.
- CAMBIO (sin probar): al dibujar J3/J4 tambien se intercambian coches[1] y coches[k] (0x80110CA0). NFSHS_4P_SIN_SWAP_COCHES=1
  lo desactiva. Copia: nfshs_4p.c.antes_swapcoches. A/B camara: prueba_4p_sin_pantalla4.bat (2 vistas) -> si J1 tambien se
  acerca/aleja, no es cosa de la pantalla en 4 (posible efecto de guest_cycle_scale en 2J o comportamiento normal).

## PUNTO ACTUAL (2026-11-17 noche) - 4P local EN PAUSA (el usuario lo deja para mas adelante)
- Estado de src/mods/nfshs_4p.c (todo compilado y probado salvo lo marcado):
  1. 4 coches en 2J, controles J3/J4 (teclado 8/0/7/9 y G/B/F/H, o mandos SDL). OK.
  2. Pantalla en 4 (cuadrantes 160x120, vistas 6/7 = J3/J4) y dibujo extra de J3/J4 con camara propia. OK.
  3. RAM 8 MB + bloques de primitivas en 0x80200000/0x80300000 (512 KB) -> ya NO se cierra (antes pisaba 0x80055250+). OK.
  4. Limites de primitivas por escena (override 0x800BEDAC, solo en carrera; fuera devuelve 0 -> K de J2 funciona). OK.
  5. Proyeccion: centro 80,60 (0x800E2614) + H/2 en SetGeomScreen 0x800EB4F8 -> J1 sin zoom. OK.
  6. guest_cycle_scale tambien en 2J: game.toml guest_cycle_scale_gated = true + gate abierto por el vblank de nfshs.4p en
     estado 1 o 2 -> mas fluido. OJO: con el mod nfshs.4p desactivado NINGUNA carrera tiene escala 2 (antes de 4P era
     guest_cycle_scale_gate estado 1; copia en analysis\game.toml.antes_gate4p). Al retomar: mover el gate a un mod siempre activo
     o volver al gate de RAM si se aparca el 4P.
  7. SIN PROBAR (compilar antes de usar): intercambio de coches[1]/coches[k] (0x80110CA0) al dibujar J3/J4, para el escenario.
     Desactivar con NFSHS_4P_SIN_SWAP_COCHES=1. Diagnostico DIFF (cuadros 300/900) sigue en el codigo (quitar cuando sobre).
- Problemas abiertos:
  a) Escenario de J3/J4 depende de la posicion de J1/J2 (aparece cerca de ellos, se borra al alejarse). No es memoria de
     primitivas (cada escena usa ~40 KB de ~128 KB) ni la logica de camaras (DIFF = 0). SUPUESTO: datos por jugador j al dibujar
     (J3/J4 se dibujan con j = 1): coches[j] (prueba 7 pendiente), tabla 0x80118064[buffer][j] (dibujo de coches 0x800B3B90).
  b) Camara de J1 se acerca/aleja al acelerar. No la causan las pasadas extra (DIFF = 0). Prueba pendiente: prueba_4p_sin_pantalla4.bat
     (2 vistas) para ver si pasa igual (entonces: guest_cycle_scale en 2J o comportamiento del juego).
  c) Despues: HUD/aguja de J3/J4 (ahora salen los tacometros viejos de las vistas 6/7), vista propia por jugador, opcion en launcher.
- Pruebas: build-windows\prueba_4p_teclado.bat (4P), prueba_4p_sin_pantalla4.bat (2 vistas), prueba_4p_sin_limite.bat.
- Copias de nfshs_4p.c de cada paso: src/mods/nfshs_4p.c.antes_{limite,8mb,geomscreen,diffcam,swapcoches}.
- Herramienta usada: perfil\disram.py <foto RAM> desde hasta (fotos en perfil\coches\ram_*.bin, carrera 2J).
- 2026-11-17 23:15 VUELTA A JUGAR NORMAL (sin 4P): game.toml restaurado EXACTO a analysis\game.toml.antes_gate4p
  (guest_cycle_scale 2 solo en carrera de 1 jugador, estado 1; el punto 6 de arriba queda deshecho) y mod nfshs.4p desactivado
  en build-windows\mods\state.toml (copia con 4P activo: analysis\state.toml.con_4p). Para retomar 4P: activar el mod en el
  launcher y volver a poner el gate (ver punto 6).

## PORT VITA (retomado 2026-11-17 noche) - paso 1: medir en el PC
- Decision del usuario: medir primero en que se va el tiempo del hilo del juego antes de tocar VitaSDK.
- Herramienta: perfil\25_perfil_vita.bat (E = normal, L = res. nativa + sin texturas HD) + perfil\vita_perfil.py:
  espera carrera 1J, 20 s de margen, mide 30 s: host_profile (muestreador de CPU siempre activo, funciones del .exe; nombres
  con nm en el VM de Claude), phase_profile (interprete/nativo/estatico/gpu), gcp_ring (funciones del juego PS1).
  Salida perfil\vita_<E|L>_*.json, perfil\vita_<E|L>.log y log del juego perfil\vita_juego_<E|L>.log (emu_busy).

## PUBLICACION "High Speed 4" (2026-11-17 noche, Vita en pausa)
- Decisiones del usuario: nombre High Speed 4; el codigo del juego se compila en el PC del jugador desde SU disco ([aot] mode = "none"
  + overlay_toolchain); texturas HD NO se reparten (solo scripts de hd/); codigo propio MIT; instalador Inno Setup + ZIP; BIOS: OpenBIOS
  incluida (MIT), no hace falta la del jugador.
- HECHO: LICENSE (MIT, Vanar/VanarTechlab), LICENCIAS.md (tabla de licencias, aviso EA/Sony, marcas), README.md nuevo (copia del
  antiguo en analysis\README.scaffold.md), packaging\release\START_HERE.txt, packaging\release.toml (HighSpeed4, exe HighSpeed4.exe,
  payload README/LICENSE/LICENCIAS), CMakeLists WINDOW_TITLE "High Speed 4" + EXE_NAME de desarrollo sin cambiar,
  portada ORIGINAL (tools\hs4_portada.py -> launcher_assets\img\boxart.png/.tga; la de EA en analysis\boxart_ea), .gitignore
  (generated/, hd/ imagenes, perfil/ datos, copias *.antes_*), installer\HighSpeed4.iss + installer\crear_instalador.bat
  (instala en %LOCALAPPDATA%\Programs\HighSpeed4, pide el .cue -> settings.toml [disc] path, tarea opcional: descarga clang
  portatil de RetroPortingToolKit/RetroPorting-Toolchains y escribe overlay_toolchain\compiler.txt).
- MEDIDO: overlay_captures.json lleva los BYTES de los overlays (bytes_b64 = codigo de EA) -> no se puede repartir. El runtime
  compila overlays con overlay_toolchain\compiler.txt (o gcc en PATH), si no con tcc (peor rendimiento).
- PENDIENTE: (1) probar si los overlays estan tal cual en el disco para precompilarlos en la instalacion sin repartir bytes
  (lista de offsets + sha256); (2) fork VanarTechlab/psxrecomp con los parches locales (package_title_release exige submodulos
  limpios) y repo publico NUEVO sin historial (el actual tiene generated/ con codigo de EA); (3) required_disc_sha256 en game.toml;
  (4) empaquetar con psxrecomp\tools\package_title_release.ps1 --channel public --platform windows; (5) instalador; (6) pagina web.
- MEDIDO 2026-11-18 (disco D:\Proyectos\NFS4\disco, 1 pista MODE2/2352, 733.165.440 bytes):
  sha256 del .bin (= required_disc_sha256 para game.toml) f0436c8dc7cf13d73974d75c2cc20bc35543726a5ca5bf51f88f961f6a28fe5d
  (no anadido aun: cambiar game.toml obliga a recompilar overlays; ponerlo al empaquetar).
- MEDIDO capturas vs disco (tools\hs4_disco\*.py): 431 capturas distintas (overlay_captures.json + .d); 400 estan TAL CUAL en el
  disco (alineadas a sector); las otras 31 coinciden en el 98,6% de trozos de 2 KB: las grandes (0x80099000 479 KB, 0x800A8000,
  0x800C4000, 0x800E0000) solo difieren ~50-60 bytes del ultimo trozo (datos despues del final del archivo, cambian en cada
  partida -> por eso hay varias versiones/CRC de 0x80099000); 0x80004000/0x80006000/0x80051000 (4 KB) no estan en el disco.
- MEDIDO disco ISO9660 (652 archivos): el motor es /NFS4.EXE (PS-X EXE, 1.245.184 bytes, sector 194; las capturas de
  0x80076000-0x800E5000 caen dentro) y el frontend /FRONT.BIN (281.048 bytes, sector 42; capturas 0x80028000-0x80041000).
  SLUS_008.26 (30 KB) es solo el cargador.
- PLAN codigo en el PC del jugador: perfil AOT de PSXRecomp (aot/overlays.json, schema "psxrecomp AOT methods v1":
  NFS4.EXE con method psx_exe, FRONT.BIN con fixed_address_extents; solo offsets/tamanos/sha256, sin bytes) y ejecutar
  psxrecomp/tools/aot_overlay_pipeline.py en la instalacion con el disco del jugador + clang -> cache local con guardas por
  bytes originales (evita el problema de los ~60 bytes variables). Pendiente: probarlo en dev (rendimiento y cobertura) y ver
  si el pipeline corre con el python/recompilador de overlay_toolchain (pide Python 3.11+).
- MEDIDO cabeceras: NFS4.EXE = PS-X EXE (t_addr 0x80010000, t_size 0x12F800, pc0 0x800E4C38); FRONT.BIN empieza por "SimpleMe"
  (no es PS-X EXE) y las capturas indican que se carga entero en 0x80010000 (encima del motor, en los menus).
- HECHO (sin probar): aot/overlays.json (psx_exe NFS4.EXE + fixed_address_files FRONT.BIN @0x80010000, disc_hashes sha256,
  expected_records 0 a proposito: el primer extract dice cuantas recetas hay) y scripts/hs4_aot.sh (extract | release,
  registro en build-aot/hs4_aot_<accion>.log). Siguiente: ajustar expected_records / exclusiones con el registro.
- MEDIDO AOT desde el disco (2026-11-18): NFS4.EXE 0x80010000-0x80055000 son ceros (hueco para FRONT.BIN en menus / bloques de
  dibujo en carrera) y el codigo acaba en 0x8010D930. Perfil final: fixed_address_extents NFS4.EXE 0x80055000+0xB8930 (sha256
  9fcff029...), FRONT.BIN fixed_address_files @0x80010000 (allow_missing), NFS4.EXE en generic_exclude_files, expected_records 2.
  bash scripts/hs4_aot.sh release -> "Staged 2 audited native pairs", 3853 filas de manifiesto, all_guards_match_known_input_bytes
  true, ~16 min con clang en el PC del usuario. Cache en build-aot/stage/cache (cg18_58f1ec8c_gc81b904fa_f0).
  Prueba en juego: perfil\27_probar_cache_aot.bat (solo esa cache, sin autocompilar) -> perfil\aot_juego.log.
- MEDIDO prueba en juego con SOLO la cache AOT (perfil\aot_juego.log): ~60 fps en menus y carrera (fps 59.6-60.0), emu_busy
  13-53% (un pico de 82% con 1 underrun de audio); el usuario lo vio todo bien.
- HECHO: CMake option HS4_DISTRIBUCION (release.toml la pone ON): sin nfshs_unlocks.c ni nfshs_4p.c y catalogo de mods generado
  en build/hs4_mods_distribucion solo con nfshs.widescreen + nfshs.jukebox. Copia: analysis\CMakeLists.antes_distribucion.txt.
  README sin la linea de desbloqueos.
- PENDIENTE instalador: ejecutar aot_overlay_pipeline.py en el PC del jugador -> hay que meter en el paquete aot/overlays.json y
  las herramientas de psxrecomp/tools que usa (aot_overlay_pipeline.py y sus modulos) y llamar con el python 3.13 y el
  psxrecomp-game.exe de overlay_toolchain y el clang descargado; tarda ~16 min en el PC del usuario.
- HECHO (sin probar) preparar en el PC del jugador: installer\hs4_preparar.py (python 3.13 de overlay_toolchain; monta
  hs4_aot\psxrecomp con las 15 tools de psxrecomp/tools que usa el pipeline + include/bios/recompiler copiados de
  overlay_toolchain; ejecuta aot_overlay_pipeline release con el .cue de settings.toml y clang de compiler.txt; copia la cache
  a <juego>\cache), "installer\Preparar juego.bat", payload en release.toml (hs4_aot/...), y el instalador lo lanza al final
  (casilla, solo si hay clang y .cue). Prueba en dev: perfil\\28_probar_preparar.bat simula la carpeta instalada (build-prueba-instalado).
- MEDIDO perfil\28_probar_preparar.bat (carpeta instalada simulada, python 3.13 incluido): "Staged 2 audited native pairs",
  mismos archivos/tamanos/tag (cg18_58f1ec8c_gc81b904fa_f0, 00010000_D9C43CAB + 00055000_F51D3A26) que la prueba con MSYS2;
  3 minutos con workers = nucleos-1. Lista la parte de preparar en el PC del jugador.
- HECHO game.toml (copia analysis\game.toml.antes_hs4): [game] name "High Speed 4", required_disc_sha256 = sha256 del .bin,
  [runtime] window_title "High Speed 4". Cambia el tag de la cache -> en dev recompilar overlays.
- HECHO scripts/hs4_preparar_github.sh: comprueba fork VanarTechlab/psxrecomp y repo vacio VanarTechlab/High-Speed-4, commit
  de los parches de psxrecomp en rama high-speed-4 y push al fork, .gitmodules -> fork, rama huerfana "publico" (sin historial,
  sin generated/) y push a main del repo nuevo. Pide confirmacion y muestra lo que se publica.
- CAMBIO 2026-11-18: instalador al estilo de Belial 64 (pedido del usuario), Inno Setup descartado (analysis\inno_descartado):
  installer\go (Go + API de Windows, adaptado de Belial-64\n64\instalador; core.go: comprueba sha256 del .bin del .cue,
  copia juego\ a la carpeta, settings.toml [disc] path, descarga clang a %LOCALAPPDATA%\retcomm\toolchains\cmake-clang-v1\latest
  si no esta, compiler.txt, ejecuta hs4_aot\hs4_preparar.py con el python incluido y muestra su salida, accesos directos),
  exes compilados en installer\bin (GOOS=windows; "Instalar High Speed 4.exe" -tags gui, instalar_hs4.exe consola),
  installer\fondo.py -> herramientas\fondo.bmp (original), installer\LEEME.txt (como el de Belial: necesitas, instalacion,
  controles medidos de input.ini/keybinds.ini, notas, licencias, ingles), installer\crear_distribucion.bat -> dist\HighSpeed4_vX.zip
  (exes + LEEME + herramientas + juego\ del paquete + LICENCIAS con el codigo del instalador). Recompilar exes: en la nube con Go.
- DECISION usuario 2026-11-18: esta primera publicacion es la version 0.8 (VERSION = 0.8; Version del instalador Go = "0.8").
- HECHO (sin probar) KIT HD en la 0.8 (decision usuario: opcion B): installer\go_hd (Go) -> "Texturas HD.exe" (ventana) y
  texturas_hd.exe (consola) en la carpeta del juego. Botones: Recoger texturas (mods\state.toml: hd-textures enabled + dump),
  Crear e instalar (dumps nuevos sin ST* ni excluidas -> Real-ESRGAN realesrgan-x4plus-anime x4 -> filtro de color como
  verificar.py (>20) -> x4/x2 segun hd_kit\decisiones.txt (7.801 x4 / 31.989 x2 del pack de Vanar) o detalle como
  preparar_pack.py (remuestreo tipo PIL, comprobado contra PIL: +-0,02) -> replacements; activa HD y apaga dump), Activar/Desactivar.
  hd_kit\excluidas.txt = 242 escaladas que no entraron al pack. Modelo viaja como .bin.modelo (el empaquetador rechaza .bin).
  Payload en release.toml; licencias BSD-3 de Real-ESRGAN/ncnn en hd_kit\LICENCIAS_kit_HD.txt; LEEME con seccion TEXTURAS HD.
  Ojo device_commit_files: no sobrescribe archivos existentes ni reenvia un stagedPath ya usado -> nombres nuevos + cat > destino.
- DECISION usuario (opcion A): pack HD como DESCARGA APARTE opcional + kit como alternativa. HECHO: Texturas HD.exe con boton
  "A) Instalar pack descargado" (zip -> todos los .png a replacements, activa HD; probado en Go con zip de prueba),
  installer\crear_pack_hd.bat -> dist\HighSpeed4_TexturasHD_vX.zip (tar de Windows: hd\pack_anime 39.790 texturas, 731 MB +
  installer\pack_hd\LEEME_TexturasHD.txt con aviso). Recomendado alojarlo FUERA del repo/cuenta del codigo. LEEME/README/LICENCIAS actualizados.
- DECISION usuario: el pack HD (HighSpeed4_TexturasHD_vX.zip) se aloja en archive.org (fuera del GitHub del codigo);
  identificador sugerido high-speed-4-texturas-hd-v0.8; la web enlazara a https://archive.org/download/<identificador>/<zip>.
- MEDIDO usuario: pack HD instalado con "A)" funciona. "Todo desbloqueado" en la prueba = la prueba copia build-windows (desarrollo)
  con mods\state.toml (nfshs.unlocks activado) y mods\bundled\nfshs.unlocks. CAMBIO: perfil\29 excluye state.toml, nfshs.unlocks y nfshs.4p.
  El paquete real (HS4_DISTRIBUCION=ON) no lleva ni el codigo ni el paquete de esos mods, ni state.toml.
