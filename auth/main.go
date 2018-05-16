package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"context"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/fluidmediaproductions/central_hotel_door_server/utils"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"strings"
	"fmt"
	"crypto/sha1"
	"errors"
)

const addr = ":80"

var db *dgo.Dgraph
var jwtSecret []byte

type JWTResp struct {
	Err string `json:"err"`
	Jwt string `json:"jwt"`
}

type ChangePasswordResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
}

type UpdateUserResp struct {
	Err     string `json:"err"`
	Success bool   `json:"success"`
}

type UserInfoResp struct {
	Err  string      `json:"err"`
	User *utils.User `json:"user"`
}

func checkForPwnage(pass string) error {
	h := sha1.New()
	h.Write([]byte(pass))
	hb := h.Sum(nil)
	hs := fmt.Sprintf("%x", hb)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", hs[:5]), nil)
	if err != nil {
		return err
	}

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respS := string(respBytes)

	hashes := strings.Split(respS, "\n")
	for _, hash := range hashes {
		parts := strings.Split(hash, ":")
		if hs[:5] + parts[0] == hs {
			return errors.New("pwned password")
		}
	}
	return nil
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		return
	}
	defer r.Body.Close()

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&JWTResp{
			Err: err.Error(),
		})
		return
	}

	email, isOk := data["email"].(string)
	if isOk {
		pass, isOk := data["pass"].(string)
		if isOk {
			ctx := context.Background()

			variables := map[string]string{"$email": email, "$pass": pass}
			q := `query Me($email: string, $pass: string){
                    login_attempt(func: has(user)) @filter(eq(email, $email)) {
                      uid
                      email
                      name
                      checkpwd(pass, $pass)
	                }
                  }`

			resp, err := db.NewTxn().QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			var login struct {
				Account []struct {
					Pass []struct {
						CheckPwd bool `json:"checkpwd"`
					} `json:"pass"`
					Email string `json:"email"`
					Name  string `json:"name"`
					ID    string `json:"uid"`
				} `json:"login_attempt"`
			}
			err = json.Unmarshal(resp.GetJson(), &login)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			if len(login.Account) == 0 {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: "user not found",
				})
				return
			}

			if login.Account[0].Pass[0].CheckPwd {
				user := &utils.User{
					Email: login.Account[0].Email,
					Name:  login.Account[0].Name,
					ID:    login.Account[0].ID,
				}

				jwt, err := utils.NewJWT(user, jwtSecret)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&JWTResp{
						Err: err.Error(),
					})
					return
				}
				json.NewEncoder(w).Encode(&JWTResp{
					Jwt: jwt,
				})
				return
			} else {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: "invalid password",
				})
				return
			}
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&JWTResp{
		Err: "bad request data",
	})
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		return
	}
	defer r.Body.Close()

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&ChangePasswordResp{
			Err: err.Error(),
		})
		return
	}

	pass, isOk := data["pass"].(string)
	if isOk {

		authHeaders, isOk := r.Header["Authorization"]
		if isOk {
			if len(authHeaders) > 0 {
				authHeader := authHeaders[0]
				jwt := strings.TrimPrefix(authHeader, "Bearer ")

				claims, err := utils.VerifyJWT(jwt, jwtSecret)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				ctx := context.Background()
				txn := db.NewTxn()

				variables := map[string]string{"$id": claims.User.Email}
				q := `query Me($id: uid){
                        user(func: uid($id)) @filter(has(user)) {
                          uid
                        }
                      }`

				resp, err := txn.QueryWithVars(ctx, q, variables)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				var user struct {
					Account []struct {
						ID string `json:"uid"`
					} `json:"user"`
				}
				err = json.Unmarshal(resp.GetJson(), &user)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				if len(user.Account) == 0 {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: "user not found",
					})
					return
				}

				err = checkForPwnage(pass)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				var mutation struct {
					ID   string `json:"uid"`
					Pass string `json:"pass"`
				}
				mutation.ID = claims.User.ID
				mutation.Pass = pass

				mutData, err := json.Marshal(&mutation)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				mu := &api.Mutation{
					SetJson: mutData,
				}

				_, err = txn.Mutate(ctx, mu)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(&ChangePasswordResp{
						Err: err.Error(),
					})
					return
				}

				json.NewEncoder(w).Encode(&ChangePasswordResp{
					Success: true,
				})
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(&ChangePasswordResp{
			Err: "no auth header",
		})
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&ChangePasswordResp{
		Err: "bad request data",
	})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		return
	}
	defer r.Body.Close()

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&JWTResp{
			Err: err.Error(),
		})
		return
	}
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&UpdateUserResp{
					Err: err.Error(),
				})
				return
			}

			ctx := context.Background()
			txn := db.NewTxn()

			variables := map[string]string{"$id": claims.User.Email}
			q := `query Me($id: uid){
                        user(func: uid($id)) @filter(has(user)) {
                          uid
                        }
                      }`

			resp, err := txn.QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			var user struct {
				Account []struct {
					ID string `json:"uid"`
				} `json:"user"`
			}
			err = json.Unmarshal(resp.GetJson(), &user)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			if len(user.Account) == 0 {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: "user not found",
				})
				return
			}

			var mutation struct {
				ID    string `json:"uid"`
				Name  string `json:"name,omitempty"`
				Email string `json:"email,omitempty"`
			}
			mutation.ID = user.Account[0].ID

			email, isOk := data["email"].(string)
			if isOk {
				mutation.Email = email
			}
			name, isOk := data["name"].(string)
			if isOk {
				mutation.Name = name
			}

			mutData, err := json.Marshal(&mutation)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			mu := &api.Mutation{
				SetJson: mutData,
			}

			_, err = txn.Mutate(ctx, mu)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			json.NewEncoder(w).Encode(&UpdateUserResp{
				Success: true,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&UpdateUserResp{
		Err: "no auth header",
	})
}

func userInfo(w http.ResponseWriter, r *http.Request) {
	authHeaders, isOk := r.Header["Authorization"]
	if isOk {
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			jwt := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(jwt, jwtSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(&UserInfoResp{
					Err: err.Error(),
				})
				return
			}

			ctx := context.Background()
			txn := db.NewTxn()

			variables := map[string]string{"$id": claims.User.Email}
			q := `query Me($id: uid){
                        user(func: uid($id)) @filter(has(user)) {
                          uid
                          email
                          name
                        }
                      }`

			resp, err := txn.QueryWithVars(ctx, q, variables)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			var user struct {
				Account []struct {
					ID    string `json:"uid"`
					Email string `json:"email"`
					Name  string `json:"name"`
				} `json:"user"`
			}
			err = json.Unmarshal(resp.GetJson(), &user)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: err.Error(),
				})
				return
			}

			if len(user.Account) == 0 {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&JWTResp{
					Err: "user not found",
				})
				return
			}

			userResp := &utils.User{
				ID:    user.Account[0].ID,
				Email: user.Account[0].Email,
				Name:  user.Account[0].Name,
			}

			json.NewEncoder(w).Encode(&UserInfoResp{
				User: userResp,
			})
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&UserInfoResp{
		Err: "no auth header",
	})
}

func router() *mux.Router {
	r := mux.NewRouter()

	r.Methods("POST").Path("/login").HandlerFunc(loginUser)
	r.Methods("POST").Path("/changePassword").HandlerFunc(changePassword)
	r.Methods("POST").Path("/updateUser").HandlerFunc(updateUser)
	r.Methods("GET").Path("/userInfo").HandlerFunc(userInfo)

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
			name: string .
			email: string @index(hash) @upsert .
            pass: password .
		`,
	})
	if err != nil {
		log.Fatalf("Error setting up schema: %v\n", err)
	}
}

func main() {
	viper.SetDefault("DB_HOST", "draph-server-public:9080")

	viper.SetEnvPrefix("TRAVELR")
	viper.AutomaticEnv()

	dbHost := viper.GetString("DB_HOST")

	jwtSecret = []byte(viper.GetString("JWT_SECRET"))

	db = newDbClient(dbHost)

	setup(db)

	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, router()))
}
