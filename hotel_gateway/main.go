package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"crypto/rand"
	"fmt"

	"github.com/fluidmediaproductions/central_hotel_door_server/hotel_comms"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const addr = ":80"
const RoomsServer = "http://rooms"

var db *gorm.DB

type Status struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

var status = &Status{}

type HotelServer struct {
	gorm.Model
	UUID      string `gorm:"unique"`
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
}

func protoServ(w http.ResponseWriter, r *http.Request) {
	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("body error:", readErr)
		return
	}
	defer r.Body.Close()

	newMsg := &hotel_comms.ProtoMsg{}
	err := proto.Unmarshal(body, newMsg)
	if err != nil {
		log.Println("Proto error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, handler := range protoHandlers {
		if handler.msgType == newMsg.GetType() {
			hotelServer := &HotelServer{
				UUID: *newMsg.UUID,
			}
			err := db.First(hotelServer).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Printf("Hotel %s not found\n", hotelServer.UUID)
					w.WriteHeader(http.StatusNotFound)
					return
				} else {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			err = verifySignature(newMsg.Msg, newMsg.Sig, hotelServer.PublicKey)
			if err != nil {
				log.Printf("Unable to verify signature from %s: %v\n", hotelServer.UUID, err)
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			err = handler.handler(hotelServer, newMsg.GetMsg(), newMsg.GetSig(), w)
			if err != nil {
				log.Printf("Error on handler for %s: %v\n", hotel_comms.MsgType_name[int32(newMsg.GetType())], err)
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func verifySignature(msg []byte, sig []byte, pubKey []byte) error {
	pub, err := x509.ParsePKIXPublicKey(pubKey)
	if err != nil {
		return err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		hash := crypto.SHA256
		h := hash.New()
		h.Write(msg)
		hashed := h.Sum(nil)
		err := rsa.VerifyPKCS1v15(pub, hash, hashed, sig)

		if err != nil {
			return err
		}
	default:
		return errors.New("invalid public key type")
	}
	return nil
}

func sendMsg(msg proto.Message, msgType hotel_comms.MsgType, w http.ResponseWriter) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	reader := rand.Reader
	hash := crypto.SHA256
	h := hash.New()
	h.Write(msgBytes)
	hashed := h.Sum(nil)
	sig, err := rsa.SignPKCS1v15(reader, status.PrivateKey, hash, hashed)
	if err != nil {
		return err
	}

	wrappedMsg := &hotel_comms.ProtoMsg{
		Type: &msgType,
		Msg:  msgBytes,
		Sig:  sig,
		UUID: proto.String(""),
	}

	wrappedMsgBytes, err := proto.Marshal(wrappedMsg)
	if err != nil {
		return err
	}

	w.Write(wrappedMsgBytes)
	return nil
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

func getRoomsByHotel(hotel uint) ([]interface{}, error) {
	req, err := http.NewRequest("GET", RoomsServer+fmt.Sprintf("/rooms/by-hotel/%d", hotel), nil)
	if err != nil {
		return nil, err
	}

	resp, err := utils.GetJson(req)
	if err != nil {
		return nil, err
	}
	respErr, isOk := resp["err"].(string)
	if isOk {
		if respErr != "" {
			return nil, errors.New(respErr)
		}
	}

	rooms, isOk := resp["rooms"].([]interface{})
	if isOk {
		return rooms, nil
	}
	return nil, nil
}

func getActions(hotelId uint) ([]*hotel_comms.Action, error) {
	actions := make([]*hotel_comms.Action, 0)
	rooms, err := getRoomsByHotel(hotelId)
	if err != nil {
		return nil, err
	}
	for _, room := range rooms {
		room, isOk := room.(map[string]interface{})
		if isOk {
			shouldOpen, isOk := room["shouldOpen"].(bool)
			if isOk {
				if shouldOpen {
					id, isOk := room["ID"].(float64)
					if isOk {
						actionType := hotel_comms.ActionType_ROOM_UNLOCK
						action := &hotel_comms.Action{
							Type: &actionType,
							Id:   proto.Int64(int64(id)),
						}
						actions = append(actions, action)
					}
				}
			}
		}
	}
	return actions, nil
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

func getAction(hotel *HotelServer, msg []byte, sig []byte, w http.ResponseWriter) error {
	newMsg := &hotel_comms.GetActions{}
	err := proto.Unmarshal(msg, newMsg)
	if err != nil {
		return err
	}

	actions, err := getActions(hotel.HotelId)
	if err != nil {
		return err
	}

	log.Println(actions)

	resp := &hotel_comms.GetActionsResp{
		Actions: actions,
	}

	return sendMsg(resp, hotel_comms.MsgType_GET_ACTIONS_RESP, w)
}

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "test.db")
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
