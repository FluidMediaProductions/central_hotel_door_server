package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"context"
)

const addr = ":80"

var db *dgo.Dgraph
var jwtSecret []byte

type Booking struct {
	ID      string    `json:"uid"`
	UserID  string    `json:"userId"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	HotelID string    `json:"hotelId"`
	RoomID  string    `json:"roomId"`
}

type BookingsResp struct {
	Err      string     `json:"err"`
	Bookings []*Booking `json:"bookings"`
}

type BookingResp struct {
	Err     string   `json:"err"`
	Booking *Booking `json:"booking"`
}

type bookingQuery struct {
	Bookings []struct {
		Start *time.Time `json:"booking.start"`
		End  *time.Time `json:"booking.end"`
		User  []struct{
			ID    string `json:"uid"`
		} `json:"booking.user"`
		Hotel  []struct{
			ID    string `json:"uid"`
		} `json:"booking.hotel"`
		Room  []struct{
			ID    string `json:"uid"`
		} `json:"booking.room"`
		ID    string `json:"uid"`
	} `json:"bookings"`
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

			ctx := context.Background()
			txn := db.NewTxn()

			variables := map[string]string{"$userID": claims.User.ID}
			q := `query q($userID: string) {
                    var (func: uid($userID)) {
		              u as uid
	                }
                    bookings(func: has(booking)) @cascade {
                      uid
                      booking.start
                      booking.end
                      booking.hotel {
                        uid
                      }
                      booking.room {
                        uid
                      }
                      booking.user @filter(uid(u)) {
                        uid
                      }
	                }
                  }`

			resp, err := txn.QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}
			var bookings bookingQuery
			err = json.Unmarshal(resp.GetJson(), &bookings)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			outBookings := make([]*Booking, 0)
			for _, booking := range bookings.Bookings {
				outBooking := &Booking{
					ID: booking.ID,
					HotelID: booking.Hotel[0].ID,
					RoomID: booking.Room[0].ID,
					Start: *booking.Start,
					End: *booking.End,
					UserID: booking.User[0].ID,
				}
				outBookings = append(outBookings, outBooking)
			}

			json.NewEncoder(w).Encode(&BookingsResp{
				Bookings: outBookings,
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

			id := vars["id"]

			ctx := context.Background()
			txn := db.NewTxn()

			variables := map[string]string{"$id": id, "$user": claims.User.ID}
			q := `query q($id: uid, $user: uid) {
                    var (func: uid($user)) {
		              u as uid
	                }
                    bookings(func: uid($id)) @filter(has(booking)) @cascade {
                      uid
                      booking.start
                      booking.end
                      booking.hotel {
                        uid
                      }
                      booking.room {
                        uid
                      }
                      booking.user @filter(uid(u)) {
                        uid
                      }
	                }
                  }`

			resp, err := txn.QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}
			var bookings bookingQuery
			err = json.Unmarshal(resp.GetJson(), &bookings)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}


			if len(bookings.Bookings) == 0 {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&BookingResp{
					Err: "booking not found",
				})
				return
			}

			booking := bookings.Bookings[0]
				outBooking := &Booking{
					ID: booking.ID,
					HotelID: booking.Hotel[0].ID,
					RoomID: booking.Room[0].ID,
					Start: *booking.Start,
					End: *booking.End,
					UserID: booking.User[0].ID,
				}

			json.NewEncoder(w).Encode(&BookingResp{
				Booking: outBooking,
			})
			return

			json.NewEncoder(w).Encode(&BookingResp{
				Booking: outBooking,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&BookingResp{
		Err: "no auth header",
	})
}

func getBookingsByRoom(w http.ResponseWriter, r *http.Request) {
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

			id, _ := vars["id"]

			ctx := context.Background()
			txn := db.NewTxn()

			variables := map[string]string{"$id": id, "$user": claims.User.ID}
			q := `query q($id: uid, $user: uid) {
                    var (func: uid($user)) {
		              u as uid
	                }
                    var (func: uid($id)) {
		              r as uid
	                }
                    bookings(func: has(booking)) @cascade {
                      uid
                      booking.start
                      booking.end
                      booking.hotel {
                        uid
                      }
                      booking.room @filter(uid(r)) {
                        uid
                      }
                      booking.user @filter(uid(u)) {
                        uid
                      }
	                }
                  }`

			resp, err := txn.QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}
			var bookings bookingQuery
			err = json.Unmarshal(resp.GetJson(), &bookings)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			outBookings := make([]*Booking, 0)
			for _, booking := range bookings.Bookings {
				outBooking := &Booking{
					ID: booking.ID,
					HotelID: booking.Hotel[0].ID,
					RoomID: booking.Room[0].ID,
					Start: *booking.Start,
					End: *booking.End,
					UserID: booking.User[0].ID,
				}
				outBookings = append(outBookings, outBooking)
			}

			json.NewEncoder(w).Encode(&BookingsResp{
				Bookings: outBookings,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&BookingResp{
		Err: "no auth header",
	})
}

func getBookingsByHotel(w http.ResponseWriter, r *http.Request) {
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

			id := vars["id"]

			ctx := context.Background()
			txn := db.NewTxn()

			variables := map[string]string{"$id": id, "$user": claims.User.ID}
			q := `query q($id: uid, $user: uid){
                    var (func: uid($user)) {
		              u as uid
	                }
                    var (func: uid($id)) {
		              h as uid
	                }
                    bookings(func: has(booking)) @cascade {
                      uid
                      booking.start
                      booking.end
                      booking.hotel @filter(uid(h)) {
                        uid
                      }
                      booking.room {
                        uid
                      }
                      booking.user @filter(uid(u)) {
                        uid
                      }
	                }
                  }`

			resp, err := txn.QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}
			var bookings bookingQuery
			json.Unmarshal(resp.GetJson(), &bookings)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&BookingsResp{
					Err: err.Error(),
				})
				return
			}

			outBookings := make([]*Booking, 0)
			for _, booking := range bookings.Bookings {
				outBooking := &Booking{
					ID: booking.ID,
					HotelID: booking.Hotel[0].ID,
					RoomID: booking.Room[0].ID,
					Start: *booking.Start,
					End: *booking.End,
					UserID: booking.User[0].ID,
				}
				outBookings = append(outBookings, outBooking)
			}

			json.NewEncoder(w).Encode(&BookingsResp{
				Bookings: outBookings,
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
	r.Methods("GET").Path("/bookings/{id}").HandlerFunc(getBooking)
	r.Methods("GET").Path("/bookings/by-room/{id}").HandlerFunc(getBookingsByRoom)
	r.Methods("GET").Path("/bookings/by-hotel/{id}").HandlerFunc(getBookingsByHotel)

	return r
}

func newDbClient(dbHost string) *dgo.Dgraph {
	d, err := grpc.Dial(dbHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error connecting: %v\n", err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}

func setupSchema(c *dgo.Dgraph) {
	err := c.Alter(context.Background(), &api.Operation{
		Schema: `
			booking.start: dateTime .
			booking.end: dateTime .
			booking.hotel: uid @reverse .
			booking.room: uid @reverse .
			booking.user: uid @reverse .
		`,
	})
	if err != nil {
		log.Fatalf("Error setting up schema: %v\n", err)
	}
}

func main() {
	viper.SetDefault("DB_HOST", "dgraph-server-public:9080")

	viper.SetEnvPrefix("TRAVELR")
	viper.AutomaticEnv()

	dbHost := viper.GetString("DB_HOST")

	jwtSecret = []byte(viper.GetString("JWT_SECRET"))

	db = newDbClient(dbHost)

	setupSchema(db)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
