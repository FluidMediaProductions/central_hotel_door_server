package main

import (
	"github.com/graphql-go/graphql"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"net/http"
	"fmt"
	"errors"
	"github.com/mitchellh/mapstructure"
)

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
			Type: graphql.NewNonNull(graphql.String),
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

func getUser(token string) (*utils.User, error) {
	user, err := getUserFromAuthServer(token)
	if err == nil {
		return user, nil
	}

	claims, err := utils.VerifyJWT(token, jwtSecret)
	if err != nil {
		return nil, err
	}

	return claims.User, nil
}

func getUserFromAuthServer(token string) (*utils.User, error) {
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
	return nil, errors.New("invalid data from auth server")
}
