#!/bin/bash

set -ev

# FIXME: println segfaults!  :(
./voidreturn

# The return in this function should keep the unfriendliness from
# showing..
./voidreturn | grep 'Die evil' && exit 1

./voidreturn | grep 'Hello world'
