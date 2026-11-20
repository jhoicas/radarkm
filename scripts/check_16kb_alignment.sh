#!/usr/bin/env bash
# Validates ELF LOAD segment alignment for RadarKm .so files (16 KB = 0x4000).
set -euo pipefail

ROOT="${1:-}"
if [[ -z "${ROOT}" ]]; then
  echo "Usage: $0 <path-to-search-for-.so>"
  exit 2
fi

if [[ ! -d "${ROOT}" && ! -f "${ROOT}" ]]; then
  echo "Path not found: ${ROOT}"
  exit 2
fi

if ! command -v llvm-readelf >/dev/null 2>&1 && ! command -v readelf >/dev/null 2>&1; then
  echo "llvm-readelf or readelf is required"
  exit 2
fi

READELF_BIN="llvm-readelf"
if ! command -v llvm-readelf >/dev/null 2>&1; then
  READELF_BIN="readelf"
fi

FAIL=0
FOUND=0
while IFS= read -r -d '' SO; do
  FOUND=1
  echo "==> Checking ${SO}"
  HEADERS="$("${READELF_BIN}" -l "${SO}" 2>/dev/null || true)"
  # llvm-readelf prints Align as the last column of LOAD rows (e.g. "... R E 0x4000").
  LOAD_ALIGNS="$(printf '%s\n' "${HEADERS}" | awk '/LOAD/ { print $NF }')"
  if [[ -z "${LOAD_ALIGNS}" ]]; then
    echo "FAIL: no LOAD segments found"
    printf '%s\n' "${HEADERS}"
    FAIL=1
    continue
  fi

  BAD=0
  while IFS= read -r ALIGN; do
    [[ -z "${ALIGN}" ]] && continue
    case "${ALIGN}" in
      0x4000|0x8000|0x10000|16384|32768|65536)
        ;;
      *)
        # Numeric hex below 0x4000 fails
        BAD=1
        echo "  LOAD Align=${ALIGN} (< 0x4000)"
        ;;
    esac
  done <<< "${LOAD_ALIGNS}"

  if [[ ${BAD} -eq 0 ]]; then
    echo "OK: all LOAD segments Align >= 0x4000 (16KB)"
    printf '%s\n' "${LOAD_ALIGNS}" | sed 's/^/  Align /'
  else
    echo "FAIL: one or more LOAD segments are below 16KB alignment"
    printf '%s\n' "${HEADERS}"
    FAIL=1
  fi
done < <(find "${ROOT}" -type f \( -name 'libradarkm_security.so' -o -name 'libradarkm_*.so' \) -print0)

if [[ ${FOUND} -eq 0 ]]; then
  echo "No libradarkm_*.so files found under ${ROOT}"
  exit 2
fi

if [[ ${FAIL} -ne 0 ]]; then
  echo "One or more libraries failed 16KB alignment check"
  exit 1
fi

echo "All discovered RadarKm .so files passed 16KB LOAD alignment check"
