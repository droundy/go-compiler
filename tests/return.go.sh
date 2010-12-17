#!/bin/bash

set -ev

./return

./return 2> err
diff -u err - <<EOF
Hello world!
EOF
