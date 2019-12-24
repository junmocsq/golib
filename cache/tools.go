package cache

import "strconv"

func HSCAN(module, key string) (map[string]string, error) {
	conn, err := getClient(module)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	var iter int
	var resMap = make(map[string]string)
	for {
		res, err := conn.Do("HSCAN", key, iter)
		if err != nil {
			break
		}
		if temp, ok := res.([]interface{}); ok {
			if len(temp) == 0 {
				break
			}
			if len(temp) == 2 {
				if t, ok := temp[0].([]uint8); ok {
					iterTemp, err := strconv.ParseInt(string(t), 0, 0)
					if err != nil {

					}
					iter = int(iterTemp)
				}

				if t, ok := temp[1].([]interface{}); ok {
					for i := 0; i < len(t); i += 2 {
						key := t[i]
						value := t[i+1]
						var k, v string
						if key, ok := key.([]byte); ok {
							k = string(key)

						}
						if value, ok := value.([]byte); ok {
							v = string(value)
						}
						resMap[k] = v
					}
				}
			}
		} else {
			break
		}
		if iter == 0 {
			break
		}
	}
	return resMap, nil
}
