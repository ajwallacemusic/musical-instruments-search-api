#!/bin/sh
 
set -eu
 
echo "Checking ES connection ..."
 
i=0
until [ $i -ge 10 ]
do
  nc -zv es01 9200 && break
 
  i=$(( i + 1 ))
 
  echo "$i: Waiting for ES 1 second ..."
  sleep 1
done
 
if [ $i -eq 10 ]
then
  echo "ES connection refused, terminating ..."
  exit 1
fi
 
echo "ES is up ..."
 
/app