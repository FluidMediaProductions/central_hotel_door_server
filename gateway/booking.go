package main

import (
	"github.com/graphql-go/graphql"
	"time"
	"net/http"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"errors"
)
var bookingType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Booking",
	Fields: graphql.Fields{
		"ID": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				booking, isOk := params.Source.(map[string]interface{})
				if isOk {
					id, isOk := booking["uid"].(string)
					if isOk {
						return id, nil
					}
				}
				return nil, nil
			},
		},
		"start": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				booking, isOk := params.Source.(map[string]interface{})
				if isOk {
					start, isOk := booking["start"].(string)
					if isOk {
						time, err := time.Parse(time.RFC3339, start)
						if err != nil {
							return nil, err
						}
						return time, nil
					}
				}
				return nil, nil
			},
		},
		"end": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				booking, isOk := params.Source.(map[string]interface{})
				if isOk {
					end, isOk := booking["end"].(string)
					if isOk {
						time, err := time.Parse(time.RFC3339, end)
						if err != nil {
							return nil, err
						}
						return time, nil
					}
				}
				return nil, nil
			},
		},
		"hotel": &graphql.Field{
			Type: hotelType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				booking, isOk := params.Source.(map[string]interface{})
				if isOk {
					hotelId, isOk := booking["hotelId"].(string)
					if isOk {
						req, err := http.NewRequest("GET", HotelsServer+fmt.Sprintf("/hotels/%s", hotelId), nil)
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

						hotel, isOk := resp["hotel"].(map[string]interface{})
						if isOk {
							return hotel, nil
						}
					}
				}
				return nil, nil
			},
		},
		"room": &graphql.Field{
			Type: roomType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				booking, isOk := params.Source.(map[string]interface{})
				if isOk {
					roomId, isOk := booking["roomId"].(string)
					if isOk {
						req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%s", roomId), nil)
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

						room, isOk := resp["room"].(map[string]interface{})
						if isOk {
							return room, nil
						}
					}
				}
				return nil, nil
			},
		},
	},
})
