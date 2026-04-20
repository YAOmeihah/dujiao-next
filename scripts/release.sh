#!/usr/bin/env bash
set -Eeuo pipefail

RELEASE_OWNER="${RELEASE_OWNER:-YAOmeihah}"
API_REPO="${API_REPO:-dujiao-next}"
ADMIN_REPO="${ADMIN_REPO:-admin}"
USER_REPO="${USER_REPO:-user}"
API_HEALTH_URL="${API_HEALTH_URL:-http://127.0.0.1:8080/health}"
GITHUB_API_ROOT="https://api.github.com/repos"
RELEASE_TMP_DIR=""

info() {
  printf '[信息] %s\n' "$*"
}

warn() {
  printf '[警告] %s\n' "$*" >&2
}

error() {
  printf '[错误] %s\n' "$*" >&2
}

die() {
  error "$*"
  exit 1
}

cleanup_temp_dir() {
  if [[ -n "${RELEASE_TMP_DIR}" && -d "${RELEASE_TMP_DIR}" ]]; then
    rm -rf "${RELEASE_TMP_DIR}"
  fi
}

build_api_asset_name() {
  local tag="$1"
  local os_name="$2"
  local arch_name="$3"

  case "${arch_name}" in
    x86_64|amd64)
      echo "dujiao-next_${tag}_${os_name}_x86_64.tar.gz"
      ;;
    aarch64|arm64)
      echo "dujiao-next_${tag}_${os_name}_arm64.tar.gz"
      ;;
    *)
      echo ""
      return 1
      ;;
  esac
}

is_confirmed_input() {
  local answer="$1"
  [[ "${answer}" == "y" || "${answer}" == "Y" ]]
}

resolve_python_cmd() {
  if command -v python3 >/dev/null 2>&1; then
    echo "python3"
    return 0
  fi
  if command -v python >/dev/null 2>&1; then
    echo "python"
    return 0
  fi
  echo ""
  return 1
}

require_deployment_dirs() {
  local root_dir="$1"
  [[ -d "${root_dir}/admin" ]] || return 1
  [[ -d "${root_dir}/api" ]] || return 1
  [[ -d "${root_dir}/user" ]] || return 1
}

require_command() {
  local command_name="$1"
  command -v "${command_name}" >/dev/null 2>&1 || die "缺少依赖命令：${command_name}"
}

check_directory_write_access() {
  local dir_path="$1"
  local probe_path="${dir_path}/.release-write-check-$$"

  [[ -d "${dir_path}" ]] || return 1
  : > "${probe_path}" || return 1
  rm -f "${probe_path}"
}

require_deployment_write_access() {
  local root_dir="$1"

  check_directory_write_access "${root_dir}" || die "当前目录无写入权限：${root_dir}"
  check_directory_write_access "${root_dir}/admin" || die "管理端目录无写入权限：${root_dir}/admin"
  check_directory_write_access "${root_dir}/api" || die "API 目录无写入权限：${root_dir}/api"
  check_directory_write_access "${root_dir}/user" || die "用户端目录无写入权限：${root_dir}/user"
}

current_os_name() {
  case "$(uname -s)" in
    Linux)
      echo "Linux"
      ;;
    *)
      echo ""
      return 1
      ;;
  esac
}

current_arch_name() {
  case "$(uname -m)" in
    x86_64|amd64)
      echo "x86_64"
      ;;
    aarch64|arm64)
      echo "aarch64"
      ;;
    *)
      echo ""
      return 1
      ;;
  esac
}

github_release_latest_url() {
  local owner="$1"
  local repo="$2"
  echo "${GITHUB_API_ROOT}/${owner}/${repo}/releases/latest"
}

github_release_tag_url() {
  local owner="$1"
  local repo="$2"
  local tag="$3"
  echo "${GITHUB_API_ROOT}/${owner}/${repo}/releases/tags/${tag}"
}

github_api_get() {
  local url="$1"
  local output_path="$2"

  curl -fsSL \
    -H 'Accept: application/vnd.github+json' \
    -H 'X-GitHub-Api-Version: 2022-11-28' \
    "${url}" \
    -o "${output_path}"
}

check_url_reachable() {
  local url="$1"
  curl -fsSIL -L --connect-timeout 10 --retry 2 "${url}" >/dev/null
}

release_tag_from_json() {
  local json_path="$1"
  "$(resolve_python_cmd)" - "${json_path}" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    payload = json.load(handle)

print(payload.get("tag_name", ""))
PY
}

release_name_from_json() {
  local json_path="$1"
  "$(resolve_python_cmd)" - "${json_path}" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    payload = json.load(handle)

print(payload.get("name", ""))
PY
}

release_asset_url_from_json() {
  local json_path="$1"
  local asset_name="$2"
  "$(resolve_python_cmd)" - "${json_path}" "${asset_name}" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    payload = json.load(handle)

needle = sys.argv[2]
for asset in payload.get("assets", []):
    if asset.get("name") == needle:
        print(asset.get("browser_download_url", ""))
        break
PY
}

download_file() {
  local url="$1"
  local output_path="$2"

  curl -fL --retry 3 --connect-timeout 15 "${url}" -o "${output_path}"
}

locate_frontend_payload_dir() {
  local stage_dir="$1"
  local match

  if [[ -f "${stage_dir}/index.html" ]]; then
    echo "${stage_dir}"
    return 0
  fi

  if [[ -f "${stage_dir}/dist/index.html" ]]; then
    echo "${stage_dir}/dist"
    return 0
  fi

  match="$(find "${stage_dir}" -mindepth 2 -maxdepth 3 -type f -name 'index.html' -print -quit 2>/dev/null || true)"
  [[ -n "${match}" ]] || return 1
  dirname "${match}"
}

validate_frontend_stage() {
  local stage_dir="$1"
  local payload_dir

  payload_dir="$(locate_frontend_payload_dir "${stage_dir}")" || {
    error "未找到前端发布文件目录：${stage_dir}"
    return 1
  }

  [[ -f "${payload_dir}/index.html" ]] || {
    error "未找到 index.html：${payload_dir}"
    return 1
  }
}

locate_api_payload_dir() {
  local stage_dir="$1"
  local match

  if [[ -f "${stage_dir}/dujiao-next" ]]; then
    echo "${stage_dir}"
    return 0
  fi

  match="$(find "${stage_dir}" -mindepth 2 -maxdepth 3 -type f -name 'dujiao-next' -print -quit 2>/dev/null || true)"
  [[ -n "${match}" ]] || return 1
  dirname "${match}"
}

validate_api_stage() {
  local stage_dir="$1"
  local payload_dir

  payload_dir="$(locate_api_payload_dir "${stage_dir}")" || {
    error "未找到 dujiao-next 可执行文件：${stage_dir}"
    return 1
  }

  [[ -f "${payload_dir}/dujiao-next" ]] || return 1
  [[ -f "${payload_dir}/data/address_divisions/cities.json" ]] || return 1
  [[ -f "${payload_dir}/data/address_divisions/districts.json" ]] || return 1
  [[ -f "${payload_dir}/data/address_divisions/provinces.json" ]] || return 1
  [[ -f "${payload_dir}/data/address_divisions/townships.json" ]] || return 1
}

replace_directory_contents() {
  local target_dir="$1"
  local source_dir="$2"

  mkdir -p "${target_dir}"
  find "${target_dir}" -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
  cp -a "${source_dir}/." "${target_dir}/"
}

replace_api_payload() {
  local stage_dir="$1"
  local api_dir="$2"
  local payload_dir

  payload_dir="$(locate_api_payload_dir "${stage_dir}")" || die "无法定位 API 发布文件目录：${stage_dir}"

  mkdir -p "${api_dir}"
  cp -f "${payload_dir}/dujiao-next" "${api_dir}/dujiao-next"
  chmod +x "${api_dir}/dujiao-next"

  rm -rf "${api_dir}/data/address_divisions"
  mkdir -p "${api_dir}/data"
  cp -a "${payload_dir}/data/address_divisions" "${api_dir}/data/address_divisions"
}

print_release_preview() {
  local tag="$1"
  local admin_release_name="$2"
  local user_release_name="$3"
  local api_release_name="$4"
  local admin_asset_name="$5"
  local user_asset_name="$6"
  local api_asset_name="$7"

  printf '\n最新公共标签：%s\n' "${tag}"
  printf '管理端发布：%s\n' "${admin_release_name}"
  printf '管理端包：%s\n' "${admin_asset_name}"
  printf '用户端发布：%s\n' "${user_release_name}"
  printf '用户端包：%s\n' "${user_asset_name}"
  printf 'API 发布：%s\n' "${api_release_name}"
  printf 'API 包：%s\n\n' "${api_asset_name}"
}

main() {
  local root_dir temp_root temp_dir answer
  local latest_api_json admin_release_json user_release_json
  local tag admin_release_name user_release_name api_release_name
  local admin_asset_name user_asset_name api_asset_name
  local admin_asset_url user_asset_url api_asset_url
  local os_name arch_name
  local admin_archive user_archive api_archive
  local admin_extract_dir user_extract_dir api_extract_dir
  local admin_payload_dir user_payload_dir

  if [[ "$#" -gt 0 ]]; then
    die "此脚本不接受任何参数。"
  fi

  root_dir="$(pwd)"
  require_deployment_dirs "${root_dir}" || die "当前目录下必须包含 admin、api、user 三个目录。"
  require_deployment_write_access "${root_dir}"

  require_command curl
  require_command tar
  require_command unzip
  require_command find
  require_command cp
  require_command chmod
  require_command rm
  require_command mktemp
  require_command uname
  resolve_python_cmd >/dev/null || die "需要安装 python3 或 python，才能解析 GitHub 发布信息。"

  os_name="$(current_os_name)" || die "当前操作系统暂不支持，脚本目前仅支持 Linux。"
  arch_name="$(current_arch_name)" || die "当前服务器 CPU 架构暂不支持。"
  build_api_asset_name "v0.0.0" "${os_name}" "${arch_name}" >/dev/null || die "无法匹配当前服务器对应的 API 发布包。"

  temp_root="${root_dir}/.deploy/tmp"
  mkdir -p "${temp_root}"
  temp_dir="$(mktemp -d "${temp_root}/release.XXXXXX")"
  RELEASE_TMP_DIR="${temp_dir}"
  trap cleanup_temp_dir EXIT

  info "正在检查 GitHub 连通性..."
  check_url_reachable "https://api.github.com" || die "当前服务器无法访问 https://api.github.com"
  check_url_reachable "https://github.com" || die "当前服务器无法访问 https://github.com"

  latest_api_json="${temp_dir}/api-latest.json"
  github_api_get "$(github_release_latest_url "${RELEASE_OWNER}" "${API_REPO}")" "${latest_api_json}" \
    || die "无法获取 ${RELEASE_OWNER}/${API_REPO} 的最新 API 发布信息。"

  tag="$(release_tag_from_json "${latest_api_json}")"
  [[ -n "${tag}" ]] || die "GitHub API 返回结果中缺少最新发布标签。"

  api_release_name="$(release_name_from_json "${latest_api_json}")"
  admin_release_json="${temp_dir}/admin-${tag}.json"
  user_release_json="${temp_dir}/user-${tag}.json"

  github_api_get "$(github_release_tag_url "${RELEASE_OWNER}" "${ADMIN_REPO}" "${tag}")" "${admin_release_json}" \
    || die "未找到 ${RELEASE_OWNER}/${ADMIN_REPO} 的 ${tag} 发布。"
  github_api_get "$(github_release_tag_url "${RELEASE_OWNER}" "${USER_REPO}" "${tag}")" "${user_release_json}" \
    || die "未找到 ${RELEASE_OWNER}/${USER_REPO} 的 ${tag} 发布。"

  admin_release_name="$(release_name_from_json "${admin_release_json}")"
  user_release_name="$(release_name_from_json "${user_release_json}")"

  admin_asset_name="dujiao-next-admin-${tag}.zip"
  user_asset_name="dujiao-next-user-${tag}.zip"
  api_asset_name="$(build_api_asset_name "${tag}" "${os_name}" "${arch_name}")" || die "无法生成适用于 ${os_name}/${arch_name} 的 API 包名。"

  admin_asset_url="$(release_asset_url_from_json "${admin_release_json}" "${admin_asset_name}")"
  user_asset_url="$(release_asset_url_from_json "${user_release_json}" "${user_asset_name}")"
  api_asset_url="$(release_asset_url_from_json "${latest_api_json}" "${api_asset_name}")"

  [[ -n "${admin_asset_url}" ]] || die "在 ${RELEASE_OWNER}/${ADMIN_REPO} 的 ${tag} 发布中未找到包：${admin_asset_name}"
  [[ -n "${user_asset_url}" ]] || die "在 ${RELEASE_OWNER}/${USER_REPO} 的 ${tag} 发布中未找到包：${user_asset_name}"
  [[ -n "${api_asset_url}" ]] || die "在 ${RELEASE_OWNER}/${API_REPO} 的 ${tag} 发布中未找到包：${api_asset_name}"

  info "正在检查发布包是否可访问..."
  check_url_reachable "${admin_asset_url}" || die "无法访问发布包：${admin_asset_name}"
  check_url_reachable "${user_asset_url}" || die "无法访问发布包：${user_asset_name}"
  check_url_reachable "${api_asset_url}" || die "无法访问发布包：${api_asset_name}"

  print_release_preview \
    "${tag}" \
    "${admin_release_name}" \
    "${user_release_name}" \
    "${api_release_name}" \
    "${admin_asset_name}" \
    "${user_asset_name}" \
    "${api_asset_name}"

  read -r -p "确认继续部署吗？[y/N]: " answer
  if ! is_confirmed_input "${answer}"; then
    info "已取消。"
    return 0
  fi

  admin_archive="${temp_dir}/${admin_asset_name}"
  user_archive="${temp_dir}/${user_asset_name}"
  api_archive="${temp_dir}/${api_asset_name}"
  admin_extract_dir="${temp_dir}/admin"
  user_extract_dir="${temp_dir}/user"
  api_extract_dir="${temp_dir}/api"

  mkdir -p "${admin_extract_dir}" "${user_extract_dir}" "${api_extract_dir}"

  info "正在下载管理端发布包..."
  download_file "${admin_asset_url}" "${admin_archive}"
  info "正在下载用户端发布包..."
  download_file "${user_asset_url}" "${user_archive}"
  info "正在下载 API 发布包..."
  download_file "${api_asset_url}" "${api_archive}"

  info "正在解压管理端发布包..."
  unzip -qq "${admin_archive}" -d "${admin_extract_dir}"
  info "正在解压用户端发布包..."
  unzip -qq "${user_archive}" -d "${user_extract_dir}"
  info "正在解压 API 发布包..."
  tar -xzf "${api_archive}" -C "${api_extract_dir}"

  validate_frontend_stage "${admin_extract_dir}" || die "管理端发布包校验失败。"
  validate_frontend_stage "${user_extract_dir}" || die "用户端发布包校验失败。"
  validate_api_stage "${api_extract_dir}" || die "API 发布包校验失败。"

  admin_payload_dir="$(locate_frontend_payload_dir "${admin_extract_dir}")"
  user_payload_dir="$(locate_frontend_payload_dir "${user_extract_dir}")"

  info "正在替换管理端文件..."
  replace_directory_contents "${root_dir}/admin" "${admin_payload_dir}"
  info "正在替换用户端文件..."
  replace_directory_contents "${root_dir}/user" "${user_payload_dir}"
  info "正在替换 API 发布文件..."
  replace_api_payload "${api_extract_dir}" "${root_dir}/api"

  info "发布文件已更新完成。"
  warn "API 文件已经更新，但当前运行中的后端进程仍在使用旧版本。"
  warn "请前往宝塔进程管理器手动重启 API 进程，然后访问 ${API_HEALTH_URL} 验证。"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
