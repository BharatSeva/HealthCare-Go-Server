package redis

import (
	"fmt"
)

func (r *Redisconn) IsAllowed(healthcare_id string) (bool, error) {
	key := fmt.Sprintf("hip:rate_limit:%s", healthcare_id)
	count, err := r.conn.Incr(r.ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set the expiration time if this is the first request
	if count == 1 {
		err = r.conn.Expire(r.ctx, key, r.window).Err()
		if err != nil {
			return false, err
		}
	}

	// Check if the count exceeds the limit
	if count > int64(r.limit) {
		// block the request as limit has been reached
		return false, nil
	}

	return true, nil
}
