package connpool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ClosedErr = errors.New("the cli is closed")
)

type ConnPooler interface {
	Begin()    //initial DB pool
	Shutdown() //shutdown DB pool

	Get() interface{}        //acquire connection
	Put(interface{})         //release connnection
	Close(interface{}) error //close connection
}

type ConnPoolConfig struct {
	MaxCap          int32
	InitCap         int32
	MaxConnLifeTime time.Duration
}

func NewPoolConfig(maxCap, initCap int32, maxConnLifeTime time.Duration) ConnPoolConfig {
	return ConnPoolConfig{
		MaxCap:          maxCap,
		InitCap:         initCap,
		MaxConnLifeTime: maxConnLifeTime,
	}
}

type ConnCli struct {
	cli interface{}
	t   time.Time
}

type factoryFunc func() (interface{}, error)
type closeFunc func(interface{}) error
type isCliValidFunc func(interface{}) bool

func newConnCli(cli interface{}) *ConnCli {
	return &ConnCli{
		cli: cli,
		t:   time.Now(),
	}
}

type ConnPool struct {
	factory    factoryFunc
	close      closeFunc
	isCliValid isCliValidFunc

	poolConfig ConnPoolConfig
	pool       chan *ConnCli

	mu      sync.Mutex
	currNum int32
}

func NewDBPool(factory factoryFunc, close closeFunc, isCliValid isCliValidFunc, config ConnPoolConfig) *ConnPool {
	return &ConnPool{
		factory:    factory,
		close:      close,
		isCliValid: isCliValid,
		poolConfig: config,
		pool:       make(chan *ConnCli, config.MaxCap),
	}
}

func (p *ConnPool) Begin() {

	if p.poolConfig.MaxCap < p.poolConfig.InitCap {
		p.poolConfig.MaxCap = p.poolConfig.InitCap
	}

	if p.poolConfig.MaxCap < 0 {

	}

	wg := new(sync.WaitGroup)

	var i int32
	for i = 0; i < p.poolConfig.InitCap; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cli, err := p.factory()

			if err != nil {

			}

			atomic.AddInt32(&p.currNum, 1)
			// p.mu.Lock()
			// p.currNum++
			// p.mu.Unlock()
			p.pool <- newConnCli(cli)
		}()
	}

	wg.Wait()
}

func (p *ConnPool) Get() interface{} {

	for {
		select {
		case cli := <-p.pool:
			if cli == nil {
				return nil
			}

			if lifeTime := p.poolConfig.MaxConnLifeTime; lifeTime > 0 {
				if cli.t.Add(lifeTime).Before(time.Now()) {
					//连接超过生命周期
					p.Close(cli.cli)
					continue
				}

				if p.isCliValid != nil {
					if p.isCliValid(cli.cli) {
						return cli.cli
					} else {
						p.Close(cli.cli)
						continue
					}
				} else {
					return cli.cli
				}
			}

			return cli.cli
		default:
			p.mu.Lock()
			if p.currNum < p.poolConfig.MaxCap {
				cli, err := p.factory()
				if err != nil || cli == nil {
					p.mu.Unlock()
					continue
				}
				p.currNum++
				p.mu.Unlock()
				return cli
			}

			p.mu.Unlock()
			continue
		}
	}
}

func (p *ConnPool) Put(cli interface{}) {

	if cli == nil {
		atomic.AddInt32(&p.currNum, 1)
		// p.mu.Lock()
		// p.currNum++
		// p.mu.Unlock()
		return
	}

	select {
	case p.pool <- newConnCli(cli):
		return
	default:
		p.Close(cli)
	}
}

func (p *ConnPool) Close(cli interface{}) error {
	atomic.AddInt32(&p.currNum, -1)
	// p.mu.Lock()
	// p.currNum--
	// p.mu.Unlock()
	return p.close(cli)
}

func (p *ConnPool) Shutdown() {

	close(p.pool)
	for cli := range p.pool {
		p.close(cli.cli)
	}

}
