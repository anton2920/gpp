#!/bin/sh

./make.sh $@ && ./make.sh vet && ./gpp -l -r testdata/
