#! /bin/bash

set -e

go test -v
golint

rm -rf release
gox -arch 'amd64' -os 'linux darwin windows freebsd openbsd netbsd' ./cmd/notes
mkdir -p release
mv notes_* release/
cd release
for bin in *; do
    if [[ "$bin" == *windows* ]]; then
        command="notes.exe"
    else
        command="notes"
    fi
    mv "$bin" "$command"
    zip "${bin}.zip" "$command"
    rm "$command"
done
