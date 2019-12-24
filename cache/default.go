package cache

import red "github.com/gomodule/redigo/redis"

var redisPool map[string]*red.Pool = make(map[string]*red.Pool)

func init() {
	// 注册redis default模块
	//RegisterRedisPool("default","127.0.0.1","6379","")

	// 注册redis sql模块
	//registerRedisPool("sql")
}

func RegisterRedisPool(module, host, port, auth string) {
	// 只注册一次
	if _, ok := redisPool[module]; !ok {
		redisPool[module] = initRedis(module, host, port, auth)
	}
}

func ExecDefault(cmd string, key string, args ...interface{}) (interface{}, error) {
	return exec(cmd, "default", key, args...)
}

func ClientDefault() (red.Conn, error) {
	return getClient("default")
}
