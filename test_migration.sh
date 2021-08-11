#!/bin/bash

function generate() {
  delete

  COUNTMIN="${1:-1}"
  COUNTSEC="${2:-1}"

  for MIN in $(seq -f "%02g" 1 $COUNTMIN)
  do
    for SEC in $(seq -f "%02g" 1 $COUNTSEC)
    do
      echo "SELECT 1;" > example/migrations/postgres/20210101${MIN}${SEC}_example_up.sql
    done
  done
}

function delete() {
  rm -f example/migrations/postgres/20210101*
}

# run function from argument
$1 $2 $3