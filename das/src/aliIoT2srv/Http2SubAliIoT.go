package feiyan2srv

// 通过HTTP/2订阅阿里云飞燕平台的数据


import (
	"encoding/hex"
	"github.com/chmike/hmacsha256"
	"github.com/dlintw/goconf"
	"github.com/labstack/gommon/log"
	"github.com/summerwind/h2spec/spec"
	"github.com/summerwind/h2spec/config"
	"golang.org/x/net/http2"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
	"fmt"
	"../core/jobque"
)

var (
	framer *http2.Framer
	start_time time.Time
	last_time time.Time
)
func AliIoT2Http2Start(conf *goconf.ConfigFile)  {
	// 获取aliIoT2http2
	endPoint, err := conf.GetString("aliIoT2http2", "endPoint")
	if err != nil {
		log.Errorf("读取aliIoT2http2配置失败，%s\n", err)
		os.Exit(1)
	}

	appKey, _ := conf.GetString("aliIoT2http2", "appKey")
	appSecret, _ := conf.GetString("aliIoT2http2", "telecom2https_port")

	conn, err := connect2AliIoT(endPoint, appKey, appSecret)

	framer = http2.NewFramer(conn.Conn, conn.Conn)
	start_time = time.Now()
	last_time = time.Now()
	// 没有数据的时候，为了维持长连接，发送PING包到阿里云飞燕平台
	go func() {
		for {
			time.Sleep(1*time.Second)
			start_time = time.Now()
			tm_spand := start_time.Unix() - last_time.Unix()
			if tm_spand >= 30 {
				data := [8]byte{'h', '2', 's', 'p', 'e', 'c'}
				conn.WritePing(false, data)
			}
		}
	}()

	// 接收HTTP2推送的数据
	go func() {
		for {
			time.Sleep(1*time.Second)
			f, err := framer.ReadFrame()
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				fmt.Println("err=", err)
				break
			}
			switch err.(type) {
			case nil:
				last_time = time.Now()
				fmt.Println(f)
				df, ok := f.(*http2.DataFrame)
				if ok {
					fmt.Println(df, string(df.Data()))
					// 加入处理队列
					jobque.JobQueue <- NewAliIoTJob(df.Data())
				}
			case http2.ConnectionError:
				// Ignore. There will be many errors of type "PROTOCOL_ERROR, DATA
				// frame with stream ID 0". Presumably we are abusing the framer.
				fmt.Println("err:", err)
			default:
				fmt.Println(err, framer.ErrorDetail())
			}
		}
	}()
}

type AliIoTJob struct {
}

func NewAliIoTJob(rawData []byte) AliIoTJob {
	return AliIoTJob{}
}

func (h2d AliIoTJob) Handle() {
	// AliIoTJob 消息处理

}

func connect2AliIoT(endPoint, appKey, appSecret string) (conn spec.Conn, err error) {
	profile := NewAppKeyProfile(endPoint, appKey, appSecret)

	var streamID uint32 = 1

	c :=  config.Config {
		Host: profile.EndPoint,
		Port: 443,
		Path: "/message/echo/success",
		Timeout: 30 * time.Second,
		TLS: true,
	}

	&conn, err = spec.Dial(&c)
	if err != nil {
		fmt.Println(err)
		return conn, err
	}

	err1 := conn.Handshake()
	if err1 != nil {
		fmt.Println(err1)
		return conn, err1
	}

	headers := spec.CommonHeaders(&c)
	var random int64
	random = rand.Int63()
	signContent := "random=" + strconv.FormatInt(random, 10);
	sign := hmacsha256.Digest(nil, []byte(appSecret), []byte(signContent))
	headers = append(headers, spec.HeaderField("x-auth-name", "appkey"))
	headers = append(headers, spec.HeaderField("x-auth-param-app-key", appKey))
	headers = append(headers, spec.HeaderField("x-auth-param-sign-method", "SHA256"))
	headers = append(headers, spec.HeaderField("x-auth-param-random", strconv.FormatInt(random, 10)))
	headers = append(headers, spec.HeaderField("x-auth-param-sign", hex.EncodeToString(sign)))
	headers = append(headers, spec.HeaderField("x-clear-session", "0"))
	headers = append(headers, spec.HeaderField("content-length", "0"))
	hp := http2.HeadersFrameParam{
		StreamID:      streamID,
		EndStream:     true,
		EndHeaders:    true,
		BlockFragment: conn.EncodeHeaders(headers),
	}
	conn.WriteHeaders(hp)

	return conn, nil
}