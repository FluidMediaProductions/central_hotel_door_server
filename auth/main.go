package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"gopkg.in/hlandau/passlib.v1"
	"strings"
)

const addr = ":80"

var db *gorm.DB
var jwtSecret []byte

type JWTResp struct {
	Err string `json:"err"`
	Jwt string `json:"jwt"`
}

type ChangePasswordResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
}

type UpdateUserResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
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
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&JWTResp{
						Err: "user not found",
					})
					return
				}
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
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

			jwt, err := utils.NewJWT(user, jwtSecret)
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

func changePassword(w http.ResponseWriter, r *http.Request) {
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

	pass, isOk := data["pass"].(string)
	if isOk {

		authHeaders, isOk := r.Header["Authorization"]
		if isOk {
			if len(authHeaders) > 0 {
				authHeader := authHeaders[0]
				jwt := strings.TrimPrefix(authHeader, "Bearer ")

				claims, err := utils.VerifyJWT(jwt, jwtSecret)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				user := &utils.User{}
				err = db.First(user, claims.User.ID).Error
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(&ChangePasswordResp{
							Err: "user not found",
						})
						return
					}
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				newHash, err := passlib.Hash(pass)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				user.Pass = newHash
				err = db.Save(user).Error
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				json.NewEncoder(w).Encode(&ChangePasswordResp{
					Success: true,
				})
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(&ChangePasswordResp{
			Err: "no auth header",
		})
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&ChangePasswordResp{
		Err: "bad request data",
	})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
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
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&UpdateUserResp{
					Err: err.Error(),
				})
				return
			}

			user := &utils.User{}
			err = db.First(user, claims.User.ID).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&UpdateUserResp{
						Err: "user not found",
					})
					return
				}
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&UpdateUserResp{
					Err: err.Error(),
				})
				return
			}

			email, isOk := data["email"].(string)
			if isOk {
				user.Email = email
			}
			name, isOk := data["name"].(string)
			if isOk {
				user.Name = name
			}

			err = db.Save(user).Error
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&UpdateUserResp{
					Err: err.Error(),
				})
				return
			}

			json.NewEncoder(w).Encode(&UpdateUserResp{
				Success: true,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&UpdateUserResp{
		Err: "no auth header",
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("POST").Path("/login").HandlerFunc(loginUser)
	r.Methods("POST").Path("/changePassword").HandlerFunc(changePassword)
	r.Methods("POST").Path("/updateUser").HandlerFunc(updateUser)

	return r
}

func main() {
	viper.SetDefault("DB_HOST", "mysql")
	viper.SetDefault("DB_USER", "travelr")
	viper.SetDefault("DB_NAME", "auth")

	viper.SetEnvPrefix("TRAVELR")
	viper.AutomaticEnv()

	dbHost := viper.GetString("DB_HOST")
	dbUser := viper.GetString("DB_USER")
	dbPass := viper.GetString("DB_PASS")
	dbName := viper.GetString("DB_NAME")

	jwtSecret = []byte(viper.GetString("JWT_SECRET"))

	config := &mysql.Config{Addr: dbHost, Net: "tcp", User: dbUser, Passwd: dbPass, DBName: dbName, ParseTime: true}

	log.Printf("Connecting to database with DSN: %s\n", config.FormatDSN())
	var err error
	db, err = gorm.Open("mysql", config.FormatDSN())
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&utils.User{})

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
