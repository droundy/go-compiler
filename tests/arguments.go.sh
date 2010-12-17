#!/bin/bash

set -ev

./arguments

./arguments 2> err
diff -u err - <<EOF
Hello world!
EOF
