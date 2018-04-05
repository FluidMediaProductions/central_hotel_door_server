package main

import (
	"net/http"
	"github.com/fluidmediaproductions/central_hotel_door_server/hotel_comms"
	"github.com/golang/protobuf/proto"
	"log"
)

func getAction(hotel *HotelServer, msg []byte, sig []byte, w http.ResponseWriter) error {
	newMsg := &hotel_comms.GetActions{}
	err := proto.Unmarshal(msg, newMsg)
	if err != nil {
		return err
	}

	actions, err := getActions(hotel.HotelId)
	if err != nil {
		return err
	}

	log.Println(actions)

	resp := &hotel_comms.GetActionsResp{
		Actions: actions,
	}

	return sendMsg(resp, hotel_comms.MsgType_GET_ACTIONS_RESP, w)
}

func getActions(hotelId uint) ([]*hotel_comms.Action, error) {
	actions := make([]*hotel_comms.Action, 0)
	rooms, err := getRoomsByHotel(hotelId)
	if err != nil {
		return nil, err
	}
	for _, room := range rooms {
		room, isOk := room.(map[string]interface{})
		if isOk {
			shouldOpen, isOk := room["shouldOpen"].(bool)
			if isOk {
				if shouldOpen {
					id, isOk := room["ID"].(float64)
					if isOk {
						actionType := hotel_comms.ActionType_ROOM_UNLOCK
						action := &hotel_comms.Action{
							Type: &actionType,
							Id:   proto.Int64(int64(id)),
						}
						actions = append(actions, action)
					}
				}
			}
		}
	}
	return actions, nil
}
