package main

import (
	"encoding/json"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
	"strconv"
	"time"
)

const addr = ":8083"

var db *gorm.DB

type Hotel struct {
	gorm.Model
	Name string `json:"name"`
	Address string `json:"address"`
	Location *utils.Location `json:"location"`
	CheckIn time.Time `json:"checkIn"`
}

type HotelsResp struct {
	Err string `json:"err"`
	Hotels []*Hotel `json:"hotels"`
}

type HotelResp struct {
	Err string `json:"err"`
	Hotel *Hotel `json:"hotel"`
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

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&HotelResp{
			Err: "id not valid",
		})
		return
	}

	hotel := &Hotel{}
	err = db.Find(&hotel, id).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&HotelResp{
			Err: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(&HotelResp{
		Hotel: hotel,
	})
}

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Hotel{})

	r := mux.NewRouter()

	r.Methods("GET").Path("/hotels").HandlerFunc(getHotels)
	r.Methods("GET").Path("/hotel/{id:[0-9]+}").HandlerFunc(getHotel)

	log.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, r)
}
