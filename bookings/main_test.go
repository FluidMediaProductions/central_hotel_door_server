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
	"testing"
	"time"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/jinzhu/gorm"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"log"
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

func getRowsForBookings(bookings []*Booking) *sqlmock.Rows {
	var bookingFieldNames = []string{"id", "hotel_id", "user_id", "room_id", "start", "end", "created_at", "updated_at", "deleted_at"}
	rows := sqlmock.NewRows(bookingFieldNames)
	for _, b := range bookings {
		rows = rows.AddRow(b.ID, b.HotelID, b.UserID, b.RoomID, b.Start, b.End, b.CreatedAt, b.UpdatedAt, b.DeletedAt)
	}
	return rows
}

func TestGetBookings(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	user := &utils.User{
		Model: gorm.Model{
			ID: 1,
		},
	}
	jwt, err := utils.NewJWT(user)
	if err != nil {
		log.Fatalf("Error making JWT: %v", err)
	}

	req := httptest.NewRequest("GET", "http://a/bookings", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()

	bookings := []*Booking{
		{
			Model: gorm.Model{
				ID:        uint(1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID:  1,
			HotelID: 1,
		},
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?))")).
		WithArgs(1).
		WillReturnRows(getRowsForBookings(bookings))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingsResp{
		Bookings: bookings,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?))")).
		WithArgs(1).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 error got %s", resp.Status)
	}

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingsResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings", nil)
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingsResp{
		Err: "no auth header",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings", nil)
	req.Header.Set("Authorization", "Bearer a")
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingsResp{
		Err: "token contains an invalid number of segments",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}

func TestGetBooking(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	user := &utils.User{
		Model: gorm.Model{
			ID: 1,
		},
	}
	jwt, err := utils.NewJWT(user)
	if err != nil {
		log.Fatalf("Error making JWT: %v", err)
	}

	req := httptest.NewRequest("GET", "http://a/bookings/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()

	booking := &Booking{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:  1,
		HotelID: 1,
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`id` = 1))")).
		WillReturnRows(getRowsForBookings([]*Booking{booking}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Booking: booking,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`id` = 1))")).
		WillReturnRows(getRowsForBookings([]*Booking{}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "booking not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`id` = 1))")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/1", nil)
	req.Header.Set("Authorization", "Bearer a")
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "token contains an invalid number of segments",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/1", nil)
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "no auth header",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	booking = &Booking{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:  2,
		HotelID: 1,
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`id` = 1))")).
		WillReturnRows(getRowsForBookings([]*Booking{booking}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "booking not owned by user",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}

func TestGetBookingByHotel(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	user := &utils.User{
		Model: gorm.Model{
			ID: 1,
		},
	}
	jwt, err := utils.NewJWT(user)
	if err != nil {
		log.Fatalf("Error making JWT: %v", err)
	}

	req := httptest.NewRequest("GET", "http://a/bookings/by-hotel/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()

	booking := &Booking{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:  1,
		HotelID: 1,
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?)"+
		" AND (`bookings`.`hotel_id` = ?))")).
		WithArgs(1, 1).
		WillReturnRows(getRowsForBookings([]*Booking{booking}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Booking: booking,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-hotel/2", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?)"+
		" AND (`bookings`.`hotel_id` = ?))")).
		WithArgs(1, 2).
		WillReturnRows(getRowsForBookings([]*Booking{}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "booking not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-hotel/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?)"+
		" AND (`bookings`.`hotel_id` = ?))")).
		WithArgs(1, 1).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-hotel/1", nil)
	req.Header.Set("Authorization", "Bearer a")
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "token contains an invalid number of segments",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-hotel/1", nil)
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "no auth header",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}

func TestGetBookingByRoom(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	user := &utils.User{
		Model: gorm.Model{
			ID: 1,
		},
	}
	jwt, err := utils.NewJWT(user)
	if err != nil {
		log.Fatalf("Error making JWT: %v", err)
	}

	req := httptest.NewRequest("GET", "http://a/bookings/by-room/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()

	booking := &Booking{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:  1,
		HotelID: 1,
		RoomID:  1,
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?)"+
		" AND (`bookings`.`room_id` = ?))")).
		WithArgs(1, 1).
		WillReturnRows(getRowsForBookings([]*Booking{booking}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Booking: booking,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-room/2", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?)"+
		" AND (`bookings`.`room_id` = ?))")).
		WithArgs(1, 2).
		WillReturnRows(getRowsForBookings([]*Booking{}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "booking not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-room/1", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `bookings` WHERE `bookings`.`deleted_at` IS NULL AND ((`bookings`.`user_id` = ?)"+
		" AND (`bookings`.`room_id` = ?))")).
		WithArgs(1, 1).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-room/1", nil)
	req.Header.Set("Authorization", "Bearer a")
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "token contains an invalid number of segments",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/bookings/by-room/1", nil)
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&BookingResp{
		Err: "no auth header",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}
