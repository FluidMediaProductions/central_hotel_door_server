package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/spf13/viper"
	"github.com/go-sql-driver/mysql"
)

const addr = ":80"

var BookingsServer = "http://bookings"

var db *gorm.DB
var jwtSecret []byte

type Room struct {
	gorm.Model
	Name       string `json:"name"`
	Floor      string `json:"floor"`
	HotelID    uint   `json:"hotelId"`
	ShouldOpen bool   `json:"shouldOpen"`
}

type RoomsResp struct {
	Err   string  `json:"err"`
	Rooms []*Room `json:"rooms"`
}

type RoomResp struct {
	Err  string `json:"err"`
	Room *Room  `json:"room"`
}

type OpenRoomResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
}

type OpenRoomSuccessResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
}

func getRooms(w http.ResponseWriter, r *http.Request) {
	rooms := make([]*Room, 0)
	err := db.Find(&rooms).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomsResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&RoomsResp{
		Rooms: rooms,
	})
}

func getRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	room := &Room{}
	err := db.Find(&room, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&RoomResp{
				Err: "room not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&RoomResp{
				Err: err.Error(),
			})
			return
		}
	}

	json.NewEncoder(w).Encode(&RoomResp{
		Room: room,
	})
}

func getRoomsByHotel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	rooms := make([]*Room, 0)
	err := db.Find(&rooms, &Room{
		HotelID: uint(id),
	}).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomsResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&RoomsResp{
		Rooms: rooms,
	})
}

func openRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	room := &Room{}
	err := db.Find(&room, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&OpenRoomResp{
				Err: "room not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&OpenRoomResp{
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
				json.NewEncoder(w).Encode(&OpenRoomResp{
					Err: err.Error(),
				})
				return
			}

			req, err := http.NewRequest("GET", BookingsServer+fmt.Sprintf("/bookings/by-room/%d", id), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&OpenRoomResp{
					Err: err.Error(),
				})
				return
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

			resp, err := utils.GetJson(req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&OpenRoomResp{
					Err: err.Error(),
				})
				return
			}
			respErr, isOk := resp["err"].(string)
			if isOk {
				if respErr != "" {
					json.NewEncoder(w).Encode(&OpenRoomResp{
						Err: respErr,
					})
					return
				}
			}

			_, isOk = resp["booking"].(map[string]interface{})
			if isOk {
				room.ShouldOpen = true
				err := db.Save(&room).Error
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&OpenRoomResp{
						Err: err.Error(),
					})
					return
				}

				json.NewEncoder(w).Encode(&OpenRoomResp{
					Success: true,
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&OpenRoomResp{
				Err: "unknown",
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&OpenRoomResp{
		Err: "no auth header",
	})
}

func openRoomSuccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])
	hotelId, _ := strconv.Atoi(vars["hotel"])

	room := &Room{}
	err := db.Find(&room, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&OpenRoomResp{
				Err: "room not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&OpenRoomResp{
				Err: err.Error(),
			})
			return
		}
	}

	if room.HotelID != uint(hotelId) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&OpenRoomResp{
			Err: "room not in hotel",
		})
		return
	}

	room.ShouldOpen = false
	err = db.Save(&room).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&OpenRoomResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&OpenRoomResp{
		Success: true,
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/rooms").HandlerFunc(getRooms)
	r.Methods("GET").Path("/rooms/{id:[0-9]+}").HandlerFunc(getRoom)
	r.Methods("GET").Path("/rooms/by-hotel/{id:[0-9]+}").HandlerFunc(getRoomsByHotel)
	r.Methods("GET").Path("/rooms/{id:[0-9]+}/open").HandlerFunc(openRoom)
	r.Methods("GET").Path("/rooms/by-hotel/{hotel:[0-9]+}/{id:[0-9]+}/open-success").HandlerFunc(openRoomSuccess)

	return r
}

func main() {
	viper.SetDefault("DB_HOST", "mysql")
	viper.SetDefault("DB_USER", "travelr")
	viper.SetDefault("DB_NAME", "rooms")

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

	db.AutoMigrate(&Room{})

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
