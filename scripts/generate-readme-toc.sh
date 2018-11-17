#!/bin/bash

set -e

if [ ! -d .git ]; then
    echo "This script must be run at root of repository" 1>&2
    exit 1
fi

if [ ! -f ./node_modules/.bin/markdown-toc ]; then
    npm install markdown-toc
fi

echo "+markdown-toc README.md" 1>&2
./node_modules/.bin/markdown-toc README.md
