package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"das/core/log"
)

var (
	mysqlPool  MysqlPool
	onceDBPool sync.Once
	ErrConn    = errors.New("db connected fail")
)

type MysqlPool struct {
	mysqlPool     *sql.DB
	url           string
	maxPoolSize   uint64
	currReconnNum int32

	ctxP       context.Context
	cancelP    context.CancelFunc
	ctxC       context.Context
	cancelC    context.CancelFunc
	mu         sync.Mutex
	reconnFlag bool
}

func Init() {
	onceDBPool.Do(initDBPool)
}

func initDBPool() {
	mysqlPool = newMysqlPool(2)
}

func DoMysqlBegin() (*sql.Tx, error) {
	return mysqlPool.Begin()
}

func DoMysqlExec(query string, args ...interface{}) (res sql.Result, err error) {
	return mysqlPool.DoMysqlExec(query, args...)
}

func DoMysqlQuery(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return mysqlPool.DoMysqlQuery(query, args...)
}

func DoMysqlRowQuery(query string, args ...interface{}) (row *sql.Row) {
	return mysqlPool.DoMysqlRowQuery(query, args...)
}

func getMysqlURI() string {
	url, _ := log.Conf.GetString("mysql", "url")
	userName, _ := log.Conf.GetString("mysql", "user")
	pwd, _ := log.Conf.GetString("mysql", "pwd")
	dbName, _ := log.Conf.GetString("mysql", "dbName")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", userName, pwd, url, dbName)
	return dsn
}

func newMysqlPool(maxPoolSize uint64) MysqlPool {
	mysqlUrl := getMysqlURI()
	mysqlPool := newMysqlClient(mysqlUrl, maxPoolSize)
	if mysqlPool == nil {
		panic(ErrConn)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctxC, cancelC := context.WithCancel(ctx)
	return MysqlPool{
		mysqlPool:   mysqlPool,
		url:         mysqlUrl,
		maxPoolSize: maxPoolSize,
		ctxP:        ctx,
		cancelP:     cancel,
		ctxC:        ctxC,
		cancelC:     cancelC,
	}
}

func newMysqlClient(url string, maxPoolSize uint64) *sql.DB {
	pool, err := sql.Open("mysql", url)
	if err != nil {
		panic(err)
	}

	pool.SetMaxOpenConns(int(maxPoolSize))

	err = pool.Ping()
	if err != nil {
		panic(err)
	}
	return pool
}

func (self *MysqlPool) Begin() (*sql.Tx, error) {
	return self.mysqlPool.Begin()
}

func (self *MysqlPool) DoMysqlExec(query string, args ...interface{}) (res sql.Result, err error) {
	return self.mysqlPool.ExecContext(self.ctxC, query, args...)
}

func (self *MysqlPool) DoMysqlRowQuery(query string, args ...interface{}) (row *sql.Row) {
	return self.mysqlPool.QueryRowContext(self.ctxC, query, args...)
}

func (self *MysqlPool) DoMysqlQuery(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return self.mysqlPool.Query(query, args...)
}

func (self *MysqlPool) reConn() error {
	isFirst := true
	//stop the msgs reproc
	self.mu.Lock()
	if self.reconnFlag {
		self.mu.Unlock()
		log.Debug("goroutine exit")
		runtime.Goexit()
	}
	//ChReprocEnd <- struct{}{}
	self.reconnFlag = true
	self.mu.Unlock()
	for {
		if !isFirst {
			time.Sleep(time.Second * 20)
		} else {
			isFirst = false
		}
		log.Infof("MysqlPool %dth Reconnecting...", self.currReconnNum+1)
		if self.mysqlPool = newMysqlClient(self.url, self.maxPoolSize); self.mysqlPool == nil {
			self.currReconnNum++
			continue
		} else {
			log.Info("MysqlPool Reconnect Successful")
			self.currReconnNum = 0
			self.ctxC, self.cancelC = context.WithCancel(self.ctxP)
			self.reconnFlag = false
			return nil
		}
	}

	return nil
}

func (self *MysqlPool) Close() {
	if self.mysqlPool != nil {
		self.mysqlPool.Close()
	}
	self.cancelP()
}

func FormatQuerySql(sqlData string, lens int) string {
	return fmt.Sprintf(sqlData, placeholders(lens))
}

func placeholders(n int) string {
	var b strings.Builder
	for i := 0; i < n-1; i++ {
		b.WriteString("?,")
	}
	if n > 0 {
		b.WriteString("?")
	}
	return b.String()
}

func Close() {
	mysqlPool.Close()
}
