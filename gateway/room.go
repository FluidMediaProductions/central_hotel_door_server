package main

import (
	"github.com/graphql-go/graphql"
	"net/http"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"errors"
)

var roomType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Room",
	Fields: graphql.Fields{
		"ID": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				room, isOk := params.Source.(map[string]interface{})
				if isOk {
					id, isOk := room["uid"].(string)
					if isOk {
						return id, nil
					}
				}
				return nil, nil
			},
		},
		"name": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				room, isOk := params.Source.(map[string]interface{})
				if isOk {
					name, isOk := room["name"].(string)
					if isOk {
						return name, nil
					}
				}
				return nil, nil
			},
		},
		"floor": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				room, isOk := params.Source.(map[string]interface{})
				if isOk {
					floor, isOk := room["floor"].(string)
					if isOk {
						return floor, nil
					}
				}
				return nil, nil
			},
		},
		"hotel": &graphql.Field{
			Type: hotelType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				room, isOk := params.Source.(map[string]interface{})
				if isOk {
					hotelId, isOk := room["hotelId"].(string)
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
	},
})
