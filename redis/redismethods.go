package redis

import (
	"encoding/json"
	"fmt"
	"time"
)

func (r *Redisconn) Set(key string, value interface{}) error {
	// First Serialize to json format
	reqData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize data1")
	}

	//set to redis server
	err = r.conn.Set(r.ctx, key, reqData, time.Hour).Err()
	return err
}

func (r *Redisconn) Get(key string) (interface{}, error) {
	val, err := r.conn.Get(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	// ttl, err := r.conn.TTL(r.ctx, key).Result()
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(ttl)
	// marshal the data first adn everytime I call it
	// var jsonBody ChangePreference
	// err = json.Unmarshal([]byte(val), &jsonBody)
	// if err != nil {
	// 	return nil, err
	// }

	return val, nil
}

func (r *Redisconn) Close() error {
	return r.conn.Close()
}
