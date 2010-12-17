#!/bin/bash

set -ev

./return

./return | grep 'Hello world'
