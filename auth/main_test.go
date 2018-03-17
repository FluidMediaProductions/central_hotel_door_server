package main

import (
	"net/http/httptest"
	"github.com/jinzhu/gorm"
	"testing"
	"fmt"
	"regexp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"strings"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"encoding/json"
	"gopkg.in/hlandau/passlib.v1"
	"bytes"
	"io/ioutil"
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

func getRowsForUsers(users []*utils.User) *sqlmock.Rows {
	var userFieldNames = []string{"id", "name", "email", "pass", "created_at", "updated_at", "deleted_at"}
	rows := sqlmock.NewRows(userFieldNames)
	for _, u := range users {
		rows = rows.AddRow(u.ID, u.Name, u.Email, u.Pass, u.CreatedAt, u.UpdatedAt, u.DeletedAt)
	}
	return rows
}

func TestLogin(t *testing.T) {
	sqlMock, dbMock := newDB(t)

	oldDb := db
	db = dbMock

	reader := strings.NewReader(`{"email": "foo@bar.com", "pass": "foobar"}`)
	req := httptest.NewRequest("POST", "http://a/login", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	hash, _ := passlib.Hash("foobar")
	user := &utils.User{
		Email: "foo@bar.com",
		Pass: hash,
	}

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL AND ((`users`.`email` = ?))" +
		" ORDER BY `users`.`id` ASC LIMIT 1")).
		WithArgs("foo@bar.com").
		WillReturnRows(getRowsForUsers([]*utils.User{user}))

	router().ServeHTTP(w, req)

	resp := w.Result()

	var respData *JWTResp
	err := json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		t.Errorf("Error decoding JSON: %v", err)
	}

	if respData.Err != "" {
		t.Error("Got error on login")
	}

	reader = strings.NewReader(`b`)
	req = httptest.NewRequest("POST", "http://a/login", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expBody := &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&JWTResp{
		Err: "invalid character 'b' looking for beginning of value",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	reader = strings.NewReader(`{"foo": "bar"}`)
	req = httptest.NewRequest("POST", "http://a/login", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&JWTResp{
		Err: "bad request data",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	reader = strings.NewReader(`{"email": "foo@bar.com", "pass": "foobar"}`)
	req = httptest.NewRequest("POST", "http://a/login", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL AND ((`users`.`email` = ?))" +
		" ORDER BY `users`.`id` ASC LIMIT 1")).
		WithArgs("foo@bar.com").
		WillReturnRows(getRowsForUsers([]*utils.User{}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&JWTResp{
		Err: "user not found",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	reader = strings.NewReader(`{"email": "foo@bar.com", "pass": "foobar2"}`)
	req = httptest.NewRequest("POST", "http://a/login", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	sqlMock.ExpectQuery(fixedFullRe("SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL AND ((`users`.`email` = ?))" +
		" ORDER BY `users`.`id` ASC LIMIT 1")).
		WithArgs("foo@bar.com").
		WillReturnRows(getRowsForUsers([]*utils.User{user}))

	router().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	expBody = &bytes.Buffer{}
	err = json.NewEncoder(expBody).Encode(&JWTResp{
		Err: "invalid password",
	})
	if err != nil {
		t.Fatalf("Error creating test JSON: %v", err)
	}

	if string(body) != string(expBody.Bytes()) {
		t.Errorf("Response not what was expected, got %s wanted %s", string(body), string(expBody.Bytes()))
	}

	db = oldDb
}
