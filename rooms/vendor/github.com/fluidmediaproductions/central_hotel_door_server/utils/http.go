package utils

import (
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
)

func GetJson(r *http.Request) (map[string]interface{}, error) {
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	data := map[string]interface{}{}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, nil
}
