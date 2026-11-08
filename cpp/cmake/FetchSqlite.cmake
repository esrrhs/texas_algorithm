# Downloads the pinned SQLite amalgamation archive on behalf of the parent
# build. The parent runs it as a separate CMake process:
#
#   cmake -DDL_URL=<url> -DDL_DEST=<file> -DDL_SHA256=<hash> -P FetchSqlite.cmake
#
# Any failure exits non-zero; the parent then only warns and disables SQLite
# support instead of aborting the configure step.

if(NOT DEFINED DL_URL OR NOT DEFINED DL_DEST OR NOT DEFINED DL_SHA256)
  message(FATAL_ERROR "FetchSqlite.cmake needs DL_URL, DL_DEST and DL_SHA256")
endif()

file(DOWNLOAD "${DL_URL}" "${DL_DEST}"
     EXPECTED_HASH SHA256=${DL_SHA256}
     TLS_VERIFY ON
     TIMEOUT 60
     STATUS dl_status
     LOG dl_log)
list(GET dl_status 0 dl_code)
list(GET dl_status 1 dl_message)
if(NOT dl_code EQUAL 0)
  file(REMOVE "${DL_DEST}")
  message(FATAL_ERROR "download failed: ${dl_message}")
endif()
