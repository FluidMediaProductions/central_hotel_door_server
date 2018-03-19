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
	"strings"
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

func getRowsForHotels(hotels []*Hotel) *sqlmock.Rows {
	var hotelFieldNames = []string{"id", "name", "address", "location", "has_car_park", "check_in", "should_door_open", "created_at", "updated_at", "deleted_at"}
	rows := sqlmock.NewRows(hotelFieldNames)
	for _, h := range hotels {
		rows = rows.AddRow(h.ID, h.Name, h.Address, h.Location, h.HasCarPark, h.CheckIn, h.ShouldDoorOpen, h.CreatedAt, h.UpdatedAt, h.DeletedAt)
	}
	return rows
}

func TestGetHotels(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/hotels", nil)
	w := httptest.NewRecorder()

	hotels := []*Hotel{
		{
			Model: gorm.Model{
				ID:        uint(1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: "foo",
		},
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL")).
		WillReturnRows(getRowsForHotels(hotels))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&HotelsResp{
		Hotels: hotels,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 error got %s", resp.Status)
	}

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&HotelsResp{
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

func TestGetHotel(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/hotels/1", nil)
	w := httptest.NewRecorder()

	hotel := &Hotel{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "foo",
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&HotelResp{
		Hotel: hotel,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&HotelResp{
		Err: "hotel not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&HotelResp{
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

func TestOpenHotel(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	req := httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
	w := httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{}))

	router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err := json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Err: "hotel not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnError(errors.New("foobar"))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
	w = httptest.NewRecorder()

	hotel := &Hotel{
		Model: gorm.Model{
			ID:        uint(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "foo",
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Err: "no auth header",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
	req.Header.Set("Authorization", "Bearer bla")
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
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

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
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

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Err: "invalid character 'b' looking for beginning of value",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
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

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
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

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	sqlMock.ExpectExec(fixedFullRe("UPDATE `hotels` SET `created_at` = ?, `updated_at` = ?, `deleted_at` = ?, `name` = ?," +
		" `address` = ?, `check_in` = ?, `has_car_park` = ?, `should_door_open` = ? WHERE `hotels`.`deleted_at` IS NULL AND `hotels`.`id` = ?")).
		WillReturnError(errors.New("foobar"))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Err: "foobar",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
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

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	sqlMock.ExpectExec(fixedFullRe("UPDATE `hotels` SET `created_at` = ?, `updated_at` = ?, `deleted_at` = ?, `name` = ?," +
		" `address` = ?, `check_in` = ?, `has_car_park` = ?, `should_door_open` = ? WHERE `hotels`.`deleted_at` IS NULL AND `hotels`.`id` = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
		Success: true,
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	req = httptest.NewRequest("GET", "http://a/hotels/1/open", nil)
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

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `hotels` WHERE `hotels`.`deleted_at` IS NULL AND ((`hotels`.`id` = 1))")).
		WillReturnRows(getRowsForHotels([]*Hotel{hotel}))

	BookingsServer = ts.URL
	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&OpenHotelResp{
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
