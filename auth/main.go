package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gopkg.in/hlandau/passlib.v1"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"encoding/json"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
)

const addr = ":8081"

var db *gorm.DB

type JWTResp struct {
	Err string `json:"err"`
	Jwt string `json:"jwt"`
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		return
	}
	defer r.Body.Close()

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&JWTResp{
			Err: err.Error(),
		})
		return
	}

	email, isOk := data["email"].(string)
	if isOk {
		pass, isOk := data["pass"].(string)
		if isOk {
			user := &utils.User{
				Email: email,
			}
			err := db.Where(user).First(user).Error
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: "user not found",
				})
				return
			}

			newHash, err := passlib.Verify(pass, user.Pass)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: "invalid password",
				})
				return
			}

			if newHash != "" {
				user.Pass = newHash
				db.Save(user)
			}

			jwt, err := utils.NewJWT(user)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}
			json.NewEncoder(w).Encode(&JWTResp{
				Jwt: jwt,
			})
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&JWTResp{
		Err: "bad request data",
	})
}

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&utils.User{})

	r := mux.NewRouter()

	r.Methods("POST").Path("/login").HandlerFunc(loginUser)

	log.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, r)
}


