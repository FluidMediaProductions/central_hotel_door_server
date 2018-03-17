package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const addr = ":80"

var db *gorm.DB

type Booking struct {
	gorm.Model
	UserID  uint      `json:"userId"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	HotelID uint      `json:"hotelId"`
	RoomID  uint      `json:"roomId"`
}

type BookingsResp struct {
	Err      string     `json:"err"`
	Bookings []*Booking `json:"bookings"`
}

type BookingResp struct {
	Err     string   `json:"err"`
	Booking *Booking `json:"booking"`
}

func getBookings(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: err.Error(),
					})
					return
				}
			}

			bookings := make([]*Booking, 0)
			err = db.Where(&Booking{
				UserID: claims.User.ID,
			}).Find(&bookings).Error
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			json.NewEncoder(w).Encode(&BookingsResp{
				Bookings: bookings,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&BookingsResp{
		Err: "no auth header",
	})
}

func getBooking(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			vars := mux.Vars(r)

			id, err := strconv.Atoi(vars["id"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: "id not valid",
				})
				return
			}

			booking := &Booking{}
			err = db.Find(&booking, id).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: err.Error(),
					})
					return
				}
			}

			if booking.UserID != claims.User.ID {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: "booking not owned by user",
				})
				return
			}

			json.NewEncoder(w).Encode(&BookingResp{
				Booking: booking,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&BookingsResp{
		Err: "no auth header",
	})
}

func getBookingByRoom(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			vars := mux.Vars(r)

			id, err := strconv.Atoi(vars["id"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: "id not valid",
				})
				return
			}

			booking := &Booking{}
			err = db.Find(&booking, &Booking{
				RoomID: uint(id),
				UserID: claims.User.ID,
			}).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: err.Error(),
					})
					return
				}
			}

			json.NewEncoder(w).Encode(&BookingResp{
				Booking: booking,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&BookingsResp{
		Err: "no auth header",
	})
}

func getBookingByHotel(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			vars := mux.Vars(r)

			id, err := strconv.Atoi(vars["id"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: "id not valid",
				})
				return
			}

			booking := &Booking{}
			err = db.Find(&booking, &Booking{
				HotelID: uint(id),
				UserID:  claims.User.ID,
			}).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&BookingsResp{
						Err: err.Error(),
					})
					return
				}
			}

			json.NewEncoder(w).Encode(&BookingResp{
				Booking: booking,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&BookingsResp{
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

	db.AutoMigrate(&Booking{})

	r := mux.NewRouter()

	r.Methods("GET").Path("/bookings").HandlerFunc(getBookings)
	r.Methods("GET").Path("/bookings/{id:[0-9]+}").HandlerFunc(getBooking)
	r.Methods("GET").Path("/bookings/by-room/{id:[0-9]+}").HandlerFunc(getBookingByRoom)
	r.Methods("GET").Path("/bookings/by-hotel/{id:[0-9]+}").HandlerFunc(getBookingByHotel)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, r))
}
