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

	ttl, err := r.conn.TTL(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// return as single interface
	response := struct {
		Value string        `json:"value"`
		TTL   time.Duration `json:"ttl"`
	}{
		Value: val,
		TTL:   ttl,
	}

	return response, nil
}

func (r *Redisconn) Close() error {
	return r.conn.Close()
}
