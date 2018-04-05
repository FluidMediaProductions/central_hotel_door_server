#!/bin/bash

HASH=`git rev-parse --verify HEAD`

sudo docker push evilben/travelr_auth:$HASH
sudo docker push evilben/travelr_bookings:$HASH
sudo docker push evilben/travelr_hotels:$HASH
sudo docker push evilben/travelr_rooms:$HASH
sudo docker push evilben/travelr_gateway:$HASH