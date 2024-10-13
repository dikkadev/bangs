#!/bin/sh

GO_BIN=/usr/local/go/bin/go
STAT_BIN=~/go/bin/benchstat

##############################
remove_logs() { # $1: file
##############################
    sed -i '0,/^goos/{/^goos/!d}' $1
}

#Clean
printf "Cleaning...\n"
rm -f naive.bench precomp.bench benchstat.txt

#Naive
printf "Naive:\n"
$GO_BIN test -bench Naive -count=8 | tee naive.bench
remove_logs naive.bench
sed -i 's#BenchmarkPrepareInputNaive/#Benchmark#g' naive.bench #Remove the benchmark nama for comparing with benchstat

#PreComp
printf "PreComp:\n"
$GO_BIN test -bench PreComp -count=8 | tee precomp.bench
remove_logs precomp.bench
sed -i 's#BenchmarkPrepareInputPreComp/#Benchmark#g' precomp.bench

printf "\n\nResults:\n"

#stat
$STAT_BIN naive.bench precomp.bench | tee benchstat.txt
