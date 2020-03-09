package cache

import red "github.com/gomodule/redigo/redis"

var redisPool map[string]*red.Pool = make(map[string]*red.Pool)
var prefixKey = ""

func init() {
	// 注册redis default模块 必须注册
	//RegisterRedisPool("default","127.0.0.1","6379","")

	// 注册redis sql模块
	//registerRedisPool("sql","127.0.0.1","6379","")
}

func RegisterRedisPool(module, host, port, auth string) {
	// 只注册一次
	if _, ok := redisPool[module]; !ok {
		redisPool[module] = initRedis(module, host, port, auth)
	}
}

// 用户的一次完整的redis操作，不需要手动关闭redis连接[redis连接自动回收回连接池]
// 此方法key不用GetKey处理
func Exec(cmd string, module string, key string, args ...interface{}) (interface{}, error) {
	return exec(cmd, module, GetKey(key), args...)
}

func Client(module string) (red.Conn, error) {
	return getClient(module)
}

// 用户的一次完整的redis操作，不需要手动关闭redis连接[redis连接自动回收回连接池]
// 此方法key不用GetKey处理
func ExecDefault(cmd string, key string, args ...interface{}) (interface{}, error) {
	return exec(cmd, "default", GetKey(key), args...)
}

func ClientDefault() (red.Conn, error) {
	return getClient("default")
}

// 用户自定义连接操作redis，需要手动关闭redis连接[即回收redis连接回连接池]
// 此方法key不用GetKey处理
func CustomConnExec(con red.Conn, cmd string, key string, args ...interface{}) (interface{}, error) {
	if err := con.Err(); err != nil {
		return nil, err
	}
	parmas := make([]interface{}, 0)
	parmas = append(parmas, GetKey(key))
	if len(args) > 0 {
		for _, v := range args {
			parmas = append(parmas, v)
		}
	}
	return con.Do(cmd, parmas...)
}

// 设置redis key前缀
// 当用户注册了key前缀，直接调用redis方法时 key一定要用GetKey返回的key来处理
func RegisterPrefixKey(prefix string) {
	prefixKey = prefix
}

// 获取redis key
func GetKey(key string) string {
	return prefixKey + "_" + key
}
