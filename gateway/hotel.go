package main

import (
	"github.com/graphql-go/graphql"
	"time"
)

var hotelType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Hotel",
	Fields: graphql.Fields{
		"ID": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				hotel, isOk := params.Source.(map[string]interface{})
				if isOk {
					id, isOk := hotel["uid"].(string)
					if isOk {
						return id, nil
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
