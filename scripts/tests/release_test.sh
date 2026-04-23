#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source "${ROOT_DIR}/scripts/release.sh"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_eq() {
  local expected="$1"
  local actual="$2"
  local message="$3"
  [[ "$expected" == "$actual" ]] || fail "${message}: expected='${expected}' actual='${actual}'"
}

test_detect_linux_asset_name() {
  assert_eq "dujiao-next_v1.2.3_Linux_x86_64.tar.gz" \
    "$(build_api_asset_name "v1.2.3" "Linux" "x86_64")" \
    "x86_64 asset name"
  assert_eq "dujiao-next_v1.2.3_Linux_arm64.tar.gz" \
    "$(build_api_asset_name "v1.2.3" "Linux" "aarch64")" \
    "arm64 asset name"
}

test_confirmation_input() {
  assert_eq "0" "$(is_confirmed_input "y"; echo $?)" "lowercase y accepted"
  assert_eq "0" "$(is_confirmed_input "Y"; echo $?)" "uppercase Y accepted"
}

test_resolve_python_command() {
  local python_cmd
  python_cmd="$(resolve_python_cmd)"
  [[ -n "${python_cmd}" ]] || fail "python command should not be empty"
}

test_require_deployment_dirs() {
  local workdir
  workdir="$(mktemp -d)"
  mkdir -p "${workdir}/admin" "${workdir}/api" "${workdir}/user"
  require_deployment_dirs "${workdir}"
  rm -rf "${workdir}"
}

test_normalize_release_target_selection() {
  assert_eq "admin" "$(normalize_release_target_selection "1")" "admin selection"
  assert_eq "user" "$(normalize_release_target_selection "2")" "user selection"
  assert_eq "api" "$(normalize_release_target_selection "3")" "api selection"
  assert_eq "all" "$(normalize_release_target_selection "4")" "all selection"
  assert_eq "1" "$(normalize_release_target_selection "9" >/dev/null 2>&1; echo $?)" "invalid selection rejected"
}

test_release_target_includes() {
  assert_eq "0" "$(release_target_includes "admin" "admin"; echo $?)" "admin includes admin"
  assert_eq "1" "$(release_target_includes "admin" "user" >/dev/null 2>&1; echo $?)" "admin excludes user"
  assert_eq "0" "$(release_target_includes "all" "admin"; echo $?)" "all includes admin"
  assert_eq "0" "$(release_target_includes "all" "user"; echo $?)" "all includes user"
  assert_eq "0" "$(release_target_includes "all" "api"; echo $?)" "all includes api"
}

test_require_selected_deployment_dirs() {
  local workdir
  workdir="$(mktemp -d)"

  mkdir -p "${workdir}/admin"
  require_selected_deployment_dirs "${workdir}" "admin"
  assert_eq "1" "$(require_selected_deployment_dirs "${workdir}" "all" >/dev/null 2>&1; echo $?)" "all requires all directories"

  rm -rf "${workdir}"
}

make_release_fixture() {
  local path="$1"
  cat >"${path}" <<'JSON'
{
  "tag_name": "v1.0.3-fork.2",
  "name": "v1.0.3-fork.2",
  "assets": [
    {
      "name": "dujiao-next-admin-v1.0.3-fork.2.zip",
      "browser_download_url": "https://example.test/admin.zip"
    },
    {
      "name": "dujiao-next-user-v1.0.3-fork.2.zip",
      "browser_download_url": "https://example.test/user.zip"
    },
    {
      "name": "dujiao-next_v1.0.3-fork.2_Linux_x86_64.tar.gz",
      "browser_download_url": "https://example.test/api.tar.gz"
    }
  ]
}
JSON
}

test_parse_release_metadata() {
  local temp_json
  temp_json="$(mktemp)"
  make_release_fixture "${temp_json}"

  assert_eq "v1.0.3-fork.2" "$(release_tag_from_json "${temp_json}")" "tag parsing"
  assert_eq "v1.0.3-fork.2" "$(release_name_from_json "${temp_json}")" "name parsing"
  assert_eq "https://example.test/admin.zip" \
    "$(release_asset_url_from_json "${temp_json}" "dujiao-next-admin-v1.0.3-fork.2.zip")" \
    "admin asset url"

  rm -f "${temp_json}"
}

make_frontend_stage() {
  local stage_dir="$1"
  mkdir -p "${stage_dir}/dist"
  printf '<!doctype html>' > "${stage_dir}/dist/index.html"
  printf 'console.log(1);' > "${stage_dir}/dist/app.js"
}

make_api_stage() {
  local stage_dir="$1"
  mkdir -p "${stage_dir}/data/address_divisions"
  printf '#!/usr/bin/env bash\necho ok\n' > "${stage_dir}/dujiao-next"
  chmod +x "${stage_dir}/dujiao-next"
  printf '[]' > "${stage_dir}/data/address_divisions/cities.json"
  printf '[]' > "${stage_dir}/data/address_divisions/districts.json"
  printf '[]' > "${stage_dir}/data/address_divisions/provinces.json"
  printf '[]' > "${stage_dir}/data/address_divisions/townships.json"
}

test_validate_frontend_stage() {
  local temp_dir
  temp_dir="$(mktemp -d)"
  make_frontend_stage "${temp_dir}"
  validate_frontend_stage "${temp_dir}"
  rm -rf "${temp_dir}"
}

test_validate_api_stage() {
  local temp_dir
  temp_dir="$(mktemp -d)"
  make_api_stage "${temp_dir}"
  validate_api_stage "${temp_dir}"
  rm -rf "${temp_dir}"
}

test_replace_directory_contents() {
  local target_dir source_dir
  target_dir="$(mktemp -d)"
  source_dir="$(mktemp -d)"
  printf 'old' > "${target_dir}/old.txt"
  printf 'new' > "${source_dir}/new.txt"
  replace_directory_contents "${target_dir}" "${source_dir}"
  [[ -f "${target_dir}/new.txt" ]] || fail "new file should exist"
  [[ ! -f "${target_dir}/old.txt" ]] || fail "old file should be removed"
  rm -rf "${target_dir}" "${source_dir}"
}

main() {
  test_detect_linux_asset_name
  test_confirmation_input
  test_resolve_python_command
  test_require_deployment_dirs
  test_normalize_release_target_selection
  test_release_target_includes
  test_require_selected_deployment_dirs
  test_parse_release_metadata
  test_validate_frontend_stage
  test_validate_api_stage
  test_replace_directory_contents
  echo "PASS"
}

main "$@"
