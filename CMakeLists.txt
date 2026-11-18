cmake_minimum_required(VERSION 3.20)
set(PSXRECOMP_ROOT "${CMAKE_CURRENT_SOURCE_DIR}/psxrecomp" CACHE PATH
    "Path to psxrecomp")
if(NOT EXISTS "${PSXRECOMP_ROOT}/runtime/runtime.cmake")
    message(FATAL_ERROR
        "PSXRecomp runtime is missing: ${PSXRECOMP_ROOT}/runtime/runtime.cmake\n"
        "This project expects the psxrecomp framework checkout at ${PSXRECOMP_ROOT}.\n"
        "Run: git submodule update --init --recursive\n"
        "If psxrecomp/ is not a submodule, set -DPSXRECOMP_ROOT=<path-to-psxrecomp>.")
endif()

project(NFS_High_Stakes_Recomp LANGUAGES C CXX)

# --- Opt-in product surfaces (OFF by default in the framework) -------------
# Scaffold prompts / flags fill these. Without recomp-ui, PSX_RECOMP_UI is OFF.
# The wizard (WIZARD_CMAKE_BLOCK / ENABLE_SETUP_WIZARD) is the first-run disc
# picker; releases are compiled from the committed generated/ C, so it never
# needs to generate or rebuild anything on a player's machine.
# recomp-ui submodule present (PSX_RECOMP_UI defaults ON).
if(EXISTS "${PSXRECOMP_ROOT}/lib/recomp-net/CMakeLists.txt")
    set(PSX_NETPLAY ON CACHE BOOL
        "Link recomp-net delay-sync (opt-in; needs recomp-net)" FORCE)
endif()
set(PSX_SETUP_WIZARD ON CACHE BOOL
    "Advertise first-run setup wizard + Generate & rebuild in recomp-ui" FORCE)
include("${PSXRECOMP_ROOT}/runtime/runtime.cmake")

# This project's own mod catalog. The FRAMEWORK stages it -- together with its
# builtin psx.* packages -- into <exe-dir>/mods/bundled, and refuses to
# configure if a catalog exists here but is not declared. Shared Internal
# Resolution is included even with a builtin allowlist, enabled at 4x on fresh
# profiles, with 1x..32x choices; do not supply a title-local copy of that id.
# Never copy the catalog by hand: five titles did, and framework 4cc04be3's rename of the staged
# directory broke all five silently (bead beads-eio.3.101).
set(PSX_PRELOADED_MODS "${CMAKE_CURRENT_SOURCE_DIR}/mods/preloaded")

# High Speed 4: el paquete de distribucion (-DHS4_DISTRIBUCION=ON, lo pone packaging/release.toml) NO lleva
# el mod de desbloqueos (nfshs.unlocks) ni el de 4 jugadores (nfshs.4p): ni su codigo ni su paquete.
# El catalogo de distribucion se genera aqui desde mods/preloaded (no hay copias a mano que se desincronicen).
option(HS4_DISTRIBUCION "Build de distribucion de High Speed 4 (sin desbloqueos ni 4 jugadores)" OFF)
set(HS4_MODS_DISTRIBUCION nfshs.widescreen nfshs.jukebox)
if(HS4_DISTRIBUCION)
    set(_hs4_mods "${CMAKE_BINARY_DIR}/hs4_mods_distribucion")
    file(REMOVE_RECURSE "${_hs4_mods}")
    file(MAKE_DIRECTORY "${_hs4_mods}/packages")
    foreach(_hs4_pkg IN LISTS HS4_MODS_DISTRIBUCION)
        file(COPY "${CMAKE_CURRENT_SOURCE_DIR}/mods/preloaded/packages/${_hs4_pkg}"
             DESTINATION "${_hs4_mods}/packages")
    endforeach()
    set(PSX_PRELOADED_MODS "${_hs4_mods}")
    message(STATUS "High Speed 4 distribucion: mods ${HS4_MODS_DISTRIBUCION} (sin nfshs.unlocks ni nfshs.4p)")
    # Las versiones publicas de PSXRecomp no pueden llevar los mods de carga rapida del framework.
    set(_hs4_excluir EXCLUDE_BUILTIN_MODS psx.enhancement.cd-speed psx.enhancement.fast-loading)
else()
    set(_hs4_excluir "")
endif()
if(NOT IS_DIRECTORY "${PSX_PRELOADED_MODS}")
    set(PSX_PRELOADED_MODS "NONE")
endif()

psxrecomp_add_game_runtime(psx-runtime
    PRELOADED_MODS_DIR "${PSX_PRELOADED_MODS}"
    ENABLE_NETPLAY_IF_PRESENT
    NETPLAY_LOBBY_URL "ws://netplay.retcomm.net:8765"
    ENABLE_SETUP_WIZARD
    WINDOW_TITLE "High Speed 4"
    # El .exe de desarrollo conserva su nombre (lo usan los .bat de prueba); el paquete lo copia como HighSpeed4.exe
    EXE_NAME "NFS_High_Stakes__Recompiled"
    DEFAULT_GAME_CONFIG_PATH "game.toml"
    MAX_PLAYERS 2
    LAUNCHER_BOXART "${CMAKE_CURRENT_SOURCE_DIR}/launcher_assets/img/boxart.tga"
    APP_ICON "${CMAKE_CURRENT_SOURCE_DIR}/assets/psxrecomp.ico"
    # Committed game C (psxrecomp_cli.py generate; git add generated). Release
    # CI configures with -DPSXRECOMP_REQUIRE_GAME_C=ON so a checkout without
    # it fails instead of building a setup host.
    GEN_MARKER "generated/SLUS_008.26_dispatch.c"
    GEN_FULL_GLOB "generated/SLUS_008.26_full_*.c"
    CODEGEN_SETUP_SOURCES "${CMAKE_CURRENT_SOURCE_DIR}/codegen_setup.c"
    ${_hs4_excluir}
)

# Stage runtime configs next to the exe for direct build/ launches. Only copy
# files that exist — never fail POST_BUILD when game_options.toml is absent.
foreach(_PSX_RUNTIME_CONFIG game.toml game_options.toml)
    if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/${_PSX_RUNTIME_CONFIG}")
        add_custom_command(TARGET psx-runtime POST_BUILD
            COMMAND ${CMAKE_COMMAND} -E copy_if_different
                "${CMAKE_CURRENT_SOURCE_DIR}/${_PSX_RUNTIME_CONFIG}"
                "$<TARGET_FILE_DIR:psx-runtime>/${_PSX_RUNTIME_CONFIG}"
            COMMENT "Staging ${_PSX_RUNTIME_CONFIG}"
            VERBATIM)
    endif()
endforeach()

# Game-owned trusted mod plugins (activated by mods/preloaded manifests).
target_sources(psx-runtime PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/src/mods/nfshs_widescreen.c"
    "${CMAKE_CURRENT_SOURCE_DIR}/src/mods/nfshs_jukebox.c")
if(NOT HS4_DISTRIBUCION)
    # Solo en desarrollo: desbloqueos y 4 jugadores (en pausa).
    # No se publican en GitHub: solo se compilan si existen en esta carpeta.
    foreach(_hs4_dev nfshs_unlocks nfshs_4p)
        if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/src/mods/${_hs4_dev}.c")
            target_sources(psx-runtime PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/src/mods/${_hs4_dev}.c")
        endif()
    endforeach()
endif()
