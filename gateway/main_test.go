package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/graphql-go/graphql"
	"github.com/jinzhu/gorm"
)

func runQuery(query string, variables map[string]interface{}, t *testing.T) *graphql.Result {
	schema, err := initSchema()
	if err != nil {
		t.Fatalf("Error creating schema: %v", err)
	}
	params := graphql.Params{Schema: schema, RequestString: query, VariableValues: variables}
	r := graphql.Do(params)
	if r.HasErrors() {
		t.Logf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	return r
}

func TestQueryAuth(t *testing.T) {
	user := &utils.User{
		Name:  "Bob",
		Email: "foo@bar.com",
		Model: gorm.Model{
			ID: 1,
		},
	}

	jwt, err := utils.NewJWT(user)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}

	query := `
		query ($token: String!) {
			auth(token: $token) {
				self {
					name
					email
					ID
				}
			}
        }
	`
	variables := map[string]interface{}{
		"token": jwt,
	}
	res := runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}

	auth, isOk := data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}

	self, isOk := auth["self"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][self], expected type map[string]interface{} got %T", auth["self"])
	}

	name, isOk := self["name"].(string)
	if !isOk {
		t.Errorf("Error getting data[auth][self][name], expected type string got %T", self["name"])
	} else {
		if name != "Bob" {
			t.Errorf("Name was not what was expected, wanted Bob got %s", name)
		}
	}

	email, isOk := self["email"].(string)
	if !isOk {
		t.Errorf("Error getting data[auth][self][email], expected type string got %T", self["email"])
	} else {
		if email != "foo@bar.com" {
			t.Errorf("Email was not what was expected, wanted foo@bar.com got %s", email)
		}
	}

	ID, isOk := self["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[auth][self][ID], expected type int got %T", self["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	variables = map[string]interface{}{
		"token": "bla",
	}
	res = runQuery(query, variables, t)
	if !res.HasErrors() {
		t.Error("Expected errors with invalid JWT got none")
	}
}

func TestQueryBookings(t *testing.T) {
	user := &utils.User{
		Name:  "Bob",
		Email: "foo@bar.com",
		Model: gorm.Model{
			ID: 1,
		},
	}

	bookingsServerResp := `
		{
			"err": "",
			"bookings": [
				{
					"ID": 1
				}
			]
		}
	`

	jwt, err := utils.NewJWT(user)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to bookings server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, bookingsServerResp)
	}))
	defer ts.Close()

	BookingsServer = ts.URL

	query := `
		query ($token: String!) {
			auth(token: $token) {
				self {
					bookings {
						ID
					}
				}
			}
        }
	`
	variables := map[string]interface{}{
		"token": jwt,
	}
	res := runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	auth, isOk := data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}
	self, isOk := auth["self"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][self], expected type map[string]interface{} got %T", auth["self"])
	}
	bookings, isOk := self["bookings"].([]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][self][bookings], expected type []interface{} got %T", self["bookings"])
	}

	if len(bookings) != 1 {
		t.Fatalf("Didn't get right length of bookings, expected 1 got %d", len(bookings))
	}
	booking, isOk := bookings[0].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][self][bookings[0], expected type map[string]interface{} got %T", bookings[0])
	}

	ID, isOk := booking["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[auth][self][bookings][0][ID], expected type int got %T", booking["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	bookingsServerResp = `
		{
			"err": "foobar",
			"bookings": []
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to bookings server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, bookingsServerResp)
	}))
	defer ts.Close()

	BookingsServer = ts.URL

	res = runQuery(query, variables, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when bookings server sent error but got none")
	}
}

func TestQueryBooking(t *testing.T) {
	user := &utils.User{
		Name:  "Bob",
		Email: "foo@bar.com",
		Model: gorm.Model{
			ID: 1,
		},
	}

	bookingsServerResp := `
		{
			"err": "",
			"booking": 
			{
				"ID": 1,
				"start": "2006-01-02T15:04:05Z",
				"end": "2006-01-02T15:04:05Z",
				"hotelId": 1,
				"roomId": 1
			}
		}
	`

	jwt, err := utils.NewJWT(user)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to bookings server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, bookingsServerResp)
	}))
	defer ts.Close()

	BookingsServer = ts.URL

	query := `
		query ($token: String!) {
			auth(token: $token) {
				booking(id: 1) {
					ID
					start
					end
				}
			}
        }
	`
	variables := map[string]interface{}{
		"token": jwt,
	}
	res := runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	auth, isOk := data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}
	booking, isOk := auth["booking"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][booking], expected type map[string]interface{} got %T", auth["booking"])
	}

	ID, isOk := booking["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[auth][booking][ID], expected type int got %T", booking["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}
	start, isOk := booking["start"].(string)
	if !isOk {
		t.Errorf("Error getting data[auth][booking][start], expected type string got %T", booking["start"])
	} else {
		if start != "2006-01-02T15:04:05Z" {
			t.Errorf("Start was not what was expected, wanted 2006-01-02T15:04:05Z got %s", start)
		}
	}
	end, isOk := booking["end"].(string)
	if !isOk {
		t.Errorf("Error getting data[auth][booking][end], expected type string got %T", booking["end"])
	} else {
		if end != "2006-01-02T15:04:05Z" {
			t.Errorf("End was not what was expected, wanted 2006-01-02T15:04:05Z got %s", end)
		}
	}

	query = `
		query ($token: String!) {
			auth(token: $token) {
				booking(id: 1) {
					hotel {
						ID
					}
				}
			}
        }
	`

	hotelsServerResp := `
		{
			"err": "",
			"hotel": 
			{
				"ID": 1
			}
		}
	`

	hotelsTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer hotelsTs.Close()

	HotelsServer = hotelsTs.URL

	res = runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}
	data, isOk = res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	auth, isOk = data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}
	booking, isOk = auth["booking"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][booking], expected type map[string]interface{} got %T", auth["booking"])
	}

	hotel, isOk := booking["hotel"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][booking][hotel], expected type map[string]interface{} got %T", booking["hotel"])
	}
	ID, isOk = hotel["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[auth][booking][hotel][ID], expected type int got %T", hotel["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	hotelsServerResp = `
		{
			"err": "foobar",
			"hotel": null
		}
	`

	hotelsTs = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer hotelsTs.Close()

	HotelsServer = hotelsTs.URL

	res = runQuery(query, variables, t)

	if !res.HasErrors() {
		t.Error("Expected errors with when hotels server sent error but got none")
	}

	query = `
		query ($token: String!) {
			auth(token: $token) {
				booking(id: 1) {
					room {
						ID
					}
				}
			}
        }
	`

	roomsServerResp := `
		{
			"err": "",
			"room": 
			{
				"ID": 1
			}
		}
	`

	roomsTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer roomsTs.Close()

	RoomsServer = roomsTs.URL

	res = runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}
	data, isOk = res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	auth, isOk = data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}
	booking, isOk = auth["booking"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][booking], expected type map[string]interface{} got %T", auth["booking"])
	}

	room, isOk := booking["room"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth][booking][room], expected type map[string]interface{} got %T", booking["hotel"])
	}
	ID, isOk = room["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[auth][booking][room][ID], expected type int got %T", room["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	roomsServerResp = `
		{
			"err": "foobar",
			"hotel": null
		}
	`

	roomsTs = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer roomsTs.Close()

	RoomsServer = roomsTs.URL

	res = runQuery(query, variables, t)

	if !res.HasErrors() {
		t.Error("Expected errors with when roomss server sent error but got none")
	}

	bookingsServerResp = `
		{
			"err": "foobar",
			"bookings": []
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, bookingsServerResp)
	}))
	defer ts.Close()

	res = runQuery(query, variables, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when bookings server sent error but got none")
	}
}

func TestQueryHotel(t *testing.T) {
	query := `
		query {
			hotel(id: 1) {
				ID
				name
				address
				hasCarPark
				checkIn
			}
        }
	`

	hotelsServerResp := `
		{
			"err": "",
			"hotel": 
			{
				"ID": 1,
				"checkIn": "1970-01-01T14:00:00Z",
				"name": "foobar",
				"address": "foobar",
				"hasCarPark": true
			}
		}
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	HotelsServer = ts.URL

	res := runQuery(query, nil, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}
	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	hotel, isOk := data["hotel"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[hotel], expected type map[string]interface{} got %T", data["hotel"])
	}
	ID, isOk := hotel["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[hotel][ID], expected type int got %T", hotel["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}
	name, isOk := hotel["name"].(string)
	if !isOk {
		t.Errorf("Error getting data[hotel][name], expected type string got %T", hotel["name"])
	} else {
		if name != "foobar" {
			t.Errorf("ID was not what was expected, wanted foobar got %s", name)
		}
	}
	address, isOk := hotel["address"].(string)
	if !isOk {
		t.Errorf("Error getting data[hotel][address], expected type string got %T", hotel["address"])
	} else {
		if address != "foobar" {
			t.Errorf("ID was not what was expected, wanted foobar got %s", address)
		}
	}
	checkIn, isOk := hotel["checkIn"].(string)
	if !isOk {
		t.Errorf("Error getting data[hotel][checkIn], expected type string got %T", hotel["checkIn"])
	} else {
		if checkIn != "1970-01-01T14:00:00Z" {
			t.Errorf("ID was not what was expected, wanted foobar got %s", checkIn)
		}
	}

	hotelsServerResp = `
		{
			"err": "",
			"hotel": {
				"checkIn": "bla"
			}
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	HotelsServer = ts.URL

	res = runQuery(query, map[string]interface{}{}, t)

	if !res.HasErrors() {
		t.Errorf("Errors expected from query but none given")
	}
	data, isOk = res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	hotel, isOk = data["hotel"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[hotel], expected type map[string]interface{} got %T", data["hotel"])
	}
	ID, isOk = hotel["ID"].(int)
	if isOk {
		t.Errorf("Error getting data[hotel][ID], expected type <nil> got %T", hotel["ID"])
	}
	name, isOk = hotel["name"].(string)
	if isOk {
		t.Errorf("Error getting data[hotel][name], expected type <nil> got %T", hotel["name"])
	}
	address, isOk = hotel["address"].(string)
	if isOk {
		t.Errorf("Error getting data[hotel][address], expected type <nil> got %T", hotel["address"])
	}
	checkIn, isOk = hotel["checkIn"].(string)
	if isOk {
		t.Errorf("Error getting data[hotel][checkIn], expected type <nil> got %T", hotel["checkIn"])
	}

	hotelsServerResp = `
		{
			"err": "foobar",
			"hotel": {}
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	HotelsServer = ts.URL

	res = runQuery(query, nil, t)

	if !res.HasErrors() {
		t.Error("No errors given when error returned from hotels server")
	}

}

func TestQueryHotels(t *testing.T) {
	hotelsServerResp := `
		{
			"err": "",
			"hotels": [
				{
					"ID": 1
				}
			]
		}
	`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	HotelsServer = ts.URL

	query := `
		query {
			hotels {
				ID
			}
        }
	`
	res := runQuery(query, nil, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	hotels, isOk := data["hotels"].([]interface{})
	if !isOk {
		t.Fatalf("Error getting data[hotels], expected type []interface{} got %T", data["hotels"])
	}

	if len(hotels) != 1 {
		t.Fatalf("Didn't get right length of hotels, expected 1 got %d", len(hotels))
	}
	hotel, isOk := hotels[0].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[hotels[0], expected type map[string]interface{} got %T", hotels[0])
	}

	ID, isOk := hotel["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[hotels][0][ID], expected type int got %T", hotel["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	hotelsServerResp = `
		{
			"err": "foobar",
			"hotels": []
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	BookingsServer = ts.URL

	res = runQuery(query, nil, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when hotels server sent error but got none")
	}
}

func TestQueryRoom(t *testing.T) {
	query := `
		query {
			room(id: 1) {
				ID
				name
				floor
			}
        }
	`

	roomsServerResp := `
		{
			"err": "",
			"room": 
			{
				"ID": 1,
				"name": "foobar",
				"floor": "foobar"
			}
		}
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	res := runQuery(query, nil, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}
	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	room, isOk := data["room"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[room], expected type map[string]interface{} got %T", data["room"])
	}
	ID, isOk := room["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[room][ID], expected type int got %T", room["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}
	name, isOk := room["name"].(string)
	if !isOk {
		t.Errorf("Error getting data[room][name], expected type string got %T", room["name"])
	} else {
		if name != "foobar" {
			t.Errorf("ID was not what was expected, wanted foobar got %s", name)
		}
	}
	floor, isOk := room["floor"].(string)
	if !isOk {
		t.Errorf("Error getting data[room][floor], expected type string got %T", room["floor"])
	} else {
		if floor != "foobar" {
			t.Errorf("ID was not what was expected, wanted foobar got %s", floor)
		}
	}

	roomsServerResp = `
		{
			"err": "",
			"room": {}
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	res = runQuery(query, map[string]interface{}{}, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}
	data, isOk = res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	room, isOk = data["room"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[room], expected type map[string]interface{} got %T", data["room"])
	}
	ID, isOk = room["ID"].(int)
	if isOk {
		t.Errorf("Error getting data[room][ID], expected type <nil> got %T", room["ID"])
	}
	name, isOk = room["name"].(string)
	if isOk {
		t.Errorf("Error getting data[room][name], expected type <nil> got %T", room["name"])
	}
	floor, isOk = room["floor"].(string)
	if isOk {
		t.Errorf("Error getting data[room][floor], expected type <nil> got %T", room["floor"])
	}

	query = `
		query {
			room(id: 1) {
				hotel {
					ID
				}
			}
        }
	`

	roomsServerResp = `
		{
			"err": "",
			"room": 
			{
				"hotelId": 1
			}
		}
	`

	hotelsServerResp := `
		{
			"err": "",
			"hotel": 
			{
				"ID": 1
			}
		}
	`

	hotelsTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer hotelsTs.Close()

	HotelsServer = hotelsTs.URL

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	res = runQuery(query, nil, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}
	data, isOk = res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	room, isOk = data["room"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[room], expected type map[string]interface{} got %T", data["room"])
	}

	hotel, isOk := room["hotel"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[room][hotel], expected type map[string]interface{} got %T", room["hotel"])
	}
	ID, isOk = hotel["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[room][hotel][ID], expected type int got %T", hotel["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	hotelsServerResp = `
		{
			"err": "foobar",
			"hotel": null
		}
	`

	hotelsTs = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer hotelsTs.Close()

	HotelsServer = hotelsTs.URL

	res = runQuery(query, nil, t)

	if !res.HasErrors() {
		t.Error("Expected errors with when hotels server sent error but got none")
	}

	roomsServerResp = `
		{
			"err": "foobar",
			"room": {}
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	res = runQuery(query, nil, t)

	if !res.HasErrors() {
		t.Error("No errors given when error returned from rooms server")
	}

}

func TestQueryRooms(t *testing.T) {
	roomsServerResp := `
		{
			"err": "",
			"rooms": [
				{
					"ID": 1
				}
			]
		}
	`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	query := `
		query {
			rooms {
				ID
			}
        }
	`
	res := runQuery(query, nil, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	rooms, isOk := data["rooms"].([]interface{})
	if !isOk {
		t.Fatalf("Error getting data[rooms], expected type []interface{} got %T", data["rooms"])
	}

	if len(rooms) != 1 {
		t.Fatalf("Didn't get right length of rooms, expected 1 got %d", len(rooms))
	}
	room, isOk := rooms[0].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[rooms[0], expected type map[string]interface{} got %T", rooms[0])
	}

	ID, isOk := room["ID"].(int)
	if !isOk {
		t.Errorf("Error getting data[rooms][0][ID], expected type int got %T", room["ID"])
	} else {
		if ID != 1 {
			t.Errorf("ID was not what was expected, wanted 1 got %d", ID)
		}
	}

	roomsServerResp = `
		{
			"err": "foobar",
			"rooms": []
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	res = runQuery(query, nil, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when rooms server sent error but got none")
	}
}

func TestOpenRoom(t *testing.T) {
	user := &utils.User{
		Name:  "Bob",
		Email: "foo@bar.com",
		Model: gorm.Model{
			ID: 1,
		},
	}

	roomsServerResp := `
		{
			"err": "",
			"success": true
		}
	`

	jwt, err := utils.NewJWT(user)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to rooms server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	query := `
		mutation ($token: String!) {
			auth(token: $token) {
				openRoom(id: 1)
			}
        }
	`
	variables := map[string]interface{}{
		"token": jwt,
	}
	res := runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	auth, isOk := data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}

	openRoom, isOk := auth["openRoom"].(bool)
	if !isOk {
		t.Errorf("Error getting data[auth][openRoom], expected type bool got %T", auth["openRoom"])
	} else {
		if openRoom != true {
			t.Errorf("Success value was not what was expected, wanted true got %t", openRoom)
		}
	}

	roomsServerResp = `
		{
			"err": "foobar",
			"success": false
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to rooms server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, roomsServerResp)
	}))
	defer ts.Close()

	RoomsServer = ts.URL

	res = runQuery(query, variables, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when rooms server sent error but got none")
	}
}

func TestOpenHotel(t *testing.T) {
	user := &utils.User{
		Name:  "Bob",
		Email: "foo@bar.com",
		Model: gorm.Model{
			ID: 1,
		},
	}

	hotelsServerResp := `
		{
			"err": "",
			"success": true
		}
	`

	jwt, err := utils.NewJWT(user)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to hotels server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	HotelsServer = ts.URL

	query := `
		mutation ($token: String!) {
			auth(token: $token) {
				openHotelDoor(id: 1)
			}
        }
	`
	variables := map[string]interface{}{
		"token": jwt,
	}
	res := runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}
	auth, isOk := data["auth"].(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data[auth], expected type map[string]interface{} got %T", data["auth"])
	}

	openHotelDoor, isOk := auth["openHotelDoor"].(bool)
	if !isOk {
		t.Errorf("Error getting data[auth][openHotelDoor], expected type bool got %T", auth["openHotelDoor"])
	} else {
		if openHotelDoor != true {
			t.Errorf("Success value was not what was expected, wanted true got %t", openHotelDoor)
		}
	}

	hotelsServerResp = `
		{
			"err": "foobar",
			"success": false
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == "" {
			t.Errorf("No JWT given to hotels server, got %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		}
		fmt.Fprint(w, hotelsServerResp)
	}))
	defer ts.Close()

	HotelsServer = ts.URL

	res = runQuery(query, variables, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when hotels server sent error but got none")
	}
}

func TestLogin(t *testing.T) {
	authServerResp := `
		{
			"err": "",
			"jwt": "foobar"
		}
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, authServerResp)
	}))
	defer ts.Close()

	AuthServer = ts.URL

	query := `
		mutation ($email: String!, $pass: String!) {
			loginUser(email: $email, pass: $pass)
        }
	`
	variables := map[string]interface{}{
		"email": "foo@bar.com",
		"pass": "foobar",
	}
	res := runQuery(query, variables, t)

	if res.HasErrors() {
		t.Errorf("Errors given from query: %v", res.Errors)
	}

	data, isOk := res.Data.(map[string]interface{})
	if !isOk {
		t.Fatalf("Error getting data, expected type map[string]interface{} got %T", res.Data)
	}

	loginUser, isOk := data["loginUser"].(string)
	if !isOk {
		t.Errorf("Error getting data[loginUser], expected type bool got %T", data["loginUser"])
	} else {
		if loginUser != "foobar" {
			t.Errorf("JWT was not what was expected, wanted foobar got %s", loginUser)
		}
	}

	authServerResp = `
		{
			"err": "foobar",
			"jwt": null
		}
	`

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, authServerResp)
	}))
	defer ts.Close()

	AuthServer = ts.URL

	res = runQuery(query, variables, t)
	if !res.HasErrors() {
		t.Error("Expected errors with when auth server sent error but got none")
	}
}