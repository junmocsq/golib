package dao

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"testing"
)

const MYSQL_DB = "default"

func TestDao(t *testing.T) {
	db := beego.AppConfig.DefaultString("mysql_db2", "")
	orm.RegisterDataBase("default", "mysql", db, 30)

	u := &UsStatistics{}
	u.Add(1, 100)
	u.Add(2, 120)
	u.GetSts()
	fmt.Println(u.GetStsByUid(1))
	fmt.Println(u.UpdateReadTimeByUid(1, 20))
	fmt.Println(u.GetStsByUid(1))
}

/*
 CREATE TABLE `us_statistics` (
  `uid` int(11) NOT NULL,
  `read_time` int(11) NOT NULL DEFAULT '0' COMMENT '阅读时长',
  PRIMARY KEY (`uid`)
)
*/

type UsStatistics struct {
	Uid      int
	ReadTime int
}

func (this *UsStatistics) TableName() string {
	return "us_statistics"
}

func (this *UsStatistics) tag() string {
	return "default_us_statistics"
}

func (this *UsStatistics) Add(uid, readTime int) bool {
	sql := "REPLACE INTO " + this.TableName() + "(uid,read_time) VALUES(?,?)"
	d := NewDao()
	res, err := d.SetDb(MYSQL_DB).SetTag(this.tag()).PrepareSql(sql, uid, readTime).Add()
	if res > 0 && err == nil {
		return true
	} else {
		return false
	}
}

func (this *UsStatistics) GetSts() []UsStatistics {
	sql := "SELECT * FROM " + this.TableName()
	var hcs []UsStatistics
	d := NewDao()
	d.SetDb(MYSQL_DB).SetTag(this.tag()).PrepareSql(sql).FetchAll(&hcs)
	return hcs
}

func (this *UsStatistics) GetStsByUid(uid int) UsStatistics {
	sql := "SELECT * FROM " + this.TableName() + " WHERE uid = ?"
	var hcs UsStatistics
	d := NewDao()
	d.SetDb(MYSQL_DB).SetTag(this.tag()).SetKey(strconv.Itoa(uid)).PrepareSql(sql, uid).FetchOne(&hcs)
	return hcs
}

func (this *UsStatistics) UpdateReadTimeByUid(uid, readTime int) bool {
	sql := "UPDATE " + this.TableName() + " SET read_time = read_time + ? WHERE uid = ?"
	d := NewDao()
	res, err := d.SetDb(MYSQL_DB).SetTag(this.tag()).SetKey(strconv.Itoa(uid)).PrepareSql(sql, readTime, uid).Update()
	if res > 0 && err == nil {
		return true
	} else {
		return false
	}
}
