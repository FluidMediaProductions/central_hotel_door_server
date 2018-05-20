#!/bin/bash

HASH=`git rev-parse --verify HEAD`

docker build -t evilben/travelr_auth:$HASH -f Dockerfile.auth ./
docker build -t evilben/travelr_bookings:$HASH -f Dockerfile.bookings ./
docker build -t evilben/travelr_hotels:$HASH -f Dockerfile.hotels ./
docker build -t evilben/travelr_rooms:$HASH -f Dockerfile.rooms ./
docker build -t evilben/travelr_gateway:$HASH -f Dockerfile.gateway ./
docker build -t evilben/travelr_hotel_gateway:$HASH -f Dockerfile.hotel_gateway ./
docker build -t evilben/travelr_hotel_mqtt_auth:$HASH -f Dockerfile.hotel_mqtt_auth ./
docker build -t evilben/travelr_mosquitto:$HASH -f Dockerfile.mosquitto ./