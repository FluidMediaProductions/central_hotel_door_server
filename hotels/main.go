package main

import (
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	//"strconv"
	//"strings"
	"time"

	"context"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const addr = ":80"

var BookingsServer = "http://bookings"

var db *dgo.Dgraph
var jwtSecret []byte

type Hotel struct {
	ID             string          `json:"uid"`
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

type hotelQuery struct {
	Hotels []struct {
		ID    string `json:"uid"`
		Name    string `json:"hotel.name"`
		Address    string `json:"hotel.address"`
		CheckIn   *time.Time `json:"hotel.checkIn"`
		HasCarPark  bool `json:"hotel.hasCarPark"`
	} `json:"hotels"`
}

func getHotels(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	txn := db.NewTxn()

	q := `query{
            hotels(func: has(hotel)) @cascade {
              uid
              hotel.name
              hotel.address
              hotel.location
              hotel.checkIn
              hotel.hasCarPark
	        }
          }`

	resp, err := txn.Query(ctx, q)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&HotelsResp{
			Err: err.Error(),
		})
		return
	}
	var hotels hotelQuery
	err = json.Unmarshal(resp.GetJson(), &hotels)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&HotelsResp{
			Err: err.Error(),
		})
		return
	}

	outHotels := make([]*Hotel, 0)
	for _, hotel := range hotels.Hotels {
		outHotel := &Hotel{
			ID: hotel.ID,
			Name: hotel.Name,
			Address: hotel.Address,
			CheckIn: *hotel.CheckIn,
			HasCarPark: hotel.HasCarPark,
		}
		outHotels = append(outHotels, outHotel)
	}

	json.NewEncoder(w).Encode(&HotelsResp{
		Hotels: outHotels,
	})
	return
}

func getHotel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]

	ctx := context.Background()
	txn := db.NewTxn()

	q := `query q($id: string) {
            hotels(func: uid($id)) @filter(has(hotel)) @cascade {
              uid
              hotel.name
              hotel.address
              hotel.location
              hotel.checkIn
              hotel.hasCarPark
	        }
          }`

	resp, err := txn.QueryWithVars(ctx, q, map[string]string{"$id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&HotelsResp{
			Err: err.Error(),
		})
		return
	}
	var hotels hotelQuery
	err = json.Unmarshal(resp.GetJson(), &hotels)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&HotelsResp{
			Err: err.Error(),
		})
		return
	}

	if len(hotels.Hotels) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&HotelsResp{
			Err: "hotel not found",
		})
		return
	}

	hotel := hotels.Hotels[0]
	outHotel := &Hotel{
			ID: hotel.ID,
			Name: hotel.Name,
			Address: hotel.Address,
			CheckIn: *hotel.CheckIn,
			HasCarPark: hotel.HasCarPark,
		}

	json.NewEncoder(w).Encode(&HotelResp{
		Hotel: outHotel,
	})
}

func openHotel(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	//id, _ := strconv.Atoi(vars["id"])
	//
	//hotel := &Hotel{}
	//err := db.Find(&hotel, id).Error
	//if err != nil {
	//	if err == gorm.ErrRecordNotFound {
	//		w.WriteHeader(http.StatusNotFound)
	//		json.NewEncoder(w).Encode(&OpenHotelResp{
	//			Err: "hotel not found",
	//		})
	//		return
	//	} else {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		json.NewEncoder(w).Encode(&OpenHotelResp{
	//			Err: err.Error(),
	//		})
	//		return
	//	}
	//}
	//
	//authHeaders, isOk := r.Header["Authorization"]
	//if isOk {
	//	if len(authHeaders) > 0 {
	//		authHeader := authHeaders[0]
	//		jwt := strings.TrimPrefix(authHeader, "Bearer ")
	//
	//		_, err := utils.VerifyJWT(jwt, jwtSecret)
	//		if err != nil {
	//			w.WriteHeader(http.StatusForbidden)
	//			json.NewEncoder(w).Encode(&OpenHotelResp{
	//				Err: err.Error(),
	//			})
	//			return
	//		}
	//
	//		req, err := http.NewRequest("GET", BookingsServer+fmt.Sprintf("/bookings/by-hotel/%d", id), nil)
	//		if err != nil {
	//			w.WriteHeader(http.StatusInternalServerError)
	//			json.NewEncoder(w).Encode(&OpenHotelResp{
	//				Err: err.Error(),
	//			})
	//			return
	//		}
	//
	//		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))
	//
	//		resp, err := utils.GetJson(req)
	//		if err != nil {
	//			w.WriteHeader(http.StatusInternalServerError)
	//			json.NewEncoder(w).Encode(&OpenHotelResp{
	//				Err: err.Error(),
	//			})
	//			return
	//		}
	//		respErr, isOk := resp["err"].(string)
	//		if isOk {
	//			if respErr != "" {
	//				json.NewEncoder(w).Encode(&OpenHotelResp{
	//					Err: respErr,
	//				})
	//				return
	//			}
	//		}
	//
	//		_, isOk = resp["booking"].(map[string]interface{})
	//		if isOk {
	//			hotel.ShouldDoorOpen = true
	//			err := db.Save(&hotel).Error
	//			if err != nil {
	//				w.WriteHeader(http.StatusInternalServerError)
	//				json.NewEncoder(w).Encode(&OpenHotelResp{
	//					Err: err.Error(),
	//				})
	//				return
	//			}
	//
	//			w.WriteHeader(http.StatusInternalServerError)
	//			json.NewEncoder(w).Encode(&OpenHotelResp{
	//				Success: true,
	//			})
	//			return
	//		}
	//		w.WriteHeader(http.StatusInternalServerError)
	//		json.NewEncoder(w).Encode(&OpenHotelResp{
	//			Err: "unknown",
	//		})
	//		return
	//	}
	//}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&OpenHotelResp{
		Err: "no auth header",
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/hotels").HandlerFunc(getHotels)
	r.Methods("GET").Path("/hotels/{id}").HandlerFunc(getHotel)
	r.Methods("GET").Path("/hotels/{id}/open").HandlerFunc(openHotel)

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

func setup(c *dgo.Dgraph) {
	err := c.Alter(context.Background(), &api.Operation{
		Schema: `
			hotel.name: string .
			hotel.address: string .
			hotel.location: geo .
			hotel.checkIn: dateTime .
			hotel.hasCarPark: bool .
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

	setup(db)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
