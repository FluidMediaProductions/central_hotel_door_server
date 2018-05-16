package main

import (
	"crypto/rsa"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/fluidmediaproductions/central_hotel_door_server/hotel_comms"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	_"github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"github.com/dgraph-io/dgo"
	"google.golang.org/grpc"
	"github.com/dgraph-io/dgo/protos/api"
	"context"
	"encoding/json"
	"encoding/base64"
)

const addr = ":80"

var db *dgo.Dgraph

type Status struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

var status = &Status{}

type HotelServer struct {
	ID      string
	UUID      string
	HotelId   string
	LastSeen  time.Time
	Online bool
	PublicKey []byte
}

type ProtoHandlerFunc func(hotel *HotelServer, msg []byte, sig []byte, w http.ResponseWriter) error

type ProtoHandler struct {
	msgType hotel_comms.MsgType
	handler ProtoHandlerFunc
}

var protoHandlers = []ProtoHandler{
	{
		msgType: hotel_comms.MsgType_HOTEL_PING,
		handler: hotelPing,
	},
	{
		msgType: hotel_comms.MsgType_GET_ACTIONS,
		handler: getAction,
	},
	{
		msgType: hotel_comms.MsgType_GET_DOORS,
		handler: getDoors,
	},
	{
		msgType: hotel_comms.MsgType_ACTION_COMPLETE,
		handler: actionComplete,
	},
}

type hotelQuery struct {
	Hotels []struct {
		ID    string `json:"uid"`
		UUID    string `json:"hotelServer.uuid"`
		LastSeen    *time.Time `json:"hotelServer.lastSeen"`
		Online    bool `json:"hotelServer.online"`
		PubKey   string `json:"hotelServer.pubKey"`
	} `json:"hotels"`
}

func getHotels() (*hotelQuery, error) {
	ctx := context.Background()
	txn := db.NewTxn()

	q := `query{
            hotels(func: has(hotelServer)) @cascade {
              uid
              hotel.uuid
              hotel.lastSeen
              hotel.pubKey
              hotel.checkIn
              hotel.hasCarPark
	        }
          }`

	resp, err := txn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	var hotels *hotelQuery
	err = json.Unmarshal(resp.GetJson(), &hotels)
	if err != nil {
		return nil, err
	}

	return hotels, nil
}

func checkHotels() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		hotelServers, err := getHotels()
		if err != nil {
			log.Println(err)
			continue
		}

		for i, hotelServer := range hotelServers.Hotels {
			if time.Since(*hotelServer.LastSeen) > time.Minute {
				log.Printf("Removing hotel %v, too old\n", hotelServer.UUID)
				hotelServers.Hotels[i].Online = false
			}
		}
		txn := db.NewTxn()
		out, err := json.Marshal(hotelServers.Hotels)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = txn.Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func hotelPing(hotel *HotelServer, msg []byte, sig []byte, w http.ResponseWriter) error {
	newMsg := &hotel_comms.HotelPing{}
	err := proto.Unmarshal(msg, newMsg)
	if err != nil {
		return err
	}

	if time.Since(time.Unix(*newMsg.Timestamp, 0)) > time.Hour {
		log.Printf("Hotle %v out of sync with server time\n", hotel.UUID)

		resp := &hotel_comms.HotelPingResp{
			Success: proto.Bool(false),
			Error:   proto.String("time out of sync"),
		}
		w.WriteHeader(http.StatusNotAcceptable)
		sendMsg(resp, hotel_comms.MsgType_HOTEL_PING_RESP, w)
		return errors.New("hotel out of sync")
	}

	hotel.LastSeen = time.Now()
	hotel.Online = true
	db.Save(hotel)

	actions, err := getActions(hotel.HotelId)
	if err != nil {
		return err
	}
	actionRequired := len(actions) != 0

	resp := &hotel_comms.HotelPingResp{
		Success:        proto.Bool(true),
		ActionRequired: proto.Bool(actionRequired),
	}

	w.WriteHeader(http.StatusOK)
	return sendMsg(resp, hotel_comms.MsgType_HOTEL_PING_RESP, w)
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
			hotelServer.uuid: string .
			hotelServer.hotel: uid @reverse .
			hotelServer.lastSeen: dateTime .
			hotelServer.online: bool .
			hotelServer.pubKey: string .
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

	setupSchema(db)

	priv, pub, err := hotel_comms.GetKeys()
	if err != nil {
		log.Fatalf("Can't get encryption keys: %v\n", err)
	}
	status.PublicKey = pub
	status.PrivateKey = priv

	go checkHotels()

	r := mux.NewRouter()
	r.Methods("POST").Path("/proto").HandlerFunc(protoServ)
	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, r))
}
