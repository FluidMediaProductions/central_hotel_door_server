package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"log"
	"net/http"
	"time"
)

const addr = ":80"

const AuthServer = "http://auth"
const BookingsServer = "http://bookings"
const HotelsServer = "http://hotels"
const RoomsServer = "http://rooms"

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"email": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"name": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"ID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				user, isOk := params.Source.(*utils.User)
				if isOk {
					return user.Model.ID, nil
				}
				return nil, nil
			},
		},
		"bookings": &graphql.Field{
			Type: graphql.NewList(bookingType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				user, isOk := params.Source.(*utils.User)
				if isOk {
					req, err := http.NewRequest("GET", BookingsServer+"/bookings", nil)
					if err != nil {
						return nil, err
					}

					jwt, err := utils.NewJWT(user)
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

					bookings, isOk := resp["bookings"].([]interface{})
					if isOk {
						return bookings, nil
					}
				}
				return nil, nil
			},
		},
	},
})

var bookingType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Booking",
	Fields: graphql.Fields{
		"ID": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				booking, isOk := params.Source.(map[string]interface{})
				if isOk {
					id, isOk := booking["ID"].(float64)
					if isOk {
						return int(id), nil
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
					hotelId, isOk := booking["hotelId"].(float64)
					if isOk {
						req, err := http.NewRequest("GET", HotelsServer+fmt.Sprintf("/hotels/%d", int(hotelId)), nil)
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
					roomId, isOk := booking["roomId"].(float64)
					if isOk {
						req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%d", int(roomId)), nil)
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

var hotelType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Hotel",
	Fields: graphql.Fields{
		"ID": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				hotel, isOk := params.Source.(map[string]interface{})
				if isOk {
					id, isOk := hotel["ID"].(float64)
					if isOk {
						return int(id), nil
					}
				}
				return nil, nil
			},
		},
		"checkIn": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				hotel, isOk := params.Source.(map[string]interface{})
				if isOk {
					checkIn, isOk := hotel["checkIn"].(string)
					if isOk {
						time, err := time.Parse(time.RFC3339, checkIn)
						if err != nil {
							return nil, err
						}
						return time, nil
					}
				}
				return nil, nil
			},
		},
		"name": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				hotel, isOk := params.Source.(map[string]interface{})
				if isOk {
					name, isOk := hotel["name"].(string)
					if isOk {
						return name, nil
					}
				}
				return nil, nil
			},
		},
		"address": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				hotel, isOk := params.Source.(map[string]interface{})
				if isOk {
					address, isOk := hotel["address"].(string)
					if isOk {
						return address, nil
					}
				}
				return nil, nil
			},
		},
		"hasCarPark": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				hotel, isOk := params.Source.(map[string]interface{})
				if isOk {
					hasCarPark, isOk := hotel["hasCarPark"].(bool)
					if isOk {
						return hasCarPark, nil
					}
				}
				return nil, nil
			},
		},
	},
})

var roomType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Room",
	Fields: graphql.Fields{
		"ID": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				room, isOk := params.Source.(map[string]interface{})
				if isOk {
					id, isOk := room["ID"].(float64)
					if isOk {
						return int(id), nil
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
					hotelId, isOk := room["hotelId"].(float64)
					if isOk {
						req, err := http.NewRequest("GET", HotelsServer+fmt.Sprintf("/hotels/%d", int(hotelId)), nil)
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

						jwt, err := utils.NewJWT(user)
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
					claims, err := utils.VerifyJWT(token)
					if err != nil {
						return nil, err
					}
					return claims.User, nil
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

var authedMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthedMutation",
	Fields: graphql.Fields{
		"openRoom": &graphql.Field{
			Type: graphql.Boolean,
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
						req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/%d/open", id), nil)
						if err != nil {
							return nil, err
						}

						jwt, err := utils.NewJWT(user)
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
		"openHotelDoor": &graphql.Field{
			Type: graphql.Boolean,
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
						req, err := http.NewRequest("GET", HotelsServer+fmt.Sprintf("/hotels/%d/open", id), nil)
						if err != nil {
							return nil, err
						}

						jwt, err := utils.NewJWT(user)
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
					claims, err := utils.VerifyJWT(token)
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

func main() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	corsH := cors.Default().Handler(h)

	http.Handle("/graphql", corsH)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
