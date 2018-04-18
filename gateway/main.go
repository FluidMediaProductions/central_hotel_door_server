package main

import (
	"log"
	"net/http"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

const addr = ":80"

var AuthServer = "http://auth"
var BookingsServer = "http://bookings"
var HotelsServer = "http://hotels"
var RoomsServer = "http://rooms"
var jwtSecret []byte

func initSchema() (graphql.Schema, error) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})

	return schema, err
}

func main() {
	viper.SetEnvPrefix("TRAVELR")
	viper.AutomaticEnv()

	jwtSecret = []byte(viper.GetString("JWT_SECRET"))

	schema, err := initSchema()
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
