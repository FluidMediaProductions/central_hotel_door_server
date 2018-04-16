package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func GetJson(r *http.Request, data *interface{}) error {
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		log.Println(err)
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(respBytes, data)
	if err != nil {
		log.Println(err)
		return  err
	}
	return nil
}
