#!/bin/bash

HASH=`git rev-parse --verify HEAD`

sudo docker build -t evilben/travelr_auth:$HASH auth/
sudo docker build -t evilben/travelr_bookings:$HASH bookings/
sudo docker build -t evilben/travelr_hotels:$HASH hotels/
sudo docker build -t evilben/travelr_rooms:$HASH rooms/
sudo docker build -t evilben/travelr_gateway:$HASH gateway/
sudo docker build -t evilben/travelr_hotel_gateway:$HASH hotel_gateway/