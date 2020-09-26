#!/usr/bin/env bash

NAME="migrago"
VERSION="$1"

if [ -z "${VERSION}" ]; then
   echo "VERSION is not set. Use ./release.sh 0.0.0" >&2
   exit 1
fi

rm -fr release
mkdir release
touch release/checksum.txt

function make_release() {
    local arch="$1"
    local os="$2"
    local release_name="$3"
    if [ -z "${arch}" ] || [ -z "${os}" ] || [ -z "${release_name}" ]; then
       echo "args are not set" >&2
       return 1
    fi
    local ext="$4"

    local dir="release/${release_name}"

    mkdir -p "${dir}"
    env GOARCH="${arch}" GOOS="${os}" go build -ldflags "-s -w -X main.Version=${VERSION}" -o "${dir}/${NAME}${ext}"

    cp LICENSE "${dir}"
    cp README.md "${dir}"
    #cp CHANGELOG.md "${dir}"

    cd release/
    case "${os}" in
        linux | darwin)
            tar -zcvf "${release_name}.tar.gz" "${release_name}"
            md5sum "${release_name}.tar.gz" >> checksum.txt
            ;;
        windows)
            zip -r "${release_name}.zip" "${release_name}"
            md5sum "${release_name}.zip" >> checksum.txt
            ;;
    esac
    rm -r "${release_name}"
    cd ../
}

make_release 386 linux "${NAME}-${VERSION}-i386_linux"
make_release amd64 linux "${NAME}-${VERSION}-amd64_linux"
make_release 386 windows "${NAME}-${VERSION}-win32" .exe
make_release amd64 windows "${NAME}-${VERSION}-win64" .exe
make_release amd64 darwin "${NAME}-${VERSION}-amd64_darwin"
