#!/bin/bash
set -e

ulimit -n 1000000
/go/bin/sensephreak $@
