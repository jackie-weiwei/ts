package ts

import (
	"encoding/binary"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

var redisConn *redis.Pool

// Setup Initialize the Redis instance
func Setup() error {
	redisConn = &redis.Pool{
		MaxIdle:     30,
		MaxActive:   30,
		IdleTimeout: 200,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			// if setting.RedisSetting.Password != "" {
			// 	if _, err := c.Do("AUTH", setting.RedisSetting.Password); err != nil {
			// 		c.Close()
			// 		return nil, err
			// 	}
			// }
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return nil
}

// Set a key/value
func Set(key string, data interface{}, time int) error {
	conn := redisConn.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, value)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", key, time)
	if err != nil {
		return err
	}

	return nil
}

func SetKey(key string, value interface{}) error {
	conn := redisConn.Get()
	defer conn.Close()

	value, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, value)
	if err != nil {
		return err
	}
	return nil
}

func Expire(key string, time int) error {
	conn := redisConn.Get()
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, time)
	if err != nil {
		return err
	}

	return nil
}

// Exists check a key
func Exists(key string) bool {
	conn := redisConn.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false
	}

	return exists
}

// Get get a key
func Get(key string) ([]byte, error) {
	conn := redisConn.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.LittleEndian.Uint64(buf))
}

func BytesToInt(bytesArr []byte) int {
	var intNum int
	if len(bytesArr) == 1 {
		bytesArr = append(bytesArr, byte(0))
		intNum = int(binary.LittleEndian.Uint16(bytesArr))
	} else if len(bytesArr) == 2 {
		intNum = int(binary.LittleEndian.Uint16(bytesArr))
	}

	return intNum
}

func GetInt(key string) (int64, error) {
	v, error := Get(key)

	if error == nil {
		ival, err1 := strconv.ParseInt(string(v), 10, 64)
		if err1 == nil {
			return ival, nil
		} else {
			return 0, err1
		}

	}

	return 0, error
}

func GetString(key string) (string, error) {
	v, err := Get(key)

	if err == nil {
		var sval string
		json.Unmarshal(v, &sval)
		return sval, err
	}

	return "", err
}

// Delete delete a kye
func Delete(key string) (bool, error) {
	conn := redisConn.Get()
	defer conn.Close()

	return redis.Bool(conn.Do("DEL", key))
}

// LikeDeletes batch delete
func LikeDeletes(key string) error {
	conn := redisConn.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", "*"+key+"*"))
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err = Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}
