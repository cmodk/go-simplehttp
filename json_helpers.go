package simplehttp

import (
	"encoding/json"
)

func (sh *SimpleHttp) GetJson(url string, dst interface{}) error {

	data, err := sh.Get(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dst); err != nil {
		return err
	}

	return nil

}

func (sh *SimpleHttp) PostJson(url string, dst interface{}, parseResponse bool) error {

	data, err := sh.Post(url, dst)
	if err != nil {
		return err
	}

	if parseResponse {
		if err := json.Unmarshal([]byte(data), dst); err != nil {
			return err
		}
	}

	return nil

}
