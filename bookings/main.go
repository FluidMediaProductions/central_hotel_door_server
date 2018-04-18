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
	_"github.com/jinzhu/gorm/dialects/mysql"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

const addr = ":80"

var db *gorm.DB
var jwtSecret []byte

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

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
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

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingResp{
					Err: err.Error(),
				})
				return
			}

			vars := mux.Vars(r)

			id, _ := strconv.Atoi(vars["id"])

			booking := &Booking{}
			err = db.Find(&booking, id).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&BookingResp{
						Err: err.Error(),
					})
					return
				}
			}

			if booking.UserID != claims.User.ID {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingResp{
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
	json.NewEncoder(w).Encode(&BookingResp{
		Err: "no auth header",
	})
}

func getBookingByRoom(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingResp{
					Err: err.Error(),
				})
				return
			}

			vars := mux.Vars(r)

			id, _ := strconv.Atoi(vars["id"])

			booking := &Booking{}
			err = db.Find(&booking, &Booking{
				RoomID: uint(id),
				UserID: claims.User.ID,
			}).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&BookingResp{
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
	json.NewEncoder(w).Encode(&BookingResp{
		Err: "no auth header",
	})
}

func getBookingByHotel(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&BookingResp{
					Err: err.Error(),
				})
				return
			}

			vars := mux.Vars(r)

			id, _ := strconv.Atoi(vars["id"])

			booking := &Booking{}
			err = db.Find(&booking, &Booking{
				HotelID: uint(id),
				UserID:  claims.User.ID,
			}).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&BookingResp{
						Err: "booking not found",
					})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&BookingResp{
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
	json.NewEncoder(w).Encode(&BookingResp{
		Err: "no auth header",
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/bookings").HandlerFunc(getBookings)
	r.Methods("GET").Path("/bookings/{id:[0-9]+}").HandlerFunc(getBooking)
	r.Methods("GET").Path("/bookings/by-room/{id:[0-9]+}").HandlerFunc(getBookingByRoom)
	r.Methods("GET").Path("/bookings/by-hotel/{id:[0-9]+}").HandlerFunc(getBookingByHotel)

	return r
}

func main() {
	viper.SetDefault("DB_HOST", "mysql")
	viper.SetDefault("DB_USER", "travelr")
	viper.SetDefault("DB_NAME", "bookings")

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
		panic(err)
	}
	defer db.Close()

	db.AutoMigrate(&Booking{})

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
