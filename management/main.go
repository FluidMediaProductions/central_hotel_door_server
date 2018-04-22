package management

import (
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/graphql-go/graphql"
	"reflect"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
	"net/http"
	"log"
	"encoding/json"
	"bytes"
	"errors"
	"github.com/spf13/viper"
)

const addr = ":80"

var AuthServer = "http://auth"
var jwtSecret []byte

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
	},
})

func paginateSlice(arg interface{}, args map[string]interface{}) []interface{} {
	slice, success := takeSliceArg(arg)
	if !success {
		return nil
	}
	offset, isOk := args["offset"].(int)
	if isOk {
		if offset > len(slice) {
			offset = len(slice)
		}
		slice = slice[offset:]
	}
	first, isOk := args["first"].(int)
	if isOk {
		if first > len(slice) {
			first = len(slice)
		}
		slice = slice[:first]
	}
	return slice
}

func takeSliceArg(arg interface{}) (out []interface{}, ok bool) {
	slice, success := takeArg(arg, reflect.Slice)
	if !success {
		ok = false
		return
	}
	c := slice.Len()
	out = make([]interface{}, c)
	for i := 0; i < c; i++ {
		out[i] = slice.Index(i).Interface()
	}
	return out, true
}

func takeArg(arg interface{}, kind reflect.Kind) (val reflect.Value, ok bool) {
	val = reflect.ValueOf(arg)
	if val.Kind() == kind {
		ok = true
	}
	return
}

func makeAuthWrapper(field *graphql.Object) *graphql.Field {
	return &graphql.Field{
		Type: field,
		Args: graphql.FieldConfigArgument{
			"token": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			tokenString, isOK := params.Args["token"].(string)
			if isOK {
				claims, err := utils.VerifyJWT(tokenString, jwtSecret)
				if err != nil {
					return nil, err
				} else {
					return claims.User, nil
				}
			}
			return nil, nil
		},
	}
}

var authenticatedQueries = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthenticatedQueries",
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
	},
})

var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"auth": makeAuthWrapper(authenticatedQueries),
	},
})

var authenticatedMutations = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthenticatedMutations",
	Fields: graphql.Fields{
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

		"auth": makeAuthWrapper(authenticatedMutations),
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    rootQuery,
	Mutation: rootMutation,
})

func main()  {
	viper.SetEnvPrefix("TRAVELR")
	viper.AutomaticEnv()

	jwtSecret = []byte(viper.GetString("JWT_SECRET"))

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