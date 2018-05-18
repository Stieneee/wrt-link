#!/bin/bash
COUNTER=0
while [  $COUNTER -lt 1000 ]; do
  ./awktest.sh $1 $2
  let COUNTER=COUNTER+1
done
