package redis

import (
	"config"
	"log"

	toml "github.com/extrame/go-toml-config"
	"github.com/garyburd/redigo/redis"
)

//TODO

type server struct {
	plugins map[string]Plugin
}

type Plugin interface {
	ParseConfig(prefix string) error
	Init() error
}

type Redis struct {
	address  *string
	password *string
	db       *int64
}

var redisPool *redis.Pool
var PoolMaxIdle = 10

// 从配置文件获取 redis配置信息   config: redis
func Init() {
	var addr, pwd, prefix string

	if config.TConfig.RedisAddr == "" {
		addr = "localhost"
	} else {
		addr = config.TConfig.RedisAddr
	}

	if config.TConfig.RedisPwd == "" {
		pwd = ""
	} else {
		pwd = config.TConfig.RedisPwd
	}

	if config.TConfig.RedisPrefix == "" {
		prefix = "default"
	} else {
		prefix = config.TConfig.RedisPrefix
	}
	log.Println("addr:", addr)
	r := Redis{address: &addr, password: &pwd}
	r.ParseConfig(prefix)
	r.Init()
}

// 从文件解析 redis配置信息
func (r *Redis) ParseConfig(prefix string) error {
	log.Println("parseConfig---------------==========================================")
	r.address = toml.String(prefix+".address", "localhost:6379")
	r.password = toml.String(prefix+".password", "")
	r.db = toml.Int64(prefix+".db", 0)
	return nil
}

func (r *Redis) Init() error {
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *r.address)

		log.Println("parseConfig---------------==========================================", *r.address)
		if err != nil {
			log.Println("--Redis--Connect redis fail:" + err.Error())
			return nil, err
		}
		if len(*r.password) > 0 {
			if _, err := c.Do("AUTH", *r.password); err != nil {
				c.Close()
				log.Println("--Redis--Auth redis fail:" + err.Error())
				return nil, err
			}
		}
		if _, err := c.Do("SELECT", *r.db); err != nil {
			c.Close()
			log.Println("--Redis--Select redis db fail:" + err.Error())
			return nil, err
		}
		return c, nil
	}, PoolMaxIdle)
	return nil
}

func Store(key, itemKey string, item interface{}) error {
	c := redisPool.Get()
	defer c.Close()
	if _, err := c.Do("HSET", key, itemKey, item); err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func Expire(key string, seconds int) error {
	c := redisPool.Get()
	defer c.Close()
	log.Println("-------Expire---")
	if _, err := c.Do("EXPIRE", key, seconds); err != nil {
		log.Println("-------Expire---", err)
		log.Println(err.Error())
		return err
	}
	return nil
}

func Exists(key, itemKey string) bool {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", key, itemKey))
	if count == 0 {
		return false
	}
	return true
}

func Get(userKey, itemKey string) (interface{}, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		res, _ := redis.Values(c.Do("HGET", userKey, itemKey))

		return res, true
	}
}

func GetBool(userKey, itemKey string) (bool, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return false, false
	} else {
		n, _ := redis.Bool(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetBytes(userKey, itemKey string) ([]byte, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Bytes(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetFloat64(userKey, itemKey string) (float64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Float64(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInt(userKey, itemKey string) (int, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Int(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInt64(userKey, itemKey string) (int64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Int64(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetIntMap(userKey, itemKey string) (map[string]int, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.IntMap(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInt64Map(userKey, itemKey string) (map[string]int64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Int64Map(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInts(userKey, itemKey string) ([]int, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Ints(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetString(userKey, itemKey string) (string, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return "", false
	} else {
		n, _ := redis.String(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetStrings(userKey, itemKey string) ([]string, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Strings(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetStringMap(userKey, itemKey string) (map[string]string, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.StringMap(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetUint64(userKey, itemKey string) (uint64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Uint64(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func RemoveItem(userKey, itemKey string) bool {
	c := redisPool.Get()
	defer c.Close()
	if _, err := c.Do("HDEL", userKey, itemKey); err != nil {
		return false
	}
	return true
}