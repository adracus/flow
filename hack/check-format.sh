#!/bin/bash

unformatted="$(gofumpt -l $@)"
if [[ "$unformatted" != "" ]]; then
    echo "Unformatted files detected: $unformatted"
    exit 1
fi

