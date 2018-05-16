package main

import (
	"net/http"
	"github.com/fluidmediaproductions/central_hotel_door_server/hotel_comms"
	"github.com/golang/protobuf/proto"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"errors"
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
					id, isOk := room["ID"].(string)
					if isOk {
						actionType := hotel_comms.ActionType_ROOM_UNLOCK
						action := &hotel_comms.Action{
							Type: &actionType,
							Id:   proto.String(id),
						}
						actions = append(actions, action)
					}
				}
			}
		}
	}
	return actions, nil
}

func actionComplete(hotel *HotelServer, msg []byte, sig []byte, w http.ResponseWriter) error {
	newMsg := &hotel_comms.ActionComplete{}
	err := proto.Unmarshal(msg, newMsg)
	if err != nil {
		return err
	}

	if newMsg.GetSuccess() {
		if newMsg.GetActionType() == hotel_comms.ActionType_ROOM_UNLOCK {
			err := completeRoomUnlock(newMsg.GetActionId(), hotel.ID)
			if err != nil {
				return err
			}
		}
	}

	resp := &hotel_comms.ActionCompleteResp{}
	return sendMsg(resp, hotel_comms.MsgType_ACTION_COMPLETE_RESP, w)
}

func completeRoomUnlock(roomId string, hotelId string) error {
	req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%s", roomId), nil)

	resp, err := utils.GetJson(req)
	if err != nil {
		return err
	}
	respErr, isOk := resp["err"].(string)
	if isOk {
		if respErr != "" {
			return errors.New(respErr)
		}
	}

	room, isOk := resp["room"].(map[string]interface{})
	if !isOk {
		return errors.New("unable to get room")
	}

	roomHotelId, isOk := room["hotelId"].(string)
	if !isOk {
		return errors.New("unable to get room")
	}

	if hotelId != roomHotelId {
		return errors.New("room not in hotel")
	}

	req, err = http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%s/open-success", roomId), nil)
	if err != nil {
		return  err
	}

	resp, err = utils.GetJson(req)
	if err != nil {
		return err
	}
	respErr, isOk = resp["err"].(string)
	if isOk {
		if respErr != "" {
			return errors.New(respErr)
		}
	}

	return nil
}
