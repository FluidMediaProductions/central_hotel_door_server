#!/bin/bash

HASH=`git rev-parse --verify HEAD`

cat auth/deployment.yaml | sed "s/(hash)/$HASH/g" | kubectl apply -f -
cat bookings/deployment.yaml | sed "s/(hash)/$HASH/g" | kubectl apply -f -