package main

import (
	"encoding/json"
	"fmt"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const addr = ":80"
const BookingsServer = "http://bookings"

var db *gorm.DB

type Room struct {
	gorm.Model
	Name    string `json:"name"`
	Floor   string `json:"floor"`
	HotelID uint   `json:"hotelId"`
	DoorID  uint   `json:"doorId"`
	ShouldOpen bool `json:"shouldOpen"`
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

func getRooms(w http.ResponseWriter, r *http.Request) {
	rooms := make([]*Room, 0)
	err := db.Find(&rooms).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&RoomsResp{
				Err: "room not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&RoomsResp{
				Err: err.Error(),
			})
			return
		}
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

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&RoomResp{
			Err: "id not valid",
		})
		return
	}

	rooms := make([]*Room, 0)
	err = db.Find(&rooms, &Room{
		HotelID: uint(id),
	}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&RoomsResp{
				Err: "room not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&RoomsResp{
				Err: err.Error(),
			})
			return
		}
	}

	json.NewEncoder(w).Encode(&RoomsResp{
		Rooms: rooms,
	})
}

func openRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&OpenRoomResp{
			Err: "id not valid",
		})
		return
	}

	room := &Room{}
	err = db.Find(&room, id).Error
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

			_, err := utils.VerifyJWT(jwt)
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
				db.Save(&room)

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
	r.Methods("GET").Path("/rooms/{id:[0-9]+}").HandlerFunc(getRoom)
	r.Methods("GET").Path("/rooms/by-hotel/{id:[0-9]+}").HandlerFunc(getRoomsByHotel)
	r.Methods("GET").Path("/rooms/{id:[0-9]+}/open").HandlerFunc(openRoom)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, r))
}
