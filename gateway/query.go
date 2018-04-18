package main

import (
	"github.com/graphql-go/graphql"
	"net/http"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/mitchellh/mapstructure"
	"errors"
)

var authedQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthedQuery",
	Fields: graphql.Fields{
		"self": &graphql.Field{
			Type: userType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				user, isOk := params.Source.(*utils.User)
				if isOk {
					return user, nil
				}
				return nil, nil
			},
		},
		"booking": &graphql.Field{
			Type: bookingType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, isOK := params.Args["id"].(int)
				if isOK {
					user, isOk := params.Source.(*utils.User)
					if isOk {
						req, err := http.NewRequest("GET", BookingsServer+fmt.Sprintf("/bookings/%d", id), nil)
						if err != nil {
							return nil, err
						}

						jwt, err := utils.NewJWT(user, jwtSecret)
						if err != nil {
							return nil, err
						}
						req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

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

						booking, isOk := resp["booking"].(map[string]interface{})
						if isOk {
							return booking, nil
						}
					}
				}
				return nil, nil
			},
		},
	},
})

var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"auth": &graphql.Field{
			Type: authedQuery,
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				token, isOK := params.Args["token"].(string)
				if isOK {
					req, err := http.NewRequest("GET", AuthServer+"/userInfo", nil)
					if err != nil {
						return nil, err
					}

					req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

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

					jsonUser, isOk := resp["user"].(map[string]interface{})
					if isOk {
						user := &utils.User{}
						mapstructure.Decode(jsonUser, user)
						return user, nil
					}
				}
				return nil, nil
			},
		},
		"hotels": &graphql.Field{
			Type: graphql.NewList(hotelType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				req, err := http.NewRequest("GET", HotelsServer+"/hotels", nil)
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

				hotels, isOk := resp["hotels"].([]interface{})
				if isOk {
					return hotels, nil
				}
				return nil, nil
			},
		},
		"hotel": &graphql.Field{
			Type: hotelType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, isOK := params.Args["id"].(int)
				if isOK {
					req, err := http.NewRequest("GET", HotelsServer+fmt.Sprintf("/hotels/%d", id), nil)
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
				return nil, nil
			},
		},
		"rooms": &graphql.Field{
			Type: graphql.NewList(roomType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				req, err := http.NewRequest("GET", RoomsServer+"/rooms", nil)
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
			},
		},
		"room": &graphql.Field{
			Type: roomType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, isOK := params.Args["id"].(int)
				if isOK {
					req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%d", id), nil)
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
				return nil, nil
			},
		},
	},
})
