package main

import (
	"github.com/graphql-go/graphql"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"net/http"
	"bytes"
	"fmt"
	"encoding/json"
	"errors"
)

var authedMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthedMutation",
	Fields: graphql.Fields{
		"changePassword": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"pass": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				pass, isOK := params.Args["pass"].(string)
				if isOK {
					user, isOk := params.Source.(*utils.User)
					if isOk {
						data := map[string]interface{}{
							"pass": pass,
						}
						dataBytes, err := json.Marshal(data)
						if err != nil {
							return nil, err
						}

						req, err := http.NewRequest("POST", AuthServer+"/changePassword", bytes.NewBuffer(dataBytes))
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

						success, isOk := resp["success"].(bool)
						if isOk {
							return success, nil
						}
					}
				}
				return nil, nil
			},
		},
		"updateUser": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"email": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				user, isOk := params.Source.(*utils.User)
				if isOk {
					data := map[string]interface{}{}

					email, isOK := params.Args["email"].(string)
					if isOK {
						data["email"] = email
					}
					name, isOK := params.Args["name"].(string)
					if isOK {
						data["name"] = name
					}

					dataBytes, err := json.Marshal(data)
					if err != nil {
						return nil, err
					}

					req, err := http.NewRequest("POST", AuthServer+"/updateUser", bytes.NewBuffer(dataBytes))
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

					success, isOk := resp["success"].(bool)
					if isOk {
						return success, nil
					}
				}
				return nil, nil
			},
		},

		"openRoom": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, isOK := params.Args["id"].(string)
				if isOK {
					user, isOk := params.Source.(*utils.User)
					if isOk {
						req, err := http.NewRequest("GET", BookingsServer+fmt.Sprintf("/bookings/by-room/%s", id), nil)
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

						_, isOk = resp["booking"].(map[string]interface{})
						if isOk {
							req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%s/open", id), nil)
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

							success, isOk := resp["success"].(bool)
							if isOk {
								return success, nil
							}
						}
					}
				}
				return nil, nil
			},
		},
		"openHotelDoor": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, isOK := params.Args["id"].(string)
				if isOK {
					user, isOk := params.Source.(*utils.User)
					if isOk {
						req, err := http.NewRequest("GET", HotelsServer+fmt.Sprintf("/hotels/%s/open", id), nil)
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

						success, isOk := resp["success"].(bool)
						if isOk {
							return success, nil
						}
					}
				}
				return nil, nil
			},
		},
	},
})

var rootMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootMutation",
	Fields: graphql.Fields{
		"loginUser": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"pass": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				email, isOK := params.Args["email"].(string)
				if isOK {
					pass, isOK := params.Args["pass"].(string)
					if isOK {
						data := map[string]interface{}{
							"email": email,
							"pass":  pass,
						}
						dataBytes, err := json.Marshal(data)
						if err != nil {
							return nil, err
						}
						req, err := http.NewRequest("POST", AuthServer+"/login", bytes.NewBuffer(dataBytes))
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
						jwt, isOk := resp["jwt"].(string)
						if isOk {
							return jwt, nil
						}
					}
				}
				return nil, nil
			},
		},
		"auth": &graphql.Field{
			Type: authedMutation,
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				token, isOK := params.Args["token"].(string)
				if isOK {
					claims, err := utils.VerifyJWT(token, jwtSecret)
					if err != nil {
						return nil, err
					}
					return claims.User, nil
				}
				return nil, nil
			},
		},
	},
})
