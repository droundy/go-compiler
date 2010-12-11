#!/bin/bash

set -ev

./function

./function | grep 'Hello world'
