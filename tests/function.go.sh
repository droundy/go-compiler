#!/bin/bash

set -ev

./function

./function 2> err
diff -u err - <<EOF
Hello world!
EOF
