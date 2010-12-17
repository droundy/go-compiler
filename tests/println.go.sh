#!/bin/bash

set -ev

./println

./println 2> err
diff -u err - <<EOF
Hello world!
EOF
