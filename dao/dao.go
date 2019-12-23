package dao

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"github.com/junmocsq/golib/cache"
	"io"
	"time"
)

// 缓存前缀
const TAG_PREFIX = "DISTR_"

// 缓存过期时间
const EXPIRE_TIME = 300

type dao struct {
	tag        string
	key        string
	isCacheKey bool
	sql        string
	params     []interface{}
	db         string
	o          orm.Ormer
}

// 不能并发使用这个dao 不是串行执行的必须每一个都NewDao
func NewDao() *dao {
	return &dao{}
}

func (d *dao) SetTag(tag string) *dao {
	return d.setTag(tag)
}

func (d *dao) setTag(tag string) *dao {
	d.tag = TAG_PREFIX + tag
	return d
}

func (d *dao) SetDb(dbname string) *dao {
	return d.setDb(dbname)
}

func (d *dao) setDb(dbname string) *dao {
	d.db = dbname
	return d
}

func (d *dao) SetKey(key string) *dao {
	return d.setKey(key)
}

func (d *dao) setKey(key string) *dao {
	d.key = key
	d.isCacheKey = true
	return d
}

func (d *dao) PrepareSql(sql string, params ...interface{}) *dao {
	return d.prepareSql(sql, params...)
}

func (d *dao) prepareSql(sql string, params ...interface{}) *dao {
	d.sql = sql
	d.params = params
	if d.isCacheKey {
		d.key = d.tag + d.key
	} else if d.tag != "" {
		// 从redis获取tag的value
		tagRes, err := cache.ExecDefault("GET", d.tag)

		if err != nil {
			logs.GetLogger("dao").Printf("prepareSql::redis get err:%v", err)
			return d
		}
		var tagStrRes string
		// tag 不存在redis value 则设置
		if tagRes == nil {
			tagStrRes = d.tag + time.Now().String()
			md := md5.New()
			_, err = io.WriteString(md, tagStrRes)
			if err != nil {
				logs.GetLogger("dao").Printf("prepareSql::io WriteString err:%v", err)
				return d
			}
			tagStrRes = fmt.Sprintf("%x", md.Sum(nil))
			_, err := cache.ExecDefault("SET", d.tag, tagStrRes, "EX", EXPIRE_TIME)
			if err != nil {
				logs.GetLogger("dao").Printf("prepareSql::redis set err:%v", err)
				return d
			}
		} else {
			tagStrRes, _ = redis.String(tagRes, err)
		}
		str := sql
		if params != nil {
			paramsStr, _ := json.Marshal(params)
			str += string(paramsStr)
		}
		md := md5.New()
		str = tagStrRes + str
		_, err = io.WriteString(md, str)
		if err != nil {
			logs.GetLogger("dao").Printf("prepareSql::io WriteString err:%v", err)
			return d
		}
		d.key = fmt.Sprintf("%x", md.Sum(nil))
	}
	return d
}

func (d *dao) setResult(result interface{}) {
	if d.key != "" {
		res, err := json.Marshal(result)
		if err != nil {
			logs.GetLogger("dao").Printf("setResult::json 转换 err:%v", err)
			return
		}
		if string(res) != "" {
			_, err = cache.ExecDefault("SET", d.key, string(res), "EX", EXPIRE_TIME)
			if err != nil {
				logs.GetLogger("dao").Printf("setResult::redis set err:%v", err)
				return
			}
		}
	}
}

func (d *dao) getResult(result interface{}) bool {
	if d.tag == "" {
		return false
	}
	res, err := cache.ExecDefault("GET", d.key)
	if err != nil {
		logs.GetLogger("dao").Printf("getResult::redis get err:%v", err)
		return false
	}
	str, _ := redis.String(res, err)
	if str == "" {
		return false
	}
	err = json.Unmarshal([]byte(str), &result)
	if err != nil {
		logs.GetLogger("dao").Printf("getResult::json 解析 err:%v", err)
		return false
	}
	return true
}

func (d *dao) FetchOne(model ...interface{}) bool {
	return d.fetchOne(model...)
}
func (d *dao) fetchOne(model ...interface{}) bool {
	defer d.clear()
	if !d.getResult(&model) {
		if d.o == nil {
			d.o = orm.NewOrm()
		}
		if d.db != "" {
			err := d.o.Using(d.db)
			if err != nil {
				logs.GetLogger("dao").Printf("fetchOne::using db err:%v", err)
				return false
			}
		}
		err := d.o.Raw(d.sql, d.params...).QueryRow(model...)
		if err != nil {
			logs.GetLogger("dao").Printf("fetchOne::sql:%s err:%v", d.sql, err)
			return false
		} else {
			d.setResult(model)
			return true
		}
	}
	return true
}

func (d *dao) FetchAll(model ...interface{}) bool {
	return d.fetchAll(model...)
}
func (d *dao) fetchAll(model ...interface{}) bool {
	defer d.clear()
	if d.getResult(&model) {
		return true
	}
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	if d.db != "" {
		err := d.o.Using(d.db)
		if err != nil {
			logs.GetLogger("dao").Printf("fetchAll::using db err:%v", err)
			return false
		}
	}
	n, err := d.o.Raw(d.sql, d.params...).QueryRows(model...)
	if err != nil || n == 0 {
		if err != nil {
			logs.GetLogger("dao").Printf("fetchAll::sql:%s err:%v", d.sql, err)
		}
		return false
	} else {
		d.setResult(model)
		return true
	}
}
func (d *dao) Insert() (int64, error) {
	return d.insert()
}
func (d *dao) insert() (int64, error) {
	defer d.clear()
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	if d.db != "" {
		err := d.o.Using(d.db)
		if err != nil {
			logs.GetLogger("dao").Printf("insert::using db err:%v", err)
			return 0, err
		}
	}
	res, err := d.o.Raw(d.sql, d.params...).Exec()
	if err != nil {
		logs.GetLogger("dao").Printf("insert::sql:%s err:%v", d.sql, err)
		return 0, err
	}
	d.clearTag()
	return res.LastInsertId()
}

// replace 获取影响行数
func (d *dao) Add() (int64, error) {
	return d.add()
}
func (d *dao) add() (int64, error) {
	defer d.clear()
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	if d.db != "" {
		err := d.o.Using(d.db)
		if err != nil {
			logs.GetLogger("dao").Printf("add::using db err:%v", err)
			return 0, err
		}
	}
	res, err := d.o.Raw(d.sql, d.params...).Exec()
	if err != nil {
		logs.GetLogger("dao").Printf("add::sql:%s err:%v", d.sql, err)
		return 0, err
	}
	d.clearTag()
	return res.RowsAffected()
}

func (d *dao) Update() (int64, error) {
	return d.update()
}
func (d *dao) update() (int64, error) {
	defer d.clear()
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	if d.db != "" {
		err := d.o.Using(d.db)
		if err != nil {
			logs.GetLogger("dao").Printf("update::using db err:%v", err)
			return 0, err
		}
	}
	res, err := d.o.Raw(d.sql, d.params...).Exec()
	if err != nil {
		logs.GetLogger("dao").Printf("update::sql:%s err:%v", d.sql, err)
		return 0, err
	}
	d.clearTag()
	return res.RowsAffected()
}

func (d *dao) Delete() (int64, error) {
	return d.delete()
}
func (d *dao) delete() (int64, error) {
	defer d.clear()
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	if d.db != "" {
		err := d.o.Using(d.db)
		if err != nil {
			logs.GetLogger("dao").Printf("delete::using db err:%v", err)
			return 0, err
		}
	}
	res, err := d.o.Raw(d.sql, d.params...).Exec()
	if err != nil {
		logs.GetLogger("dao").Printf("delete::sql:%s err:%v", d.sql, err)
		return 0, err
	}
	d.clearTag()
	return res.RowsAffected()
}

func (d *dao) Begin() error {
	return d.begin()
}

func (d *dao) begin() error {
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	return d.o.Begin()
}

func (d *dao) Rollback() error {
	return d.rollback()
}

func (d *dao) rollback() error {
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	return d.o.Rollback()
}

func (d *dao) Commit() error {
	return d.commit()
}

func (d *dao) commit() error {
	if d.o == nil {
		d.o = orm.NewOrm()
	}
	return d.o.Commit()
}

func (d *dao) ClearTag() {
	d.clearTag()
}

// 清除缓存
func (d *dao) clearTag() {
	if d.isCacheKey {
		_, err := cache.ExecDefault("DEL", d.key)
		if err != nil {
			logs.GetLogger("dao").Printf("clearTag::redis err:%v", err)
		}
	} else if d.tag != "" {
		_, err := cache.ExecDefault("DEL", d.tag)
		if err != nil {
			logs.GetLogger("dao").Printf("clearTag::redis err:%v", err)
		}
	}
}

// mysql执行完的扫尾操作
func (d *dao) clear() {
	if d.tag != "" {
		_, err := cache.ExecDefault("EXPIRE", d.tag, EXPIRE_TIME)
		if err != nil {
			logs.GetLogger("dao").Printf("clear::redis err:%v", err)
		}
	}
	if d.key != "" {
		_, err := cache.ExecDefault("EXPIRE", d.key, EXPIRE_TIME)
		if err != nil {
			logs.GetLogger("dao").Printf("clear::redis err:%v", err)
		}
	}
	d.isCacheKey = false
	d.tag = ""
	d.key = ""
	d.db = ""
	d.sql = ""
	d.params = nil
}
