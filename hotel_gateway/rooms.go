package main

import (
	"net/http"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"errors"
)

const RoomsServer = "http://rooms"

func getRoomsByHotel(hotel uint) ([]interface{}, error) {
	req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/by-hotel/%d", hotel), nil)
	if err != nil {
		return nil, err
	}

	resp, err := utils.GetJson(req)
	if err != nil {
		return nil, err
	}
	respErr, isOk := resp["err"].(string)
	if isOk {
		if respErr != "" {
			return nil, errors.New(respErr)
		}
	}

	rooms, isOk := resp["rooms"].([]interface{})
	if isOk {
		return rooms, nil
	}
	return nil, nil
}
