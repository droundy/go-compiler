#!/bin/bash

set -ev

# FIXME: println segfaults!  :(
./function

./function | grep 'Hello world'
