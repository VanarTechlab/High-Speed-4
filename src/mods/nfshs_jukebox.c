/* NFS High Stakes Recompiled - Jukebox (trusted static plugin).
 *
 * La musica de carrera esta en ZZZZYMUS.VIV (archivo BIGF de EA, LBA 126428):
 * 18 temas "gameNN-..." + "lose" + "win". El juego elige un tema al azar al
 * empezar la carrera, al acabar el tema anterior y al quitar la pausa, hace
 * SetLoc al sector donde empieza y lo lee en flujo.
 *
 * Jukebox: no toca la RAM del juego. Con un tema elegido, cuando el juego lee
 * sectores del tema que el escogio, PSXRecomp le entrega (byte a byte, mismas
 * posiciones relativas) los datos del tema elegido -- hook en cdrom.c
 * (psx_cdrom_set_user_remap). Si el elegido es mas corto se repite; si es mas
 * largo se corta cuando el juego cree que su tema termino (y el siguiente
 * vuelve a ser el elegido). "Aleatorio" = juego original.
 *
 * Menu de pausa: con la carrera en pausa (0x80118234 != 0, medido) el titulo
 * "Game Paused" (11 letras, se busca en RAM) muestra el tema ("Tema 05/18" o
 * "Aleatorio") y L1/R1 lo cambian. Al quitar la pausa el juego empieza un tema
 * nuevo, que el desvio convierte en el elegido.
 */
#include "mod_plugins.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define PKG  "nfshs.jukebox"
#define FEAT "jukebox"
#define VIV_LBA 126428
#define NTRACKS 18

typedef int (*psx_cdrom_user_remap_fn)(int lba, uint8_t *user);
extern void psx_cdrom_set_user_remap(psx_cdrom_user_remap_fn fn);
extern int  psx_cdrom_read_user_sector(int lba, uint8_t *out);
extern int  cdrom_get_setloc_lba(void);

/* Offsets/tamanos (bytes, dentro del VIV) leidos del disco SLUS-00826. */
static const uint32_t k_off[NTRACKS] = {
    808u, 4588128u, 9703928u, 13646320u, 17562312u, 21903728u, 26781016u,
    30997712u, 35294392u, 39627472u, 42886776u, 47244736u, 51272440u,
    56405040u, 61418232u, 66969856u, 72690424u, 77646320u };
static const uint32_t k_size[NTRACKS] = {
    4587320u, 5115800u, 3942392u, 3915992u, 4341416u, 4877288u, 4216696u,
    4296680u, 4333080u, 3259304u, 4357960u, 4027704u, 5132600u, 5013192u,
    5551624u, 5720568u, 4955896u, 3729496u };

static const char *const k_name[NTRACKS] = {
    "game01-amorphous being-saki kaskas",
    "game02-roll the dice-lunatic calm",
    "game03-rock this-rom di prisco",
    "game04-naked and ashamed-dylan rhymes",
    "game05-war-junkie xl",
    "game06-bionic-crispin hands",
    "game07-i am electro-the funk lab",
    "game08-liquid plasma-rom di prisco",
    "game09-fight-junkie xl",
    "game10-clutch-dj icey",
    "game11-bring that beat back-rom di prisco",
    "game12-dude in the moon-dastrix",
    "game13-def beat-junkie xl",
    "game14-cost of freedom-the experiment",
    "game15-globular cluster-saki kaskas",
    "game16-no remorse-junkie xl",
    "game17-insanity sauce -surreal madrid",
    "game18-road warrior-rom di prisco",
};

/* Nombres en la cabecera del VIV que el juego tiene en RAM (direccion variable,
 * medido 0x801D8B78 / 0x801EEA78 / 0x801F0A88). El cuadro "tema que suena" lee
 * de ahi: se escribe el nombre del elegido en el hueco de cada entrada (mismo
 * largo, relleno con espacios, nunca se mueve un 0 de sitio). */
static uint32_t s_hdr;
static uint32_t s_state;   /* estado del reproductor de musica (ver find_state) */
static int s_names_choice = 0;   /* que nombres hay escritos (0 = originales) */

static int is_music_header(uint32_t a) {
    return psx_mod_read_word(a) == 0x46474942u &&                 /* "BIGF" */
           psx_mod_read_byte(a + 4) == 0x05 && psx_mod_read_byte(a + 5) == 0x88 &&
           psx_mod_read_byte(a + 6) == 0x06 && psx_mod_read_byte(a + 7) == 0x68 &&
           psx_mod_read_byte(a + 11) == 20;
}
static uint32_t find_header(void) {
    if (s_hdr && is_music_header(s_hdr)) return s_hdr;
    s_hdr = 0; s_names_choice = 0;
    static unsigned throttle;
    if (throttle++ % 60) return 0;          /* buscar como mucho 1 vez por segundo */
    for (uint32_t a = 0x80100000u; a < 0x80200000u; a += 4)
        if (psx_mod_read_word(a) == 0x46474942u && is_music_header(a)) return s_hdr = a;
    return 0;
}
static int state_ok(uint32_t st, uint32_t h) {
    uint32_t ptr = psx_mod_read_word(st + 0x38u);
    return psx_mod_read_word(st + 0x18u) == NTRACKS &&
           psx_mod_read_word(st + 0x1Cu) < NTRACKS &&
           ptr > h + 16u && ptr < h + 808u;
}
static uint32_t find_state(uint32_t h) {
    if (s_state && state_ok(s_state, h)) return s_state;
    s_state = 0;
    static unsigned throttle;
    if (throttle++ % 60) return 0;
    for (uint32_t a = 0x80100000u; a < 0x80200000u; a += 4) {
        uint32_t v = psx_mod_read_word(a);
        if (v > h + 16u && v < h + 808u && a >= 0x80100038u && state_ok(a - 0x38u, h)) {
            s_state = a - 0x38u;
            fprintf(stderr, "[nfshs.jukebox] estado de la musica en 0x%08X (cabecera 0x%08X)\n",
                    (unsigned)s_state, (unsigned)h);
            return s_state;
        }
    }
    return 0;
}
static void fit_name(int c, int len, char *out) {
    const char *n = k_name[c];
    int L = (int)strlen(n);
    if (L > len) {                       /* quitar el artista si no cabe */
        const char *dash = strrchr(n, '-');
        int base = dash ? (int)(dash - n) : L;
        L = base < len ? base : len;
    }
    for (int i = 0; i < len; i++) out[i] = i < L ? n[i] : ' ';
}
static void apply_names(int choice) {
    uint32_t h = find_header();
    if (!h || choice == s_names_choice) return;
    uint32_t p = h + 16u;
    char buf[64];
    for (int i = 0; i < NTRACKS; i++) {
        int len = (int)strlen(k_name[i]);
        uint32_t nm = p + 8u;
        /* comprobar que el hueco sigue siendo el esperado (0 al final) */
        if (psx_mod_read_byte(nm + (uint32_t)len) != 0) { s_hdr = 0; return; }
        if (choice) fit_name(choice - 1, len, buf); else memcpy(buf, k_name[i], (size_t)len);
        for (int k = 0; k < len; k++)
            if (psx_mod_read_byte(nm + (uint32_t)k) != (uint8_t)buf[k])
                psx_mod_write_byte(nm + (uint32_t)k, (uint8_t)buf[k]);
        p = nm + (uint32_t)len + 1u;
    }
    s_names_choice = choice;
}

static volatile int s_choice = 0;   /* 0 = aleatorio, 1..18 = tema fijo */
static int s_cur = -1;              /* tema que el juego cree estar leyendo */
static int s_last_setloc = -1;
static int s_cache_lba = -1;
static uint8_t s_cache[2048];

/* Estado del reproductor de musica (direccion variable, medido 0x801C6A00):
 *   +0x18 = 18 (numero de temas), +0x1C = indice (0-17) del tema actual,
 *   +0x38 = puntero al nombre del tema dentro de la cabecera del VIV.
 * Se localiza buscando ese puntero (apunta dentro de la cabecera): find_state. */
#define ADDR_GAMESTATE 0x8010FB90u   /* 1 = en carrera (y su pausa) */
#define ADDR_PAUSED    0x80118234u   /* != 0 con el menu de pausa abierto */
/* "Game Paused\0Continue" esta en los datos de un overlay cuya direccion varia
 * entre sesiones (medido 0x8015261C y 0x8015265C): se busca al abrir la pausa. */
#define SEARCH_LO      0x80140000u
#define SEARCH_HI      0x80170000u
static uint32_t s_title_addr;        /* 0 = no encontrado */
static int s_searched;               /* fotogramas buscando en esta pausa */
#define ADDR_TITLE     (s_title_addr)
#define ADDR_CONTINUE  (s_title_addr + 12u)
#define TITLE_LEN      11
#define BTN_L1 (1u << 10)
#define BTN_R1 (1u << 11)
#define BTN_UP   (1u << 4)
#define BTN_DOWN (1u << 6)
#define BTN_RIGHT (1u << 5)
/* Refresco del titulo: el juego solo redibuja el menu al mover el cursor, asi
 * que tras L1/R1 el mod "pulsa" abajo y luego arriba (el cursor queda igual).
 * Bucle cerrado sobre el cursor (0x80152A28, 1-4, medido): pulsa hasta que el
 * cursor se mueve, suelta, y pulsa al reves hasta que vuelve a su sitio. */
#define ADDR_CURSOR 0x80152A28u
static volatile uint32_t s_press;   /* botones que el mod mantiene pulsados */
static int s_inj_phase;             /* 0 nada, 1 ida, 2 pausa, 3 vuelta, 4 suelta */
static int s_inj_timer, s_inj_c0;
static uint32_t s_inj_first, s_inj_back;
static volatile uint32_t s_edge;     /* pulsaciones L1/R1 pendientes */
static uint32_t s_prev_held;
static int s_title_written;
static volatile uint32_t s_pad_calls;  /* llamadas al observador del mando */
static uint32_t s_pad_calls_seen, s_frames_without_pad;

static int start_sector(int i) { return VIV_LBA + (int)(k_off[i] / 2048u); }
static int end_sector(int i)   { return VIV_LBA + (int)((k_off[i] + k_size[i] - 1u) / 2048u); }

static int parse_choice(const char* v) {
    if (!v || !*v || strcmp(v, "random") == 0) return 0;
    int n = atoi(v);
    return (n >= 1 && n <= NTRACKS) ? n : 0;
}

static void track_current(void) {
    int sl = cdrom_get_setloc_lba();
    if (sl == s_last_setloc) return;
    s_last_setloc = sl;
    for (int i = 0; i < NTRACKS; i++)
        if (sl == start_sector(i)) { s_cur = i; return; }
}

static int src_byte(uint32_t viv_pos, uint8_t *out) {
    int lba = VIV_LBA + (int)(viv_pos / 2048u);
    if (lba != s_cache_lba) {
        if (!psx_cdrom_read_user_sector(lba, s_cache)) return 0;
        s_cache_lba = lba;
    }
    *out = s_cache[viv_pos % 2048u];
    return 1;
}

static int jukebox_remap(int lba, uint8_t *user) {
    track_current();
    int c = s_choice - 1;
    if (c < 0) return 0;
    int cur = s_cur;
    if (cur < 0 || cur == c) return 0;
    if (lba < start_sector(cur) || lba > end_sector(cur)) return 0;
    int64_t sector_pos = (int64_t)(lba - VIV_LBA) * 2048;
    for (int k = 0; k < 2048; k++) {
        int64_t p = sector_pos + k - (int64_t)k_off[cur];   /* posicion dentro del tema */
        if (p < 0 || p >= (int64_t)k_size[cur]) continue;  /* byte de un tema vecino */
        uint32_t q = (uint32_t)(p % (int64_t)k_size[c]);
        if (!src_byte(k_off[c] + q, &user[k])) return 0;
    }
    return 1;
}

/* Observa el mando (no lo cambia: devuelve 0 = mando original). */
static int jukebox_pad(const PSXModPadFrame *f, PSXModPadOutput *out) {
    if (s_pad_calls++ == 0)
        fprintf(stderr, "[nfshs.jukebox] mando observado (player %u, botones 0x%04x)\n",
                (unsigned)f->player, (unsigned)f->buttons);
    uint32_t held = (~f->buttons) & 0xFFFFu;       /* activo-bajo -> pulsado=1 */
    s_edge |= held & ~s_prev_held & (BTN_L1 | BTN_R1);
    s_prev_held = held;
    if (!s_inj_phase) return 0;
    uint32_t press = s_press;
    out->buttons = (out->buttons | BTN_RIGHT) & ~press & 0xFFFFu;
    return 1;
}

static int title_is(const char *t) {
    for (int i = 0; i <= TITLE_LEN; i++) {
        uint8_t c = psx_mod_read_byte(ADDR_TITLE + (uint32_t)i);
        if (c != (uint8_t)t[i]) return 0;
        if (!t[i]) return 1;
    }
    return 1;
}
static int continue_at(uint32_t a) {
    static const char k[] = "Continue";
    for (int i = 0; i < 9; i++)
        if (psx_mod_read_byte(a + (uint32_t)i) != (uint8_t)k[i]) return 0;
    return 1;
}
static int title_ok(uint32_t t) {
    uint8_t c0 = psx_mod_read_byte(t), c1 = psx_mod_read_byte(t + 1);
    return (c0 == 'G' && c1 == 'a') || (c0 == 'T' && c1 == 'e') || (c0 == 'A' && c1 == 'l');
}
static int menu_text_present(void) {
    if (s_title_addr && continue_at(s_title_addr + 12u)) return 1;
    s_title_addr = 0;
    /* Reintentar cada 15 fotogramas: al abrir la pausa el texto puede tardar en cargarse. */
    if (s_searched++ % 15) return 0;
    /* "Game Paused" son 11 letras + 0: "Continue\0" empieza 12 bytes despues. */
    for (uint32_t a = SEARCH_LO; a < SEARCH_HI; a += 4) {
        if (psx_mod_read_word(a) != 0x746E6F43u) continue;     /* "Cont" alineado */
        if (continue_at(a) && psx_mod_read_byte(a - 1) == 0 && title_ok(a - 12u)) { s_title_addr = a - 12u; return 1; }
    }
    for (uint32_t a = SEARCH_LO; a < SEARCH_HI; a++)          /* sin alinear */
        if (psx_mod_read_byte(a) == 'C' && continue_at(a) && psx_mod_read_byte(a - 1) == 0 &&
            title_ok(a - 12u)) {
            s_title_addr = a - 12u; return 1;
        }
    return 0;
}
static void write_title(const char *t) {
    int n = (int)strlen(t);
    for (int i = 0; i < TITLE_LEN; i++)
        psx_mod_write_byte(ADDR_TITLE + (uint32_t)i, i < n ? (uint8_t)t[i] : 0);
    psx_mod_write_byte(ADDR_TITLE + TITLE_LEN, 0);
}
static void label(char *b) {
    if (s_choice == 0) { strcpy(b, "Aleatorio"); return; }
    b[0]='T'; b[1]='e'; b[2]='m'; b[3]='a'; b[4]=' ';
    b[5]=(char)('0' + s_choice / 10); b[6]=(char)('0' + s_choice % 10);
    b[7]='/'; b[8]='1'; b[9]='8'; b[10]=0;
}


/* --- Diagnostico temporal: buscar donde guarda el juego el numero del tema ---
 * Cada vez que el juego empieza un tema (s_cur cambia), se cruzan las
 * direcciones de RAM cuyo byte vale s_cur (0-18) o s_cur+1; se imprime lo que
 * queda. Activado solo con PSX_JUKEBOX_DIAG=1. */
static uint8_t s_cand0[0x200000 / 8], s_cand1[0x200000 / 8];
static int s_diag_rounds = -1, s_diag_last_cur = -1;
static void diag_round(int cur) {
    int first = (s_diag_rounds < 0);
    unsigned n0 = 0, n1 = 0;
    for (uint32_t o = 0; o < 0x200000u; o++) {
        uint8_t v = psx_mod_read_byte(0x80000000u + o);
        uint8_t m = (uint8_t)(1u << (o & 7));
        if (first) {
            if (v == cur) s_cand0[o >> 3] |= m;
            if (v == cur + 1) s_cand1[o >> 3] |= m;
        } else {
            if (v != cur) s_cand0[o >> 3] &= (uint8_t)~m;
            if (v != cur + 1) s_cand1[o >> 3] &= (uint8_t)~m;
        }
        if (s_cand0[o >> 3] & m) n0++;
        if (s_cand1[o >> 3] & m) n1++;
    }
    s_diag_rounds = first ? 1 : s_diag_rounds + 1;
    fprintf(stderr, "[nfshs.jukebox diag] ronda %d: tema del juego %d -> quedan %u (=indice) y %u (=indice+1)\n",
            s_diag_rounds, cur + 1, n0, n1);
    if (n0 <= 40 || n1 <= 40) {
        for (uint32_t o = 0; o < 0x200000u; o++) {
            uint8_t m = (uint8_t)(1u << (o & 7));
            if (n0 <= 40 && (s_cand0[o >> 3] & m)) fprintf(stderr, "  idx   0x%08X\n", 0x80000000u + o);
            if (n1 <= 40 && (s_cand1[o >> 3] & m)) fprintf(stderr, "  idx+1 0x%08X\n", 0x80000000u + o);
        }
    }
}
static void diag_tick(void) {
    static int on = -1;
    if (on < 0) { const char *e = getenv("PSX_JUKEBOX_DIAG"); on = e && e[0] == '1'; }
    if (!on || s_cur < 0 || s_cur == s_diag_last_cur) return;
    s_diag_last_cur = s_cur;
    /* El tema 1 empieza en el mismo sector que la cabecera del VIV (el juego la
     * lee al empezar la carrera): esas rondas no son fiables, se saltan. */
    if (s_cur == 0) return;
    diag_round(s_cur);
}

static void register_pad(void) {
    static PSXModPadTransform t;
    memset(&t, 0, sizeof t);
    t.struct_size = sizeof t;
    t.allowed_types = PSX_MOD_PAD_TYPE_BIT(PSX_MOD_PAD_DIGITAL) |
                      PSX_MOD_PAD_TYPE_BIT(PSX_MOD_PAD_DUALSHOCK);
    t.initial_type = PSX_MOD_PAD_DUALSHOCK;
    t.transform = jukebox_pad;
    int ok = psx_mod_set_pad_transform(0, &t);
    fprintf(stderr, "[nfshs.jukebox] observador del mando registrado: %s\n", ok ? "si" : "NO");
}

static void nfshs_jukebox_vblank(void) {
    if (!psx_mod_game_started()) return;
    if (psx_mod_read_word(ADDR_GAMESTATE) == 1u) {
        apply_names(s_choice);
        /* Numero del cuadro "tema que suena": el juego lo saca de su indice.
         * Solo se cambia si ese byte vale justo el tema que el juego empezo
         * (s_cur), para no escribir nunca en una direccion equivocada. */
        int c = s_choice - 1;
        uint32_t st = s_hdr ? find_state(s_hdr) : 0;
        if (st && c >= 0 && s_cur >= 0 && s_cur != c &&
            psx_mod_read_word(st + 0x1Cu) == (uint32_t)s_cur) {
            psx_mod_write_word(st + 0x1Cu, (uint32_t)c);
            fprintf(stderr, "[nfshs.jukebox] indice del juego %d -> %d\n", s_cur + 1, c + 1);
        }
        diag_tick();
    }
    else s_hdr = 0;
    /* Si un reinicio de sesion quito el observador, volver a ponerlo. */
    if (s_pad_calls != s_pad_calls_seen) { s_pad_calls_seen = s_pad_calls; s_frames_without_pad = 0; }
    else if (++s_frames_without_pad == 120) register_pad();
    int paused = psx_mod_read_word(ADDR_GAMESTATE) == 1u &&
                 psx_mod_read_byte(ADDR_PAUSED) != 0;
    if (paused && !menu_text_present()) paused = 0;
    uint32_t e = s_edge; s_edge = 0;
    if (!paused) {
        if (s_title_written && s_title_addr && continue_at(s_title_addr + 12u))
            write_title("Game Paused");
        s_title_written = 0;
        s_searched = 0;
        s_inj_phase = 0; s_press = 0;
        return;
    }
    /* Refresco del titulo: el juego redibuja el menu de pausa al pulsar
     * izquierda/derecha, sin mover el cursor (probado 2026-11-16). Se pulsa
     * derecha 4 fotogramas y se suelta otros 4. */
    if (s_inj_phase) {
        s_inj_timer++;
        if (s_inj_phase == 1) {
            s_press = BTN_RIGHT;
            if (s_inj_timer >= 4) { s_inj_phase = 2; s_inj_timer = 0; s_press = 0; }
        } else {
            s_press = 0;
            if (s_inj_timer >= 4) s_inj_phase = 0;
        }
    }
    int c = s_choice;
    if (s_inj_phase) e = 0;          /* esperar a que termine el refresco anterior */
    if (e & BTN_R1) c = (c + 1) % (NTRACKS + 1);
    if (e & BTN_L1) c = (c + NTRACKS) % (NTRACKS + 1);
    static int no_refresh = -1;     /* PSX_JUKEBOX_NO_REFRESH=1: sin refresco (pruebas) */
    if (no_refresh < 0) { const char *ev = getenv("PSX_JUKEBOX_NO_REFRESH"); no_refresh = ev && ev[0] == '1'; }
    if (c != s_choice && !s_inj_phase && !no_refresh) {          /* forzar redibujado */
        s_inj_phase = 1; s_inj_timer = 0;
    }
    s_choice = c;
    char b[16]; label(b);
    if (!title_is(b)) write_title(b);
    s_title_written = 1;
}

static void read_choice(void) {
    char v[16];
    s_choice = psx_mod_option_value(PKG, FEAT, "track", v, sizeof v) ? parse_choice(v) : 0;
}

static void nfshs_jukebox_activate(void) {
    read_choice();
    s_cur = -1; s_last_setloc = -1; s_cache_lba = -1;
    psx_cdrom_set_user_remap(jukebox_remap);
    register_pad();
    s_title_written = 0; s_edge = 0; s_prev_held = 0;
}

static int nfshs_jukebox_option_changed(const char* option_id, const char* value) {
    if (option_id && strcmp(option_id, "track") == 0) {
        s_choice = parse_choice(value);
        return 1;
    }
    return 0;
}

PSX_MOD_CONSTRUCTOR(nfshs_jukebox_register) {
    psx_mod_register_activation_plugin("nfshs.jukebox", nfshs_jukebox_activate);
    psx_mod_register_option_changed_plugin("nfshs.jukebox", nfshs_jukebox_option_changed);
    psx_mod_register_vblank_plugin("nfshs.jukebox", nfshs_jukebox_vblank);
}
