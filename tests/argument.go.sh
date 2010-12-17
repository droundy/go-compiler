#!/bin/bash

set -ev

./argument

./argument 2> err
diff -u err - <<EOF
Hello world!
EOF
