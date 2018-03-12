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
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const addr = ":8080"

const AuthServer = "http://localhost:8081"
const BookingsServer = "http://localhost:8082"
const HotelsServer = "http://localhost:8083"

func getJson(r *http.Request) (map[string]interface{}, error) {
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	data := map[string]interface{}{}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, nil
}

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

					resp, err := getJson(req)
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
						req, err := http.NewRequest("GET", BookingsServer+fmt.Sprintf("/booking/%d", id), nil)
						if err != nil {
							return nil, err
						}

						jwt, err := utils.NewJWT(user)
						if err != nil {
							return nil, err
						}
						req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

						resp, err := getJson(req)
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
			Type: hotelType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				req, err := http.NewRequest("GET", HotelsServer+"/hotels", nil)
				if err != nil {
					return nil, err
				}

				resp, err := getJson(req)
				if err != nil {
					return nil, err
				}
				respErr, isOk := resp["err"].(string)
				if isOk {
					if respErr != "" {
						return nil, errors.New(respErr)
					}
				}

				log.Println(resp)
				booking, isOk := resp["hotels"].([]interface{})
				if isOk {
					return booking, nil
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
						resp, err := getJson(req)
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
