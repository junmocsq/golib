package cache

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"testing"
)

func init() {
	// 注册redis default模块 必须注册
	RegisterRedisPool("default", "127.0.0.1", "6379", "")

	// 注册redis sql模块
	RegisterRedisPool("sql", "127.0.0.1", "6379", "")
	RegisterPrefixKey("junmo")
}

func TestClient(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			c, err := getClient("default")
			if err != nil {
				t.Log(err)
			}
			c.Close()
		}()
		go func() {
			c, err := getClient("sql")
			if err != nil {
				t.Log(err)
			}
			c.Close()
		}()
	}
}

func TestRedisSet(t *testing.T) {
	client, err := Client("sql")
	if err != nil {
		return
	}
	fmt.Println(ExecDefault("SET", "csq", "lixiaoerssss"))

	fmt.Println(redis.String(CustomConnExec(client, "GET", "csq")))

	fmt.Println(redis.String(client.Do("GET", GetKey("csq"))))
	client.Close()

}
