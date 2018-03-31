#!/bin/bash

HASH=`git rev-parse --verify HEAD`

sudo docker push evilben/travel_auth:$HASH
sudo docker push evilben/travel_bookings:$HASH