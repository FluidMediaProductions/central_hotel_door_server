package main

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/dgraph-io/dgo"
	"google.golang.org/grpc"
	"github.com/dgraph-io/dgo/protos/api"
	"context"
)

const addr = ":80"

var db *dgo.Dgraph

type Room struct {
	ID string `json:"uid"`
	Name       string `json:"name"`
	Floor      string `json:"floor"`
	HotelID    string   `json:"hotelId"`
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

type roomQuery struct {
	Rooms []struct {
		ID    string `json:"uid"`
		Name    string `json:"room.name"`
		Floor    string `json:"room.floor"`
		Hotel  []struct{
			ID    string `json:"uid"`
		} `json:"room.hotel"`
	} `json:"rooms"`
}

func getRooms(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	txn := db.NewTxn()

	q := `query {
            rooms(func: has(room)) @cascade {
              uid
              room.name
              room.floor
              room.hotel {
                uid
              }
	        }
          }`

	resp, err := txn.Query(ctx, q)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomsResp{
			Err: err.Error(),
		})
		return
	}
	var rooms roomQuery
	err = json.Unmarshal(resp.GetJson(), &rooms)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomsResp{
			Err: err.Error(),
		})
		return
	}

	outRooms:= make([]*Room, 0)
	for _, room := range rooms.Rooms {
		outRoom := &Room{
			ID:         room.ID,
			Name:       room.Name,
			Floor: room.Floor,
			HotelID: room.Hotel[0].ID,
		}
		outRooms = append(outRooms, outRoom)
	}

	json.NewEncoder(w).Encode(&RoomsResp{
		Rooms: outRooms,
	})
}

func getRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]

	ctx := context.Background()
	txn := db.NewTxn()

	q := `query q($id: string) {
            rooms(func: uid($id)) @filter(has(room)) @cascade {
              uid
              room.name
              room.floor
              room.hotel {
                uid
              }
	        }
          }`

	resp, err := txn.QueryWithVars(ctx, q, map[string]string{"$id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomResp{
			Err: err.Error(),
		})
		return
	}
	var rooms roomQuery
	err = json.Unmarshal(resp.GetJson(), &rooms)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&RoomResp{
			Err: err.Error(),
		})
		return
	}

	if len(rooms.Rooms) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&RoomResp{
			Err: "room not found",
		})
		return
	}

	room := rooms.Rooms[0]
	outRoom := &Room{
		ID:         room.ID,
		Name:       room.Name,
		Floor: room.Floor,
		HotelID: room.Hotel[0].ID,
	}

	json.NewEncoder(w).Encode(&RoomResp{
		Room: outRoom,
	})
}

func getRoomsByHotel(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//
	//id := vars["id"]
	//
	//rooms := make([]*Room, 0)
	//err := db.Find(&rooms, &Room{
	//	HotelID: id,
	//}).Error
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	json.NewEncoder(w).Encode(&RoomsResp{
	//		Err: err.Error(),
	//	})
	//	return
	//}
	//
	//json.NewEncoder(w).Encode(&RoomsResp{
	//	Rooms: rooms,
	//})
}

func openRoom(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//
	//id, _ := strconv.Atoi(vars["id"])
	//
	//room := &Room{}
	//err := db.Find(&room, id).Error
	//if err != nil {
	//	if err == gorm.ErrRecordNotFound {
	//		w.WriteHeader(http.StatusNotFound)
	//		json.NewEncoder(w).Encode(&OpenRoomResp{
	//			Err: "room not found",
	//		})
	//		return
	//	} else {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		json.NewEncoder(w).Encode(&OpenRoomResp{
	//			Err: err.Error(),
	//		})
	//		return
	//	}
	//}
	//
	//room.ShouldOpen = true
	//err = db.Save(&room).Error
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	json.NewEncoder(w).Encode(&OpenRoomResp{
	//		Err: err.Error(),
	//	})
	//	return
	//}
	//
	//json.NewEncoder(w).Encode(&OpenRoomResp{
	//	Success: true,
	//})
}

func openRoomSuccess(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//
	//id, _ := strconv.Atoi(vars["id"])
	//
	//room := &Room{}
	//err := db.Find(&room, id).Error
	//if err != nil {
	//	if err == gorm.ErrRecordNotFound {
	//		w.WriteHeader(http.StatusNotFound)
	//		json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
	//			Err: "room not found",
	//		})
	//		return
	//	} else {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
	//			Err: err.Error(),
	//		})
	//		return
	//	}
	//}
	//
	//room.ShouldOpen = false
	//err = db.Save(&room).Error
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
	//		Err: err.Error(),
	//	})
	//	return
	//}
	//
	//json.NewEncoder(w).Encode(&OpenRoomSuccessResp{
	//	Success: true,
	//})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/rooms").HandlerFunc(getRooms)
	r.Methods("GET").Path("/rooms/{id}").HandlerFunc(getRoom)
	r.Methods("GET").Path("/rooms/by-hotel/{id}").HandlerFunc(getRoomsByHotel)
	r.Methods("GET").Path("/rooms/{id}/open").HandlerFunc(openRoom)
	r.Methods("GET").Path("/rooms/{id}/open-success").HandlerFunc(openRoomSuccess)

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
			room.name: string .
			room.floor: string .
			room.hotel: uid .
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

	db = newDbClient(dbHost)

	setup(db)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
