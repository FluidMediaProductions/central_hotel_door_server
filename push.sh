#!/bin/bash

HASH=`git rev-parse --verify HEAD`

sudo docker build -t evilben/travel_auth:$HASH auth/
sudo docker push evilben/travel_auth:$HASH

sudo docker build -t evilben/travel_bookings:$HASH bookings/
sudo docker push evilben/travel_bookings:$HASH