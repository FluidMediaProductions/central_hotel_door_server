#!/bin/bash

HASH=`git rev-parse --verify HEAD`

docker push evilben/travelr_auth:$HASH
docker push evilben/travelr_bookings:$HASH
docker push evilben/travelr_hotels:$HASH
docker push evilben/travelr_rooms:$HASH
docker push evilben/travelr_gateway:$HASH
docker push evilben/travelr_hotel_gateway:$HASH