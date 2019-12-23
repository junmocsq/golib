package cache

import red "github.com/gomodule/redigo/redis"

var redisPool map[string]*red.Pool = make(map[string]*red.Pool)

func init() {
	// 注册redis default模块
	registerRedisPool("default")

	// 注册redis sql模块
	//registerRedisPool("sql")
}

func registerRedisPool(module string) {
	// 只注册一次
	if _, ok := redisPool[module]; !ok {
		redisPool[module] = initRedis(module)
	}
}

func ExecDefault(cmd string, key string, args ...interface{}) (interface{}, error) {
	return exec(cmd, "default", key, args...)
}

func ClientDefault() (red.Conn, error) {
	return getClient("default")
}
