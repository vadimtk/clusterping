package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

var user string
var password string

var dsnnodes []string
var pings = 1000

type PingTime struct {
	TimeDiff string `db:"timediff"`
	SoftTime int64  `db:"softtime"`
}

func pinger(idx int) {
	defer wg.Done()

	var db1 *sqlx.DB
	dsn := user + ":" + password + "@tcp(" + dsnnodes[idx] + ")/__percona"
	db1 = sqlx.MustConnect("mysql", dsn)
	defer db1.Close()

	fmt.Println("Running pinger to " + dsn)

	td := PingTime{}
	i := 1
	for i < pings {
		//err := db1.Get(&td, "SELECT NOW(6)-pingtime AS timediff FROM __clusterping WHERE id="+strconv.Itoa(i))
		err := db1.Get(&td, "SELECT softtime FROM __clusterping WHERE id=?", i)
		if err != nil {
			continue
			//      panic(err.Error())
		}
		//fmt.Printf("node: %d, ping: %d, time: %f\n", idx, i, (time.Now().UnixNano()-td.SoftTime)/(time.Microsecond))
		fmt.Printf("node: %d, ping: %d, time: %f\n", idx, i, time.Duration(time.Now().UnixNano()-td.SoftTime).Seconds())
		i = i + 1
	}

	fmt.Println("Done")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	nodesPtr := flag.String("nodes", "", "list of nodes (in HOST:PORT format)")
	hostPtr := flag.String("host", "", "the primary host to connect (in HOST:PORT format)")
	userPtr := flag.String("user", "root", "database user")
	passwordPtr := flag.String("password", "", "database password")

	flag.Parse()
	dsnnodes = strings.Split(*nodesPtr, ",")
	user = *userPtr
	password = *passwordPtr

	var db *sqlx.DB
	dsn := user + ":" + password + "@tcp(" + *hostPtr + ")"
	db = sqlx.MustConnect("mysql", dsn+"/mysql")
	defer db.Close()

	db.MustExec("CREATE DATABASE IF NOT EXISTS __percona")

	var db1 *sqlx.DB
	db1 = sqlx.MustConnect("mysql", dsn+"/__percona")
	defer db1.Close()

	db1.MustExec("DROP TABLE IF EXISTS __clusterping")
	db1.MustExec("CREATE TABLE __clusterping (id int not null primary key, pingtime DATETIME(6) not null, softtime bigint not null)")

	for i, c := range dsnnodes {
		fmt.Printf("Start pinger: %d, %s\n", i, c)
		wg.Add(1)
		go pinger(i)
	}

	for i := 1; i < pings; i++ {
		fmt.Printf("ping: %d\n", i)
		db1.MustExec("INSERT INTO __clusterping (id,pingtime,softtime) VALUES (?,NOW(6),?)", strconv.Itoa(i), time.Now().UnixNano())
		time.Sleep(1 * time.Second)
	}

	wg.Wait()
}
