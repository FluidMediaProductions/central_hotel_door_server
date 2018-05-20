package main

import (
	"log"
	"github.com/spf13/viper"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"strings"
)

const addr = ":80"

var jwtSecret []byte

func getJWT(r *http.Request) (*utils.MQTTJWTClaims, bool) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyMQTTJWT(jwt, jwtSecret)
			if err != nil {
				log.Printf("Auth fail for %s", jwt)
				return nil, false
			}
			log.Printf("Auth success for %s", jwt)
			return claims, true
		}
	}
	return nil, false
}

func auth(w http.ResponseWriter, r *http.Request)  {
	claims, success := getJWT(r)
	if !success {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if claims.User.IsHotel || claims.User.IsServer || claims.User.IsSuperUser {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

func superuser(w http.ResponseWriter, r *http.Request)  {
	claims, success := getJWT(r)
	if !success {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if claims.User.IsSuperUser {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

func checkTopic(pattern string, topic string, uuid string) bool {
	pattern = strings.TrimPrefix(pattern, "/")
	topic = strings.TrimPrefix(topic, "/")
	patternSections := strings.Split(pattern, "/")
	topicSections := strings.Split(topic, "/")

	i := 0
	for i < len(topicSections) {
		if i == len(patternSections) {
			if i-1 >= 0 {
				if patternSections[i-1] == "#" {
					return true
				}
			}
			return false
		} else if i > len(patternSections) {
			return false
		}
		topicSection := topicSections[i]
		patternSection := patternSections[i]

		if patternSection == "%u" {
			if topicSection != uuid {
				return false
			}
		} else if patternSection == "+" {
			i++
			continue
		} else if patternSection == "#" {
			if topicSection == "+" {
				i++
				continue
			}
			if i+1 < len(patternSections) {
				nextSection := patternSections[i+1]
				for topicSections[i] != nextSection {
					i++
					if i == len(topicSections) {
						return false
					}
				}
			} else {
				return true
			}
		} else if patternSection != topicSection {
			return false
		}
		i++
	}
	return true
}

func acl(w http.ResponseWriter, r *http.Request)  {
	claims, success := getJWT(r)
	if !success {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	r.ParseForm()
	topic    := r.Form.Get("topic")
	acc := r.Form.Get("acc") // 1 == READ, 2 == WRITE, 4 == SUB
	if acc == "4" {
		acc = "1"
	}

	if claims.User.IsHotel {
		if checkTopic("hotels/%u/room/open", topic, claims.User.UUID) {
			if acc == "1" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		if checkTopic("hotels/%u/ping", topic, claims.User.UUID) {
			if acc == "2" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	} else if claims.User.IsServer {
		if checkTopic("hotels/+/room/open", topic, claims.User.UUID) {
			if acc == "2" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		if checkTopic("hotels/+/ping", topic, claims.User.UUID) {
			if acc == "1" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	w.WriteHeader(http.StatusForbidden)
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("POST").Path("/auth").HandlerFunc(auth)
	r.Methods("POST").Path("/superuser").HandlerFunc(superuser)
	r.Methods("POST").Path("/acl").HandlerFunc(acl)

	return r
}

func main() {
	viper.SetEnvPrefix("TRAVELR")
	viper.AutomaticEnv()

	jwtSecret = []byte(viper.GetString("JWT_SECRET"))

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}