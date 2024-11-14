package redis

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

func (r *Redisconn) IsAllowed(healthcare_id string) (bool, error) {
	key := fmt.Sprintf("hip:rate_limit:%s", healthcare_id)
	count, err := r.conn.Incr(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	
	// count total request per user based on healthcare_id
	total_count_session := fmt.Sprintf("hip:total_count:%s", healthcare_id)
	total_count, err := r.conn.Incr(r.ctx, total_count_session).Result()
	if err != nil {
		return false, err
	}
	// if total request per session crossed 300 then logout
	// block the request forever untill quota has been reset again
	if total_count > 300 {
		// block request permanently
		return false, err
	}

	if count == 1 {
		err = r.conn.Expire(r.ctx, key, r.window).Err()
		if err != nil {
			return false, err
		}
	}


	if count > r.limit {
		// block the request as limit has been reached
		err = r.conn.Expire(r.ctx, key, 5*time.Minute).Err()
		if err != nil {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func (r *Redisconn) IsAllowed_leaky_bucket(healthcare_id string) (bool, error) {
	key := fmt.Sprintf("hip:leaky_bucket:%s", healthcare_id)

	now := time.Now()
	timePassed := now.Sub(r.lastchecked).Seconds()
	allowedOutflow := int64(timePassed * 50) // 50 req second
	r.lastchecked = now

	_, err := r.conn.DecrBy(r.ctx, key, allowedOutflow).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	count, err := r.conn.Get(r.ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return false, err
	}

	if count >= r.limit {
		return false, nil
	}

	_, err = r.conn.Incr(r.ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 0 {
		err = r.conn.Expire(r.ctx, key, time.Minute).Err()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
