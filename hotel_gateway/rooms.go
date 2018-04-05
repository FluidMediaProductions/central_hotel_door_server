package main

import (
	"net/http"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"errors"
	"github.com/fluidmediaproductions/central_hotel_door_server/hotel_comms"
	"github.com/golang/protobuf/proto"
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

func getDoors(hotel *HotelServer, msg []byte, sig []byte, w http.ResponseWriter) error {
	newMsg := &hotel_comms.GetDoors{}
	err := proto.Unmarshal(msg, newMsg)
	if err != nil {
		return err
	}

	rooms, err := getRoomsByHotel(hotel.HotelId)
	if err != nil {
		return err
	}

	doors := make([]*hotel_comms.Door, 0)
	for _, room := range rooms {
		room, isOk := room.(map[string]interface{})
		if isOk {
			name, isOk := room["name"].(string)
			if isOk {
				id, isOk := room["ID"].(float64)
				if isOk {
					door := &hotel_comms.Door{
						Id:   proto.Int64(int64(id)),
						Name: proto.String(name),
					}
					doors = append(doors, door)
				}
			}
		}
	}

	resp := &hotel_comms.GetDoorsResp{
		Doors: doors,
	}

	w.WriteHeader(http.StatusOK)
	return sendMsg(resp, hotel_comms.MsgType_HOTEL_PING_RESP, w)
}
