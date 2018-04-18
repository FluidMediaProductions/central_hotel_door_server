package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/spf13/viper"
)

const addr = ":80"

var db *gorm.DB

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

	room.ShouldOpen = true
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

func openRoomSuccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	room := &Room{}
	err := db.Find(&room, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
				Err: "room not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
				Err: err.Error(),
			})
			return
		}
	}

	room.ShouldOpen = false
	err = db.Save(&room).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
		Success: true,
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/rooms").HandlerFunc(getRooms)
	r.Methods("GET").Path("/rooms/{id:[0-9]+}").HandlerFunc(getRoom)
	r.Methods("GET").Path("/rooms/by-hotel/{id:[0-9]+}").HandlerFunc(getRoomsByHotel)
	r.Methods("GET").Path("/rooms/{id:[0-9]+}/open").HandlerFunc(openRoom)
	r.Methods("GET").Path("/rooms/{id:[0-9]+}/open-success").HandlerFunc(openRoomSuccess)

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
