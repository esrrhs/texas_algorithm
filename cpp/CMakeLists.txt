cmake_minimum_required(VERSION 3.16)
project(texas_algorithm C CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Threads REQUIRED)

# zstddeclib.c is the official decompression-only single-file build of zstd. It
# is vendored here so users never need a system libzstd: cloning and building
# is enough. Regenerate it with zstd's build/single_file_libs/create_single_file_decoder.sh.
add_library(zstddeclib STATIC third_party/zstd/zstddeclib.c)
target_include_directories(zstddeclib PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/third_party/zstd)

# SQLite is optional. TEXAS_SQLITE picks how the library is obtained:
#   auto   - use the system SQLite when found, otherwise download the pinned
#            amalgamation (default)
#   system - only the system SQLite; disable the feature when it is missing
#   fetch  - always download the pinned amalgamation
#   off    - no SQLite support at all
# Loading a packed database is a convenience, never a requirement, so a failure
# to find or download SQLite only warns and turns the feature off; the build and
# every other test still succeed.
set(TEXAS_SQLITE "auto" CACHE STRING
    "SQLite support: auto (system, else download), system, fetch, off")
set_property(CACHE TEXAS_SQLITE PROPERTY STRINGS auto system fetch off)
if(NOT TEXAS_SQLITE MATCHES "^(auto|system|fetch|off)$")
  message(FATAL_ERROR
      "TEXAS_SQLITE must be auto, system, fetch or off (got '${TEXAS_SQLITE}')")
endif()

# Pinned upstream release, so the fetched source is reproducible and verifiable.
# Bump the version only together with the file name and the SHA256.
set(TEXAS_SQLITE_VERSION "3.45.1")
set(TEXAS_SQLITE_ZIP "sqlite-amalgamation-3450100.zip")
set(TEXAS_SQLITE_URL "https://www.sqlite.org/2024/${TEXAS_SQLITE_ZIP}")
set(TEXAS_SQLITE_SHA256 "5592243caf28b2cdef41e6ab58d25d653dfc53deded8450eb66072c929f030c4")

set(TEXAS_SQLITE_MODE "off")
set(TEXAS_SQLITE_SOURCE_DIR "")
set(TEXAS_SQLITE_WANT_DOWNLOAD FALSE)

if(TEXAS_SQLITE STREQUAL "fetch")
  set(TEXAS_SQLITE_WANT_DOWNLOAD TRUE)
elseif(NOT TEXAS_SQLITE STREQUAL "off")
  # auto and system both prefer the system library; auto falls back to a
  # download while system gives up.
  find_package(SQLite3 QUIET)
  if(SQLite3_FOUND)
    set(TEXAS_SQLITE_MODE "system")
  elseif(TEXAS_SQLITE STREQUAL "auto")
    set(TEXAS_SQLITE_WANT_DOWNLOAD TRUE)
  endif()
endif()

if(TEXAS_SQLITE_WANT_DOWNLOAD)
  # Download into the build tree first, so a network failure can be handled
  # here instead of aborting the configure step inside FetchContent.
  set(TEXAS_SQLITE_CACHE_DIR "${CMAKE_BINARY_DIR}/_deps")
  set(TEXAS_SQLITE_CACHE "${TEXAS_SQLITE_CACHE_DIR}/${TEXAS_SQLITE_ZIP}")
  file(MAKE_DIRECTORY "${TEXAS_SQLITE_CACHE_DIR}")
  set(TEXAS_SQLITE_DL_CODE 1)
  if(EXISTS "${TEXAS_SQLITE_CACHE}")
    # Reuse a previous download only when it still matches the pinned hash.
    file(SHA256 "${TEXAS_SQLITE_CACHE}" TEXAS_SQLITE_CACHE_HASH)
    if(TEXAS_SQLITE_CACHE_HASH STREQUAL TEXAS_SQLITE_SHA256)
      set(TEXAS_SQLITE_DL_CODE 0)
    else()
      file(REMOVE "${TEXAS_SQLITE_CACHE}")
    endif()
  endif()
  if(NOT TEXAS_SQLITE_DL_CODE EQUAL 0)
    message(STATUS "SQLite not found locally; downloading ${TEXAS_SQLITE_URL}")
    # Run the download in a separate CMake process: even a hard failure there
    # (unwritable cache, no network) stays a warning instead of aborting the
    # configure step.
    execute_process(
        COMMAND ${CMAKE_COMMAND}
                "-DDL_URL=${TEXAS_SQLITE_URL}"
                "-DDL_DEST=${TEXAS_SQLITE_CACHE}"
                "-DDL_SHA256=${TEXAS_SQLITE_SHA256}"
                -P "${CMAKE_CURRENT_SOURCE_DIR}/cmake/FetchSqlite.cmake"
        RESULT_VARIABLE TEXAS_SQLITE_DL_CODE
        OUTPUT_VARIABLE TEXAS_SQLITE_DL_OUTPUT
        ERROR_VARIABLE TEXAS_SQLITE_DL_ERROR)
  endif()
  if(TEXAS_SQLITE_DL_CODE EQUAL 0 AND EXISTS "${TEXAS_SQLITE_CACHE}")
    include(FetchContent)
    FetchContent_Declare(sqlite_amalgamation
        URL "${TEXAS_SQLITE_CACHE}"
        URL_HASH SHA256=${TEXAS_SQLITE_SHA256}
        DOWNLOAD_EXTRACT_TIMESTAMP TRUE)
    FetchContent_MakeAvailable(sqlite_amalgamation)
    set(TEXAS_SQLITE_SOURCE_DIR "${sqlite_amalgamation_SOURCE_DIR}")
    # The zip keeps everything under sqlite-amalgamation-<version>/.
    if(NOT EXISTS "${TEXAS_SQLITE_SOURCE_DIR}/sqlite3.c")
      set(TEXAS_SQLITE_SOURCE_DIR
          "${TEXAS_SQLITE_SOURCE_DIR}/sqlite-amalgamation-3450100")
    endif()
    if(EXISTS "${TEXAS_SQLITE_SOURCE_DIR}/sqlite3.c")
      set(TEXAS_SQLITE_MODE "fetch")
    else()
      # A stale or foreign archive: drop it so the next configure retries.
      message(WARNING "Downloaded SQLite archive did not contain sqlite3.c; removed it and disabled SQLite support.")
      file(REMOVE "${TEXAS_SQLITE_CACHE}")
      set(TEXAS_SQLITE_MODE "off")
    endif()
  else()
    message(STATUS "SQLite download failed: ${TEXAS_SQLITE_DL_ERROR}")
    file(REMOVE "${TEXAS_SQLITE_CACHE}")
    set(TEXAS_SQLITE_MODE "off")
  endif()
endif()

if(TEXAS_SQLITE_MODE STREQUAL "system")
  message(STATUS "SQLite support: system library")
  # FindSQLite3 exports SQLite3::SQLite3 in current CMake and SQLite::SQLite3
  # in older releases; accept whichever this one provides.
  if(TARGET SQLite3::SQLite3)
    set(TEXAS_SQLITE_TARGET SQLite3::SQLite3)
  else()
    set(TEXAS_SQLITE_TARGET SQLite::SQLite3)
  endif()
  add_library(texas_sqlite INTERFACE)
  target_link_libraries(texas_sqlite INTERFACE ${TEXAS_SQLITE_TARGET})
  target_compile_definitions(texas_sqlite INTERFACE TEXAS_HAVE_SQLITE=1)
elseif(TEXAS_SQLITE_MODE STREQUAL "fetch")
  message(STATUS "SQLite support: downloaded ${TEXAS_SQLITE_VERSION} amalgamation")
  # Load-extension support is compiled out because the packed database is
  # opened read-only and never loads code.
  add_library(texas_sqlite STATIC "${TEXAS_SQLITE_SOURCE_DIR}/sqlite3.c")
  target_include_directories(texas_sqlite PUBLIC "${TEXAS_SQLITE_SOURCE_DIR}")
  target_compile_definitions(texas_sqlite PUBLIC TEXAS_HAVE_SQLITE=1)
  target_compile_definitions(texas_sqlite PRIVATE
      SQLITE_THREADSAFE=1
      SQLITE_OMIT_LOAD_EXTENSION=1
      SQLITE_OMIT_DEPRECATED=1)
else()
  if(TEXAS_SQLITE STREQUAL "off")
    message(STATUS "SQLite support: disabled (TEXAS_SQLITE=off)")
  elseif(TEXAS_SQLITE STREQUAL "system")
    message(WARNING
        "TEXAS_SQLITE=system but no system SQLite was found; "
        "SQLite support is disabled and the build continues. "
        "Install libsqlite3 (apt install libsqlite3-dev / brew install sqlite) "
        "or reconfigure with -DTEXAS_SQLITE=fetch.")
  else()
    message(WARNING
        "SQLite is unavailable: no system library and the download failed; "
        "SQLite support is disabled and the build continues. "
        "Install a system libsqlite3, or reconfigure with -DTEXAS_SQLITE=fetch "
        "while online.")
  endif()
  add_library(texas_sqlite INTERFACE)
  target_compile_definitions(texas_sqlite INTERFACE TEXAS_HAVE_SQLITE=0)
endif()

add_library(texas_algorithm
    poke.cpp
    texas_card_util.cpp
    gen_util.cpp
    sorter.cpp
    gen_opt_util.cpp
    gen_extra_util.cpp
    gen_extra_opt_util.cpp
    gen_trans_util.cpp
    gen_trans_opt_util.cpp
    gen_hand_util.cpp
    texas_algorithm_util.cpp
)
target_include_directories(texas_algorithm PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(texas_algorithm PUBLIC Threads::Threads zstddeclib texas_sqlite)

add_executable(texas_algorithm_demo cmd/texas_algorithm/main.cpp)
target_link_libraries(texas_algorithm_demo PRIVATE texas_algorithm)

enable_testing()

add_executable(texas_card_util_test tests/texas_card_util_test.cpp)
target_link_libraries(texas_card_util_test PRIVATE texas_algorithm)
# The data tables are committed once at the repo root, not copied per language.
add_test(NAME texas_card_util_test COMMAND texas_card_util_test
         WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/..)

add_executable(texas_algorithm_util_test tests/texas_algorithm_util_test.cpp)
target_link_libraries(texas_algorithm_util_test PRIVATE texas_algorithm)
add_test(NAME texas_algorithm_util_test COMMAND texas_algorithm_util_test
         WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/..)
