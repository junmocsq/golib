package cache

import (
	"fmt"
	"github.com/astaxie/beego"
	red "github.com/gomodule/redigo/redis"
	"time"
)

type config struct {
	host string
	port string
	auth string
}

type GRPool struct {
}

// 基于beego 格式为
//redis.default.host = 127.0.0.1
//redis.default.port = 6382
//redis.default.auth =
func getConfig(module string) config {
	conf := config{
		host: beego.AppConfig.DefaultString("redis."+module+".host", "127.0.0.1"),
		port: beego.AppConfig.DefaultString("redis."+module+".port", "6379"),
		auth: beego.AppConfig.DefaultString("redis." + module + ".auth",""),

	}
	return conf
}

func initRedis(module string) *red.Pool {
	fmt.Println("init redis ", module, "pool")
	conf := getConfig(module)
	pool := &red.Pool{
		MaxIdle:     256,  // 最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态。
		MaxActive:   1000, // 最大的连接数，表示同时最多有N个连接。0表示不限制。
		IdleTimeout: time.Duration(120),
		Wait:        true,
		Dial: func() (red.Conn, error) {
			return red.Dial(
				"tcp",
				conf.host+":"+conf.port,
				red.DialReadTimeout(time.Duration(1000)*time.Millisecond),
				red.DialWriteTimeout(time.Duration(1000)*time.Millisecond),
				red.DialConnectTimeout(time.Duration(1000)*time.Millisecond),
				red.DialDatabase(0),
				red.DialPassword(conf.auth),
			)
		},
	}
	return pool
}

func getPool(module string) *red.Pool {
	if c, ok := redisPool[module]; ok {
		return c
	}
	panic(module + " 不存在")
}

func exec(cmd string, module string, key string, args ...interface{}) (interface{}, error) {
	con := getPool(module).Get()
	if err := con.Err(); err != nil {
		return nil, err
	}
	defer con.Close()
	parmas := make([]interface{}, 0)
	parmas = append(parmas, key)
	if len(args) > 0 {
		for _, v := range args {
			parmas = append(parmas, v)
		}
	}
	return con.Do(cmd, parmas...)
}

func getClient(module string) (red.Conn, error) {
	con := getPool(module).Get()
	if err := con.Err(); err != nil {
		return nil, err
	}
	return con, nil
}
