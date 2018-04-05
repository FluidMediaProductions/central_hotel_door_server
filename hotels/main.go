package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/go-sql-driver/mysql"
)

const addr = ":80"

var BookingsServer = "http://bookings"

var db *gorm.DB
var jwtSecret []byte

type Hotel struct {
	gorm.Model
	Name           string          `json:"name"`
	Address        string          `json:"address"`
	Location       *utils.Location `json:"location"`
	CheckIn        time.Time       `json:"checkIn"`
	HasCarPark     bool            `json:"hasCarPark"`
	ShouldDoorOpen bool            `json:"shouldDoorOpen"`
}

type HotelsResp struct {
	Err    string   `json:"err"`
	Hotels []*Hotel `json:"hotels"`
}

type HotelResp struct {
	Err   string `json:"err"`
	Hotel *Hotel `json:"hotel"`
}

type OpenHotelResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
}

func getHotels(w http.ResponseWriter, r *http.Request) {
	hotels := make([]*Hotel, 0)
	err := db.Find(&hotels).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&HotelsResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&HotelsResp{
		Hotels: hotels,
	})
}

func getHotel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	hotel := &Hotel{}
	err := db.Find(&hotel, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&HotelResp{
				Err: "hotel not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&HotelResp{
				Err: err.Error(),
			})
			return
		}
	}

	json.NewEncoder(w).Encode(&HotelResp{
		Hotel: hotel,
	})
}

func openHotel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	hotel := &Hotel{}
	err := db.Find(&hotel, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&OpenHotelResp{
				Err: "hotel not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&OpenHotelResp{
				Err: err.Error(),
			})
			return
		}
	}

	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			_, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&OpenHotelResp{
					Err: err.Error(),
				})
				return
			}

			req, err := http.NewRequest("GET", BookingsServer+fmt.Sprintf("/bookings/by-hotel/%d", id), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&OpenHotelResp{
					Err: err.Error(),
				})
				return
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

			resp, err := utils.GetJson(req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&OpenHotelResp{
					Err: err.Error(),
				})
				return
			}
			respErr, isOk := resp["err"].(string)
			if isOk {
				if respErr != "" {
					json.NewEncoder(w).Encode(&OpenHotelResp{
						Err: respErr,
					})
					return
				}
			}

			_, isOk = resp["booking"].(map[string]interface{})
			if isOk {
				hotel.ShouldDoorOpen = true
				err := db.Save(&hotel).Error
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&OpenHotelResp{
						Err: err.Error(),
					})
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&OpenHotelResp{
					Success: true,
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&OpenHotelResp{
				Err: "unknown",
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&OpenHotelResp{
		Err: "no auth header",
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/hotels").HandlerFunc(getHotels)
	r.Methods("GET").Path("/hotels/{id:[0-9]+}").HandlerFunc(getHotel)
	r.Methods("GET").Path("/hotels/{id:[0-9]+}/open").HandlerFunc(openHotel)

	return r
}

func main() {
	viper.SetDefault("DB_HOST", "mysql")
	viper.SetDefault("DB_USER", "travelr")
	viper.SetDefault("DB_NAME", "hotels")

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

	db.AutoMigrate(&Hotel{})

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
