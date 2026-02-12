#!/bin/sh
set -eu

# opencc installer - download and install from GitHub Releases
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall

REPO="dopejs/opencc"
BIN_TARGET="/usr/local/bin/opencc"
CC_ENVS_DIR="$HOME/.cc_envs"

# --- Helpers ---

info()  { printf "\033[1;34m==>\033[0m %s\n" "$1"; }
ok()    { printf "\033[1;32m==>\033[0m %s\n" "$1"; }
err()   { printf "\033[1;31mError:\033[0m %s\n" "$1" >&2; }

need_sudo() {
  if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
      echo "sudo"
    else
      err "Need root privileges. Run with sudo or as root."
      exit 1
    fi
  fi
}

detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
  esac

  case "$OS" in
    darwin|linux) ;;
    *) err "Unsupported OS: $OS"; exit 1 ;;
  esac

  info "Detected platform: ${OS}/${ARCH}"
}

get_latest_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null
  else
    err "curl or wget is required"
    exit 1
  fi
}

fetch() {
  _url="$1"
  _out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fL --progress-bar -o "$_out" "$_url"
  elif command -v wget >/dev/null 2>&1; then
    wget --show-progress -qO "$_out" "$_url" 2>&1
  fi
}

# --- Completion install paths ---

zsh_comp_dir() {
  if command -v zsh >/dev/null 2>&1; then
    _fpath_list="$(zsh -ic 'for d in $fpath; do echo "$d"; done' 2>/dev/null)"
    for _d in $_fpath_list; do
      case "$_d" in
        /usr/share/*) continue ;;
        */site-functions)
          echo "$_d"
          return 0
          ;;
      esac
    done
  fi
  echo "/usr/local/share/zsh/site-functions"
}

bash_comp_dir() {
  if [ "$(uname -s)" = "Darwin" ]; then
    if [ -d "/opt/homebrew/etc/bash_completion.d" ]; then
      echo "/opt/homebrew/etc/bash_completion.d"
    elif [ -d "/usr/local/etc/bash_completion.d" ]; then
      echo "/usr/local/etc/bash_completion.d"
    fi
  else
    if [ -d "/usr/share/bash-completion/completions" ]; then
      echo "/usr/share/bash-completion/completions"
    elif [ -d "/etc/bash_completion.d" ]; then
      echo "/etc/bash_completion.d"
    fi
  fi
}

fish_comp_dir() {
  echo "$HOME/.config/fish/completions"
}

# --- First provider setup ---

setup_first_provider() {
  printf "\n"
  info "No providers configured yet. Let's set up your first one."
  printf "  (Press Ctrl+C to skip and configure later with 'opencc config')\n\n"

  # Use /dev/tty for input since stdin may be a pipe
  if [ ! -t 0 ] && [ -e /dev/tty ]; then
    _input="/dev/tty"
  else
    _input="/dev/stdin"
  fi

  printf "  Provider name (e.g. work, personal): "
  read -r _name < "$_input"
  _name="$(printf '%s' "$_name" | tr -d '[:space:]')"
  if [ -z "$_name" ]; then
    err "Name cannot be empty. Run 'opencc config' to set up later."
    return
  fi

  printf "  ANTHROPIC_BASE_URL: "
  read -r _base_url < "$_input"
  _base_url="$(printf '%s' "$_base_url" | tr -d '[:space:]')"
  if [ -z "$_base_url" ]; then
    err "Base URL cannot be empty. Run 'opencc config' to set up later."
    return
  fi

  printf "  ANTHROPIC_AUTH_TOKEN: "
  read -r _token < "$_input"
  _token="$(printf '%s' "$_token" | tr -d '[:space:]')"
  if [ -z "$_token" ]; then
    err "Auth token cannot be empty. Run 'opencc config' to set up later."
    return
  fi

  printf "  ANTHROPIC_MODEL (leave empty for default): "
  read -r _model < "$_input"
  _model="$(printf '%s' "$_model" | tr -d '[:space:]')"

  # Write .env file
  _env_path="$CC_ENVS_DIR/${_name}.env"
  printf "ANTHROPIC_BASE_URL=%s\n" "$_base_url" > "$_env_path"
  printf "ANTHROPIC_AUTH_TOKEN=%s\n" "$_token" >> "$_env_path"
  if [ -n "$_model" ]; then
    printf "ANTHROPIC_MODEL=%s\n" "$_model" >> "$_env_path"
  fi

  # Write to fallback.conf
  printf "%s\n" "$_name" > "$CC_ENVS_DIR/fallback.conf"

  printf "\n"
  ok "Provider '${_name}' created and set as default fallback."
  printf "  Run 'opencc' to start with this provider\n"
  printf "  Run 'opencc config' to add more providers\n"
}

# --- Install ---

do_install() {
  detect_platform
  SUDO="$(need_sudo)"

  info "Fetching latest release info..."
  RELEASE_JSON="$(get_latest_version)"

  # Extract version tag (simple grep, no jq dependency)
  VERSION="$(printf '%s' "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')"
  if [ -z "$VERSION" ]; then
    err "Failed to determine latest version"
    exit 1
  fi
  info "Latest version: ${VERSION}"

  # Download binary
  ASSET_NAME="opencc-${OS}-${ARCH}"
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"

  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  info "Downloading ${DOWNLOAD_URL}..."
  fetch "$DOWNLOAD_URL" "${TMPDIR}/opencc"

  # Install binary
  chmod +x "${TMPDIR}/opencc"
  info "Installing opencc to ${BIN_TARGET}"
  $SUDO cp "${TMPDIR}/opencc" "$BIN_TARGET"
  $SUDO chmod +x "$BIN_TARGET"
  ok "Installed ${BIN_TARGET}"

  # Create envs dir
  if [ ! -d "$CC_ENVS_DIR" ]; then
    info "Creating ${CC_ENVS_DIR}"
    mkdir -p "$CC_ENVS_DIR"
  fi

  # Create fallback.conf if it doesn't exist
  if [ ! -f "$CC_ENVS_DIR/fallback.conf" ]; then
    touch "$CC_ENVS_DIR/fallback.conf"
  fi

  # Install completions
  install_completions

  printf "\n"
  ok "opencc ${VERSION} installed!"

  # Guide first provider setup if no .env files exist
  _env_count="$(find "$CC_ENVS_DIR" -maxdepth 1 -name '*.env' 2>/dev/null | wc -l | tr -d ' ')"
  if [ "$_env_count" = "0" ]; then
    setup_first_provider
  else
    printf "  Run 'opencc list' to list configurations\n"
    printf "  Run 'opencc use <name>' to start claude with a configuration\n"
  fi
}

install_completions() {
  SUDO="$(need_sudo)"

  # zsh
  if command -v zsh >/dev/null 2>&1; then
    zdir="$(zsh_comp_dir)"
    if [ -n "$zdir" ]; then
      info "Installing zsh completion to $zdir/_opencc"
      "$BIN_TARGET" completion zsh > /tmp/_opencc_comp
      $SUDO mkdir -p "$zdir"
      $SUDO cp /tmp/_opencc_comp "$zdir/_opencc"
      rm -f /tmp/_opencc_comp
      rm -f "$HOME"/.zcompdump*
      ok "zsh completion installed"
    fi
  fi

  # bash
  if command -v bash >/dev/null 2>&1; then
    bdir="$(bash_comp_dir)"
    if [ -n "$bdir" ]; then
      info "Installing bash completion to $bdir/opencc"
      "$BIN_TARGET" completion bash > /tmp/_opencc_bash_comp
      $SUDO cp /tmp/_opencc_bash_comp "$bdir/opencc"
      rm -f /tmp/_opencc_bash_comp
      ok "bash completion installed"
    fi
  fi

  # fish
  if command -v fish >/dev/null 2>&1; then
    fdir="$(fish_comp_dir)"
    info "Installing fish completion to $fdir/opencc.fish"
    mkdir -p "$fdir"
    "$BIN_TARGET" completion fish > "$fdir/opencc.fish"
    ok "fish completion installed"
  fi
}

# --- Uninstall ---

do_uninstall() {
  SUDO="$(need_sudo)"
  info "Uninstalling opencc..."

  if [ -f "$BIN_TARGET" ]; then
    $SUDO rm -f "$BIN_TARGET"
    ok "Removed $BIN_TARGET"
  fi

  if command -v zsh >/dev/null 2>&1; then
    zdir="$(zsh_comp_dir)"
    if [ -n "$zdir" ] && [ -f "$zdir/_opencc" ]; then
      $SUDO rm -f "$zdir/_opencc"
      ok "Removed $zdir/_opencc"
    fi
  fi

  if command -v bash >/dev/null 2>&1; then
    bdir="$(bash_comp_dir)"
    if [ -n "$bdir" ] && [ -f "$bdir/opencc" ]; then
      $SUDO rm -f "$bdir/opencc"
      ok "Removed $bdir/opencc"
    fi
  fi

  fdir="$(fish_comp_dir)"
  if [ -f "$fdir/opencc.fish" ]; then
    rm -f "$fdir/opencc.fish"
    ok "Removed $fdir/opencc.fish"
  fi

  printf "\n"
  ok "opencc has been uninstalled"
  info "~/.cc_envs/ was preserved (your configurations)"
}

# --- Main ---

case "${1:-}" in
  --uninstall)
    do_uninstall
    ;;
  *)
    do_install
    ;;
esac
