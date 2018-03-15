package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
	"strconv"
)

const addr = ":80"

var db *gorm.DB

type Room struct {
	gorm.Model
	Name string `json:"name"`
	Floor string `json:"floor"`
	HotelID uint `json:"hotelId"`
	DoorID uint `json:"doorId"`
}

type RoomsResp struct {
	Err string `json:"err"`
	Rooms []*Room `json:"rooms"`
}

type RoomResp struct {
	Err string `json:"err"`
	Room *Room `json:"room"`
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

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&RoomResp{
			Err: "id not valid",
		})
		return
	}

	room := &Room{}
	err = db.Find(&room, id).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&RoomResp{
		Room: room,
	})
}

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Room{})

	r := mux.NewRouter()

	r.Methods("GET").Path("/rooms").HandlerFunc(getRooms)
	r.Methods("GET").Path("/room/{id:[0-9]+}").HandlerFunc(getRoom)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, r))
}
