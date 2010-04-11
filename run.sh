#!/bin/bash
PWD=`pwd`
BASE=`basename $PWD`
if [[ ! -e $BASE.machine ]]; then
    echo "$BASE.machine does not exist"
    exit 1
fi
if [[ ! -e $BASE.inputs ]]; then
    echo "$BASE.inputs does not exist"
    exit 1
fi
cat $BASE.machine | ../machine2dot.pl --html | dot -Tpng -o $BASE.png
cat $BASE.inputs | xargs ../assemble.pl $BASE.machine $@
