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
	"github.com/jinzhu/gorm"
	_"github.com/jinzhu/gorm/dialects/mysql"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

const addr = ":80"

var db *gorm.DB

type Status struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

var status = &Status{}

type HotelServer struct {
	gorm.Model
	UUID      string
	HotelId   uint
	LastSeen  time.Time
	Online    bool
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
}

func checkHotels() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		hotelServers := make([]*HotelServer, 0)
		db.Find(&hotelServers, &HotelServer{Online: true})

		for _, hotelServer := range hotelServers {
			if time.Since(hotelServer.LastSeen) > time.Minute {
				log.Printf("Removing hotel %v, too old\n", hotelServer.UUID)
				hotelServer.Online = false
				db.Save(&hotelServer)
			}
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

func main() {
	viper.SetDefault("DB_HOST", "mysql")
	viper.SetDefault("DB_USER", "travelr")
	viper.SetDefault("DB_NAME", "hotel_gateway")

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

	db.AutoMigrate(&HotelServer{})

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
