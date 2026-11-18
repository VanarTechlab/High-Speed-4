/* NFS High Stakes Recompiled - widescreen (trusted static plugin).
 * Activated by mods/preloaded/packages/nfshs.widescreen.
 *  1. Selects a fixed display aspect before the renderer/window starts; the
 *     framework then widens the GTE projection (wider FOV, no stretch).
 *  2. Speedometer needle fix: the tacho routine (0x800D4A68) draws the needle
 *     as a PolyF3 + LineF shadow inside its own 28x28 draw area / draw offset
 *     (E3/E4/E5). auto_ui_squash only proportion-corrects axis-aligned quads,
 *     so the needle stayed stretched and drifted right of its dial. We wrap
 *     the routine and, while widescreen is active, squash the clip window and
 *     origin about the display centre and the needle X about its pivot. */
#include "mod_plugins.h"
#include "func_override.h"
#include <string.h>

#define NFS_TACHO_FN      0x800D4A68u
#define NFS_PKT_CURSOR    0x1F800004u   /* scratchpad: next free packet ptr */
#define NFS_DIAL_SIZE     28            /* dial sprite is 28x28 (dial-local) */
#include <stdio.h>
static int s_log_left = 6;   /* first calls logged to stderr (diagnostics) */
static int s_fixed;
static int s_dump;        /* dump raw needle prims this call */
static uint32_t s_calls;

static int32_t s_m;   /* live widescreen margin (px per side, 320 basis) */

static int32_t squash_rel(int32_t v, int32_t centre) {
    /* v' = centre + (v - centre) * 160 / (160 + m), rounded */
    int32_t d = (v - centre) * 160;
    int32_t den = 160 + s_m;
    return centre + (d >= 0 ? (d + den / 2) / den : -((-d + den / 2) / den));
}
static int32_t buf_origin(int32_t x) { return x >= 320 ? 320 : 0; }

/* Packets are written needle-first, then the E3/E4/E5 window packet, so do two
 * passes: find the dial window, then squash both the window and the needle. */
static int is_window_cmd(uint32_t w, int* full) {
    uint32_t code = w >> 24;
    if (code == 0xE3 || code == 0xE4) {
        int32_t x = (int32_t)(w & 0x3FF), o = buf_origin(x);
        *full = (code == 0xE3) ? (x == o) : (x == o + 319);
        return 1;
    }
    if (code == 0xE5) {
        int32_t x = (int32_t)(w & 0x7FF);
        if (x & 0x400) x -= 0x800;
        *full = (x == buf_origin(x));
        return 1;
    }
    return 0;
}

static int32_t sq(int32_t v) { return v * v; }
/* Clip segment (x0,y0)-(x1,y1) to [0,w]x[0,h] (Liang-Barsky, t in 1/4096).
 * A segment fully outside collapses onto its clamped first point. */
static void clip_line(int32_t* x, int32_t* y, int32_t w, int32_t h) {
    int32_t dx = x[1] - x[0], dy = y[1] - y[0];
    int64_t t0 = 0, t1 = 4096;
    int32_t p[4] = { -dx, dx, -dy, dy };
    int32_t q[4] = { x[0], w - x[0], y[0], h - y[0] };
    for (int i = 0; i < 4; i++) {
        if (p[i] == 0) { if (q[i] < 0) { t0 = 1; t1 = 0; break; } continue; }
        int64_t t = (int64_t)q[i] * 4096 / p[i];
        if (p[i] < 0) { if (t > t0) t0 = t; } else { if (t < t1) t1 = t; }
    }
    if (t0 > t1) {
        int32_t cx = x[0] < 0 ? 0 : (x[0] > w ? w : x[0]);
        int32_t cy = y[0] < 0 ? 0 : (y[0] > h ? h : y[0]);
        x[0] = x[1] = cx; y[0] = y[1] = cy;
        return;
    }
    int32_t nx0 = x[0] + (int32_t)(dx * t0 / 4096), ny0 = y[0] + (int32_t)(dy * t0 / 4096);
    int32_t nx1 = x[0] + (int32_t)(dx * t1 / 4096), ny1 = y[0] + (int32_t)(dy * t1 / 4096);
    x[0] = nx0; y[0] = ny0; x[1] = nx1; y[1] = ny1;
    /* integer rounding can leave an end 1 px outside (e.g. y = -1): clamp,
     * so the piece is fully ENCLOSED by the dial and PSXRecomp's auto-UI
     * moves it together with the dial (an un-enclosed piece keeps raw X). */
    for (int i = 0; i < 2; i++) {
        if (x[i] < 0) x[i] = 0; else if (x[i] > w) x[i] = w;
        if (y[i] < 0) y[i] = 0; else if (y[i] > h) y[i] = h;
    }
}

static void fix_tacho_packets(uint32_t start, uint32_t end) {
    uint32_t p; int guard, has_window = 0, full;
    int32_t tl_x = -1, br_x = -1, tl_y = 0, br_y = 0;
    for (p = start, guard = 0; p + 4 <= end && guard++ < 256; ) {
        uint32_t n = psx_mod_read_word(p) >> 24;
        if (n == 0 || n > 16) { p += 4; continue; }
        for (uint32_t i = 0; i < n; i++) {
            uint32_t w = psx_mod_read_word(p + 4 + i * 4), c = w >> 24;
            if (!is_window_cmd(w, &full) || full) continue;
            if (c == 0xE5) has_window = 1;
            if (c == 0xE3) { tl_x = (int32_t)(w & 0x3FF); tl_y = (int32_t)((w >> 10) & 0x1FF); }
            if (c == 0xE4) { br_x = (int32_t)(w & 0x3FF); br_y = (int32_t)((w >> 10) & 0x1FF); }
        }
        p += 4 + n * 4;
    }
    /* The dial window packet (E3/E4/E5) is usually written by the CALLER after
     * this routine returns, so it is often not in [start,end). The needle is
     * in dial-local coords and the dial is a fixed 28x28 sprite, so fall back
     * to that size when the window is not visible here. */
    int32_t win_w, win_h;
    if (has_window && tl_x >= 0 && br_x >= tl_x) {
        win_w = squash_rel(br_x + 1, buf_origin(br_x) + 160)
              - squash_rel(tl_x, buf_origin(tl_x) + 160);
        win_h = br_y - tl_y + 1;
    } else {
        win_w = squash_rel(NFS_DIAL_SIZE, 0) - 1;
        win_h = NFS_DIAL_SIZE - 1;
    }
    for (p = start, guard = 0; p + 4 <= end && guard++ < 256; ) {
        uint32_t n = psx_mod_read_word(p) >> 24;
        if (n == 0 || n > 16) { p += 4; continue; }
        uint32_t a = p + 4, w = psx_mod_read_word(a), code = w >> 24;
        if (is_window_cmd(w, &full)) {
            for (uint32_t i = 0; i < n; i++) {
                uint32_t ai = p + 4 + i * 4, wi = psx_mod_read_word(ai);
                if (!is_window_cmd(wi, &full) || full) continue;
                if ((wi >> 24) == 0xE5) {
                    int32_t x = (int32_t)(wi & 0x7FF);
                    if (x & 0x400) x -= 0x800;
                    x = squash_rel(x, buf_origin(x) + 160);
                    psx_mod_write_word(ai, (wi & ~0x7FFu) | ((uint32_t)x & 0x7FF));
                } else {
                    int32_t x = (int32_t)(wi & 0x3FF);
                    x = squash_rel(x, buf_origin(x) + 160);
                    psx_mod_write_word(ai, (wi & ~0x3FFu) | ((uint32_t)x & 0x3FF));
                }
            }
        } else if ((code & 0xFC) == 0x20 || (code & 0xFC) == 0x40) {
            /* PolyF3 / LineF in window-relative coords: needle and shadow */
            uint32_t nv = ((code & 0xFC) == 0x20) ? 3 : 2, k;
            int small = (n >= 1 + nv);
            for (k = 0; small && k < nv; k++) {
                uint32_t v = psx_mod_read_word(a + 4 + k * 4);
                int32_t x = (int16_t)(v & 0xFFFF), y = (int16_t)(v >> 16);
                if (x < -512 || x > 512 || y < -512 || y > 512) small = 0;
            }
            if (s_dump) {
                fprintf(stderr, "[nfshs.tacho]  prim %02X small=%d", (unsigned)code, small);
                for (k = 0; k < nv; k++) {
                    uint32_t v = psx_mod_read_word(a + 4 + k * 4);
                    fprintf(stderr, " (%d,%d)", (int16_t)(v & 0xFFFF), (int16_t)(v >> 16));
                }
                fprintf(stderr, " win=%dx%d\n", (int)win_w, (int)win_h);
            }
            if (small) {
                int32_t vx[3], vy[3], base = 0, best = -1;
                for (k = 0; k < nv; k++) {
                    uint32_t v = psx_mod_read_word(a + 4 + k * 4);
                    vx[k] = squash_rel((int16_t)(v & 0xFFFF), 0);   /* about window origin */
                    vy[k] = (int16_t)(v >> 16);
                }
                if (nv == 2) {
                    /* shadow line: simply hide it in widescreen. A line whose
                     * width is >= 1024 px is skipped by the GPU (real hardware
                     * and PSXRecomp alike), so nothing is drawn. */
                    vx[0] = -1000; vx[1] = 1000; vy[0] = vy[1] = 0;
                } else {
                    /* needle triangle: the two tip vertices sit 2-3 px apart,
                     * the pivot is the vertex NOT in that closest pair. */
                    int32_t d01 = sq(vx[0]-vx[1]) + sq(vy[0]-vy[1]);
                    int32_t d02 = sq(vx[0]-vx[2]) + sq(vy[0]-vy[2]);
                    int32_t d12 = sq(vx[1]-vx[2]) + sq(vy[1]-vy[2]);
                    base = (d12 <= d01 && d12 <= d02) ? 0 : (d02 <= d01 ? 1 : 2);
                    (void)best;
                    for (k = 0; k < nv; k++) {
                        if ((int32_t)k == base) continue;
                        int32_t ex[2] = { vx[base], vx[k] }, ey[2] = { vy[base], vy[k] };
                        clip_line(ex, ey, win_w, win_h);
                        vx[k] = ex[1]; vy[k] = ey[1];
                        vx[base] = ex[0]; vy[base] = ey[0];
                    }
                }
                s_fixed++;
                if (s_dump) {
                    fprintf(stderr, "[nfshs.tacho]   -> fixed");
                    for (k = 0; k < nv; k++) fprintf(stderr, " (%d,%d)", (int)vx[k], (int)vy[k]);
                    fprintf(stderr, " @0x%08X\n", (unsigned)a);
                }
                for (k = 0; k < nv; k++) {
                    uint32_t va = a + 4 + k * 4;
                    psx_mod_write_word(va, ((uint32_t)(vy[k] & 0xFFFF) << 16) | ((uint32_t)vx[k] & 0xFFFF));
                }
            }
        }
        p += 4 + n * 4;
    }
}

static int nfshs_tacho_wrap(struct CPUState* cpu) {
    uint32_t start = psx_mod_read_word(NFS_PKT_CURSOR);
    func_override_call_original(cpu);
    uint32_t end = psx_mod_read_word(NFS_PKT_CURSOR);
    s_m = psx_mod_widescreen_x_margin();
    s_fixed = 0;
    s_dump = ((s_calls++ % 120u) == 0u);
    if (s_m > 0 && end > start && end - start < 0x1000)
        fix_tacho_packets(start, end);
    if (s_log_left > 0) {
        s_log_left--;
        fprintf(stderr, "[nfshs.tacho] margin=%d pkt 0x%08X..0x%08X needle_prims_fixed=%d\n",
                (int)s_m, (unsigned)start, (unsigned)end, s_fixed);
    }
    return 1;
}

static void nfshs_widescreen_activate(void) {
    char v[16];
    uint32_t num = 16, den = 9;
    if (psx_mod_current_option_value("aspect", v, sizeof v) && strcmp(v, "21:9") == 0) {
        num = 21; den = 9;
    }
    psx_mod_set_fixed_display_aspect(num, den);
}

static const uint32_t k_tacho_guard[4] = {
    0x27BDFFA8u, 0xAFB10034u, 0x00808821u, 0xAFBF0054u };

PSX_MOD_CONSTRUCTOR(nfshs_widescreen_register) {
    psx_mod_register_activation_plugin("nfshs.widescreen", nfshs_widescreen_activate);
    psx_mod_register_function_override("nfshs.widescreen:tacho", NFS_TACHO_FN,
                                       nfshs_tacho_wrap, k_tacho_guard, 4,
                                       FO_CREDIT_SELF);
}
