#! /bin/bash
echo "source file lines:"
find src -name "*.go" | grep -v "_test.go" | xargs cat | grep -c -v ^$
echo "test file lines:"
find src -name "*.go" | grep "_test.go" | xargs cat | grep -c -v ^$