package main

import (
	"net/http"
	"io/ioutil"
	"log"
	"github.com/fluidmediaproductions/central_hotel_door_server/hotel_comms"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"crypto/x509"
	"crypto/rsa"
	"crypto"
	"crypto/rand"
	"errors"
)

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
