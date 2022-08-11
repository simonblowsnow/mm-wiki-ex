#!/bin/bash
export PYTHONPATH=`pwd`
export PYTHONPATH=$PYTHONPATH:`pwd`/packages
echo $PYTHONPATH
python src/main.py
