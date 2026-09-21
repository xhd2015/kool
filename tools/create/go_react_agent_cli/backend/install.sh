#!/usr/bin/env bash
set -eo pipefail

if [[ ${OS:-} = Windows_NT ]]; then
    echo 'error: please install __PROJECT_NAME__ using Windows Subsystem for Linux'
    exit 1
fi

error() {
    echo "$@" >&2
    exit 1
}

# --- ~/.local/bin PATH marker helpers (ported from spl install.sh) ---
LOCAL_BIN_CHECKER_BEGIN='# ----- BEGIN ~/.local/bin checker -----'
LOCAL_BIN_CHECKER_END='# ----- END ~/.local/bin checker -----'

default_local_bin_dir() {
    if [[ -z "${HOME:-}" ]]; then
        echo "error: HOME is unset; cannot install to ~/.local/bin" >&2
        return 1
    fi
    printf '%s\n' "${HOME}/.local/bin"
}

dest_is_default_local_bin() {
    local dest="${1:-}"
    local want
    dest="${dest%/}"
    want=$(default_local_bin_dir) || return 1
    want="${want%/}"
    [[ "$dest" == "$want" ]]
}

canonical_local_bin_checker_block() {
    cat <<'EOF'
# ----- BEGIN ~/.local/bin checker -----
case ":$PATH:" in
  *":$HOME/.local/bin:"*) ;;
  *) export PATH="$HOME/.local/bin:$PATH" ;;
esac
# ----- END ~/.local/bin checker -----
EOF
}

# Prints: created|appended|replaced|unchanged|skipped_missing
# $2 = 1 create if missing, 0 skip if missing.
ensure_local_bin_checker_in_file() {
    local file="$1"
    local create="${2:-1}"
    local canonical
    canonical=$(canonical_local_bin_checker_block)

    if [[ ! -e "$file" ]]; then
        if [[ "$create" != "1" ]]; then
            printf '%s\n' skipped_missing
            return 0
        fi
        if ! printf '%s\n' "$canonical" >"$file"; then
            echo "warning: could not update ${file} (permission denied); add this to your shell rc:" >&2
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" >&2
            return 1
        fi
        printf '%s\n' created
        return 0
    fi

    if [[ ! -f "$file" ]]; then
        echo "warning: ${file} is not a regular file; skipping" >&2
        printf '%s\n' skipped_missing
        return 0
    fi

    if [[ ! -w "$file" ]]; then
        echo "warning: could not update ${file} (permission denied); add this to your shell rc:" >&2
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" >&2
        return 1
    fi

    local -a lines=()
    local line
    while IFS= read -r line || [[ -n "$line" ]]; do
        line="${line%$'\r'}"
        lines+=("$line")
    done < "$file"

    local n=${#lines[@]}
    local -a start_idxs=()
    local -a end_idxs=()
    local -a orphan_idxs=()
    local i
    local open=-1
    for ((i = 0; i < n; i++)); do
        if [[ "${lines[i]}" == "$LOCAL_BIN_CHECKER_BEGIN" ]]; then
            if [[ $open -ge 0 ]]; then
                orphan_idxs+=("$open")
            fi
            open=$i
        elif [[ "${lines[i]}" == "$LOCAL_BIN_CHECKER_END" && $open -ge 0 ]]; then
            start_idxs+=("$open")
            end_idxs+=("$i")
            open=-1
        fi
    done
    if [[ $open -ge 0 ]]; then
        orphan_idxs+=("$open")
    fi

    local -a canon_lines=()
    while IFS= read -r line || [[ -n "$line" ]]; do
        canon_lines+=("$line")
    done <<EOF
$canonical
EOF

    local nblocks=${#start_idxs[@]}
    local norphans=${#orphan_idxs[@]}
    local identical=0
    if [[ $nblocks -eq 1 && $norphans -eq 0 ]]; then
        local s=${start_idxs[0]}
        local e=${end_idxs[0]}
        local blen=$((e - s + 1))
        if [[ $blen -eq ${#canon_lines[@]} ]]; then
            identical=1
            local j
            for ((j = 0; j < blen; j++)); do
                if [[ "${lines[s + j]}" != "${canon_lines[j]}" ]]; then
                    identical=0
                    break
                fi
            done
        fi
    fi

    if [[ $identical -eq 1 ]]; then
        printf '%s\n' unchanged
        return 0
    fi

    local -a out=()
    local emitted=0
    local bi
    i=0
    while [[ $i -lt $n ]]; do
        local is_block=0
        for ((bi = 0; bi < nblocks; bi++)); do
            if [[ $i -eq ${start_idxs[bi]} ]]; then
                is_block=1
                if [[ $emitted -eq 0 ]]; then
                    out+=("${canon_lines[@]}")
                    emitted=1
                fi
                i=$((${end_idxs[bi]} + 1))
                break
            fi
        done
        if [[ $is_block -eq 1 ]]; then
            continue
        fi
        out+=("${lines[i]}")
        i=$((i + 1))
    done

    local action
    if [[ $emitted -eq 0 ]]; then
        if [[ ${#out[@]} -gt 0 ]]; then
            local last=$((${#out[@]} - 1))
            if [[ -n "${out[last]}" ]]; then
                out+=("")
            fi
        fi
        out+=("${canon_lines[@]}")
        action=appended
        if [[ $norphans -gt 0 ]]; then
            echo "warning: ${file} has a BEGIN ~/.local/bin checker without END; appended a new block" >&2
        fi
    else
        action=replaced
        if [[ $norphans -gt 0 ]]; then
            echo "warning: ${file} has a BEGIN ~/.local/bin checker without END; left it in place" >&2
        fi
    fi

    local tmp
    tmp=$(mktemp "${file}.XXXXXX") || return 1
    local ol
    for ol in "${out[@]}"; do
        printf '%s\n' "$ol"
    done >"$tmp"
    local mode
    mode=$(stat -f '%Lp' "$file" 2>/dev/null || stat -c '%a' "$file" 2>/dev/null || true)
    if [[ -n "$mode" ]]; then
        chmod "$mode" "$tmp" 2>/dev/null || true
    fi
    if ! mv "$tmp" "$file"; then
        rm -f "$tmp"
        echo "warning: could not update ${file} (permission denied); add this to your shell rc:" >&2
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" >&2
        return 1
    fi
    printf '%s\n' "$action"
}

export_local_bin_on_path() {
    export PATH="${HOME}/.local/bin:${PATH:-}"
}

pretty_rc_name() {
    printf '~/%s\n' "$(basename "$1")"
}

ensure_local_bin_on_path() {
    local dest_dir="${1:-}"
    dest_is_default_local_bin "$dest_dir" || return 0

    export_local_bin_on_path

    local -a updated=()
    local status f
    for f in "${HOME}/.bash_profile" "${HOME}/.bashrc" "${HOME}/.zshrc"; do
        status=$(ensure_local_bin_checker_in_file "$f" 1) || continue
        case "$status" in
            created|appended|replaced) updated+=("$f") ;;
        esac
    done
    for f in "${HOME}/.zprofile" "${HOME}/.profile"; do
        [[ -e "$f" ]] || continue
        status=$(ensure_local_bin_checker_in_file "$f" 0) || continue
        case "$status" in
            created|appended|replaced) updated+=("$f") ;;
        esac
    done

    if [[ ${#updated[@]} -gt 0 ]]; then
        local names=""
        local u
        for u in "${updated[@]}"; do
            if [[ -n "$names" ]]; then
                names+=", "
            fi
            names+=$(pretty_rc_name "$u")
        done
        echo "Added ~/.local/bin to PATH in ${names}"
        echo "Open a new terminal, or run: export PATH=\"\$HOME/.local/bin:\$PATH\""
    else
        echo "PATH already includes ~/.local/bin"
    fi
}

# Sourced by tests: define helpers only.
if [[ "${KOOL_INSTALL_TEST_LIB:-}" == "1" ]]; then
    return 0 2>/dev/null || exit 0
fi
# --- end ~/.local/bin PATH marker helpers ---

command -v tar >/dev/null || error 'tar is required to install __PROJECT_NAME__'

case $(uname -ms) in
    'Darwin x86_64')
        target=darwin-amd64
    ;;
    'Darwin arm64')
        target=darwin-arm64
    ;;
    'Linux aarch64' | 'Linux arm64')
        target=linux-arm64
    ;;
    'Linux x86_64' | *)
        target=linux-amd64
    ;;
esac

if [[ "$INSTALL_TAG" != "" ]];then
    install_version=$INSTALL_VERSION
    if [[ -z "$install_version" ]];then
        install_version=$INSTALL_TAG
    fi
    install_version=${install_version/#"v"}
    file="__PROJECT_NAME__-v${install_version}-${target}"
    uri="https://github.com/__OWNER__/__REPO__/releases/download/${INSTALL_TAG}/${file}"
else
    latestURL="https://github.com/__OWNER__/__REPO__/releases/latest"
    headers=$(curl "$latestURL" -so /dev/null -D -)
    if [[ "$headers" != *302* ]];then
        error "expect 302 from $latestURL"
    fi

    location=$(echo "$headers"|grep "location: ")
    if [[ -z $location ]];then
        error "expect 302 location from $latestURL"
    fi
    locationURL=${location/#"location: "}
    locationURL=${locationURL/%$'\n'}
    locationURL=${locationURL/%$'\r'}

    versionName=""
    if [[ "$locationURL" = *'/__PROJECT_NAME__-v'* ]];then
        versionName=${locationURL/#*'/__PROJECT_NAME__-v'}
        elif [[ "$locationURL" = *'/tag/v'* ]];then
        versionName=${locationURL/#*'/tag/v'}
    fi

    if [[ -z $versionName ]];then
        error "expect tag format: __PROJECT_NAME__-v1.x.x, actual: $locationURL"
    fi

    file="__PROJECT_NAME__-v${versionName}-${target}"
    uri="$latestURL/download/$file"
fi

dest_dir=$(default_local_bin_dir) || exit 1
mkdir -p "$dest_dir"
export_local_bin_on_path

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

curl --fail --location --progress-bar --output "${tmp_dir}/${file}" "$uri" || error "failed to download __PROJECT_NAME__ from \"$uri\""

chmod +x "${tmp_dir}/${file}"

mv "${tmp_dir}/${file}" "${tmp_dir}/__PROJECT_NAME__"

echo "installing __PROJECT_NAME__ to ${dest_dir}"
install "${tmp_dir}/__PROJECT_NAME__" "${dest_dir}/__PROJECT_NAME__"

ensure_local_bin_on_path "$dest_dir" || true

echo "Successfully installed, to get started, run:"
echo "  __PROJECT_NAME__ --help"
