package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

type config struct {
	loginCode string
	downNum   int
	downHost  string
	dbHost    string
	dbPort    string
	dbUser    string
	dbPass    string
	location  string
}

func initDb(con *config) {
	var err error
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/yuanxin", con.dbUser, con.dbPass, con.dbHost, con.dbPort)
	Db, err = sql.Open("mysql", url)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	con := getArgs()
	initDb(con)

	// find user
	sql := fmt.Sprintf("select id from zyb_employee where code = '%s';", con.loginCode)

	fmt.Println("查询用户: " + sql)
	var empId int
	err := Db.QueryRow(sql).Scan(&empId)
	if err != nil {
		log.Println("查询用户错误")
		panic(err)
	}

	sql = fmt.Sprintf("select  concat('http://%s/api/eop-download/',path)  as url, name from zyb_message_attachment where zyb_employee_id = %d order by id desc limit %d;", con.downHost, empId, con.downNum)

	fmt.Println("查询附件: " + sql)

	rows, err := Db.Query(sql)

	if err != nil {
		log.Println("查询附件错误")
		panic(err)
	}
	defer rows.Close()

	var wg sync.WaitGroup
	for rows.Next() {
		var url string
		var name string
		rows.Scan(&url, &name)
		wg.Add(1)
		go download(url, name, con.location, &wg)

	}
	wg.Wait()

}

func download(url string, name string, localtion string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := resty.New()
	fmt.Printf("开始下载: %s url: %s\n", name, url)

	// HTTP response gets saved into file, similar to curl -o flag
	_, err := client.R().SetOutput(localtion + name).Get(url)

	if err != nil {
		fmt.Printf("文件下载错误,err: %v\n", err)
	}
}

func getArgs() *config {

	con := config{}

	flag.StringVar(&con.loginCode, "u", getStringArgFromEnv("loginCode", "dc5"), "账号, "+envHelpMsg("loginCode"))
	flag.IntVar(&con.downNum, "n", getIntArgFromEnv("downNum", 1), "下载数量")
	flag.StringVar(&con.downHost, "ip", getStringArgFromEnv("downHost", "40.18.14.8"), "下载ip, "+envHelpMsg("downHost"))
	flag.StringVar(&con.location, "l", getStringArgFromEnv("location", "./"), "下载位置, "+envHelpMsg("localtion"))
	flag.StringVar(&con.dbHost, "dh", getStringArgFromEnv("dbHost", "40.18.14.196"), "数据库ip, "+envHelpMsg("dbHost"))
	flag.StringVar(&con.dbPort, "dp", getStringArgFromEnv("dbPort", "13306"), "数据库 port, "+envHelpMsg("dbPort"))
	flag.StringVar(&con.dbUser, "du", getStringArgFromEnv("dbUser", "root"), "数据库用户, "+envHelpMsg("dbUser"))
	flag.StringVar(&con.dbPass, "dP", getStringArgFromEnv("dbPass", "xxxxx"), "数据库密码, "+envHelpMsg("dbPass"))
	flag.Parse()
	return &con
}

func getStringArgFromEnv(envName string, defaultValue string) string {
	v := os.Getenv(getEnvName(envName))
	if len(v) == 0 {
		return defaultValue
	}
	return v

}

func getIntArgFromEnv(envName string, defaultValue int) int {
	v := os.Getenv(getEnvName(envName))
	if len(v) == 0 {
		return defaultValue
	}
	iv, err := strconv.Atoi(v)
	if err != nil {
		fmt.Printf("从环境变量读取 %s 值错误，不是一个合法的数字\n", envName)
		panic(err)
	}
	return iv
}

func getEnvName(envName string) string {
	return "yx_" + envName
}

func envHelpMsg(envName string) string {
	return "支持设置环境变量 " + getEnvName(envName) + " 修改默认值"
}
