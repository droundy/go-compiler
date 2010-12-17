#!/bin/bash

set -ev

./voidreturn

./voidreturn 2> err
diff -u err - <<EOF
Hello world!
EOF
