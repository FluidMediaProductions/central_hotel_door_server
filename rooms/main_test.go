package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/jinzhu/gorm"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

func newDB(t *testing.T) (sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("can't create sqlmock: %s", err)
	}

	gormDB, gerr := gorm.Open("mysql", db)
	if gerr != nil {
		t.Fatalf("can't open gorm connection: %s", err)
	}

	return mock, gormDB.Set("gorm:update_column", true)
}

func getRowsForRooms(rooms []*Room) *sqlmock.Rows {
	var roomFieldNames = []string{"id", "name", "floor", "hotel_id", "door_id", "should_open", "created_at", "updated_at", "deleted_at"}
	rows := sqlmock.NewRows(roomFieldNames)
	for _, r := range rooms {
		rows = rows.AddRow(r.ID, r.Name, r.Floor, r.HotelID, r.DoorID, r.ShouldOpen, r.CreatedAt, r.UpdatedAt, r.DeletedAt)
	}
	return rows
}

func TestGetRooms(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/rooms", nil)
	w := httptest.NewRecorder()

	rooms := []*Room{
		{
			Model: gorm.Model{
				ID:        uint(1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:  "1",
			Floor: "1",
		},
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL")).
		WillReturnRows(getRowsForRooms(rooms))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&RoomsResp{
		Rooms: rooms,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 error got %s", resp.Status)
	}

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&RoomsResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}

func TestGetRoom(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/rooms/1", nil)
	w := httptest.NewRecorder()

	room := &Room{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:  "1",
		Floor: "1",
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&RoomResp{
		Room: room,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&RoomResp{
		Err: "room not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&RoomResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}

func TestGetRoomByHotel(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/rooms/by-hotel/1", nil)
	w := httptest.NewRecorder()

	rooms := []*Room{
		{
			Model: gorm.Model{
				ID:        uint(1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "1",
			Floor:   "1",
			HotelID: 1,
		},
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`hotel_id` = ?))")).
		WithArgs(1).
		WillReturnRows(getRowsForRooms(rooms))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&RoomsResp{
		Rooms: rooms,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/by-hotel/1", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`hotel_id` = ?))")).
		WithArgs(1).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&RoomsResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}

func TestOpenRoom(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	w := httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "room not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	w = httptest.NewRecorder()

	room := &Room{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "1",
		Floor:   "1",
		HotelID: 1,
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "no auth header",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	req.Header.Set("Authorization", "Bearer bla")
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "token contains an invalid number of segments",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	user := &utils.User{
		Model: gorm.Model{
			ID: 1,
		},
		Name: "Bob",
	}
	jwt, err := utils.NewJWT(user)
	if err != nil {
		t.Fatalf("Failed to make JWT: %v", err)
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != jwt {
			t.Errorf("Wrong JWT given to bookings server, got %s expected %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), jwt)
		}
		fmt.Fprint(w, `
			bla
		`)
	}))
	defer ts.Close()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "invalid character 'b' looking for beginning of value",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != jwt {
			t.Errorf("Wrong JWT given to bookings server, got %s expected %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), jwt)
		}
		fmt.Fprint(w, `
			{
				"err": "foobar"
			}
		`)
	}))
	defer ts.Close()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != jwt {
			t.Errorf("Wrong JWT given to bookings server, got %s expected %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), jwt)
		}
		fmt.Fprint(w, `
			{
				"booking": {
				}
			}
		`)
	}))
	defer ts.Close()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	sqlMock.ExpectExec(fixedFullRe("UPDATE `rooms` SET `created_at` = ?, `updated_at` = ?, `deleted_at` = ?, `name` = ?," +
		" `floor` = ?, `hotel_id` = ?, `door_id` = ?, `should_open` = ? WHERE `rooms`.`deleted_at` IS NULL AND `rooms`.`id` = ?")).
		WillReturnError(errors.New("foobar"))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != jwt {
			t.Errorf("Wrong JWT given to bookings server, got %s expected %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), jwt)
		}
		fmt.Fprint(w, `
			{
				"booking": {
				}
			}
		`)
	}))
	defer ts.Close()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	sqlMock.ExpectExec(fixedFullRe("UPDATE `rooms` SET `created_at` = ?, `updated_at` = ?, `deleted_at` = ?, `name` = ?," +
		" `floor` = ?, `hotel_id` = ?, `door_id` = ?, `should_open` = ? WHERE `rooms`.`deleted_at` IS NULL AND `rooms`.`id` = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Success: true,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/rooms/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != jwt {
			t.Errorf("Wrong JWT given to bookings server, got %s expected %s",
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), jwt)
		}
		fmt.Fprint(w, `
			{
				"foo": "bar"
			}
		`)
	}))
	defer ts.Close()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `rooms` WHERE `rooms`.`deleted_at` IS NULL AND ((`rooms`.`id` = 1))")).
		WillReturnRows(getRowsForRooms([]*Room{room}))

	sqlMock.ExpectExec(fixedFullRe("UPDATE `rooms` SET `created_at` = ?, `updated_at` = ?, `deleted_at` = ?, `name` = ?," +
		" `floor` = ?, `hotel_id` = ?, `door_id` = ?, `should_open` = ? WHERE `rooms`.`deleted_at` IS NULL AND `rooms`.`id` = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenRoomResp{
		Err: "unknown",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}
