package model


import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_"github.com/go-sql-driver/mysql"
)

var Db *sqlx.DB

func init()  {
	database,err := sqlx.Open("mysql","root:sql123@tcp(127.0.0.1:3306)/ginblog")
	if err != nil {
		fmt.Println("打开MySQL失败",err)
		return
	}
	Db = database
	fmt.Println("MySQL数据库打开成功！")
}

