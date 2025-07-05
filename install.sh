#!/usr/bin/env bash
set -eo pipefail

if [[ ${OS:-} = Windows_NT ]]; then
    echo 'error: please install kool using Windows Subsystem for Linux'
    exit 1
fi

error() {
    echo "$@" >&2
    exit 1
}

# command bypass builtin alias, looking for real command
if [[ $PREFER_GO = true ]];then
    if command -v go >/dev/null ;then
        echo "go install github.com/xhd2015/kool@latest"
        go install github.com/xhd2015/kool@latest
        exit
    fi
fi

command -v tar >/dev/null || error 'tar is required to install kool'

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

#for pre-release like:
#   https://github.com/xhd2015/kool/releases/download/v1.1.0-alpha/kool-v1.1.0-linux-amd64

if [[ "$INSTALL_TAG" != "" ]];then
    install_version=$INSTALL_VERSION
    if [[ -z "$install_version" ]];then
        install_version=$INSTALL_TAG
    fi
    # trim v prefix
    install_version=${install_version/#"v"}
    file="kool-v${install_version}-${target}"
    uri="https://github.com/xhd2015/kool/releases/download/${INSTALL_TAG}/${file}"
else
    latestURL="https://github.com/xhd2015/kool/releases/latest"
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
    if [[ "$locationURL" = *'/kool-v'* ]];then
        versionName=${locationURL/#*'/kool-v'}
        elif [[ "$locationURL" = *'/tag/v'* ]];then
        versionName=${locationURL/#*'/tag/v'}
    fi
    
    if [[ -z $versionName ]];then
        error "expect tag format: kool-v1.x.x, actual: $locationURL"
    fi
    
    file="kool-v${versionName}-${target}"
    uri="$latestURL/download/$file"
fi

install_dir=$HOME/.xgo
bin_dir=$install_dir/bin

if [[ ! -d $bin_dir ]]; then
    mkdir -p "$bin_dir" || error "failed to create install directory \"$bin_dir\""
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

curl --fail --location --progress-bar --output "${tmp_dir}/${file}" "$uri" || error "failed to download kool from \"$uri\""

chmod +x "${tmp_dir}/${file}"

mv "${tmp_dir}/${file}" "${tmp_dir}/kool"

# install fails if target already exists
echo "installing kool to /usr/local/bin, which may require sudo"
if [[ -f /usr/local/bin/kool ]];then
    sudo mv /usr/local/bin/{kool,kool_backup}
fi
sudo install "${tmp_dir}/kool" /usr/local/bin
sudo rm -f /usr/local/bin/kool_backup || true

echo "Successfully installed, to get started, run:"
echo "  kool help"
echo ""