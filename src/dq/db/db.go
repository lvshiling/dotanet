package db

import (
	"database/sql"
	"dq/conf"
	"dq/log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	Mydb *sql.DB
}

var DbOne *DB

func CreateDB() {
	DbOne = new(DB)
	DbOne.Init()
}

func (a *DB) Init() {

	ip := conf.Conf.DataBaseInfo["Ip"].(string)
	nameandpassword := conf.Conf.DataBaseInfo["NameAndPassword"].(string)
	databasename := conf.Conf.DataBaseInfo["DataBaseName"].(string)
	db, err := sql.Open("mysql", nameandpassword+"@"+ip+"/"+databasename)
	if err != nil {
		log.Error(err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Error(err.Error())
	}
	a.Mydb = db

	a.Mydb.SetMaxOpenConns(10000)
	a.Mydb.SetMaxIdleConns(500)
	a.Mydb.Ping()
}

//创建快速新玩家
func (a *DB) CreateQuickPlayer(machineid string, platfom string, name string) int {

	id, _ := a.newUser(machineid, platfom, "", "", name)

	return id
}

//创建新玩家基础信息
func (a *DB) newUserBaseInfo(id int, name string) error {

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("INSERT userbaseinfo (uid,name) values (?,?)",
		id, name)
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("INSERT userbaseinfo err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

//创建新玩家信息
func (a *DB) newUser(machineid string, platfom string, phonenumber string, openid string, name string) (int, error) {

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("INSERT user (phonenumber,platform,machineid,wechat_id) values (?,?,?,?)",
		phonenumber, platfom, machineid, openid)
	n, e := res.RowsAffected()
	id, err2 := res.LastInsertId()
	if err1 != nil || n == 0 || e != nil || err2 != nil {
		log.Info("INSERT user err")
		return -1, tx.Rollback()
	}
	if name == "" {
		name = "yk_" + strconv.Itoa(int(id))
	}

	//day := time.Now().Format("2006-01-02")

	res, err1 = tx.Exec("INSERT userbaseinfo (uid,name) values (?,?)",
		id, name)
	//插入名字失败
	if err1 != nil {

		name = "yk_" + strconv.Itoa(int(id))
		res, err1 = tx.Exec("INSERT userbaseinfo (uid,name) values (?,?)",
			id, name)
		if err1 != nil {
			log.Info("INSERT userbaseinfo err")
			return -1, tx.Rollback()
		}
	}
	n, e = res.RowsAffected()
	if n == 0 || e != nil {
		log.Info("INSERT userbaseinfo err")
		return -1, tx.Rollback()
	}

	err1 = tx.Commit()
	if err1 == nil {
		return int(id), nil
	}
	return -1, err1

}

//检查快速登录
func (a *DB) CheckQuickLogin(machineid string, platfom string) int {
	var uid int

	stmt, err := a.Mydb.Prepare("SELECT uid FROM user where BINARY (machineid=? and platform=?)")

	if err != nil {
		log.Info(err.Error())
		return -1
	}
	defer stmt.Close()
	rows, err := stmt.Query(machineid, platfom)
	if err != nil {
		log.Info(err.Error())
		return uid
		//创建新账号
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&uid)
	} else {
		log.Info("no user:%s,%s", machineid, platfom)
	}

	return uid

}

func (a *DB) test() {

}

func (a *DB) Close() {
	a.Mydb.Close()
}
