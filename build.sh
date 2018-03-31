#!/bin/bash

HASH=`git rev-parse --verify HEAD`

sudo docker build -t evilben/travelr_auth:$HASH auth/
sudo docker build -t evilben/travelr_bookings:$HASH bookings/