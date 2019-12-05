package h2client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"

	"../log"
)

const (
	ClientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

var (
	ReadFrameErr = errors.New("read frame error")
	TLSDialErr   = errors.New("tls dial error")
	ConnCloseErr = errors.New("conn is closed")
)

type H2Client struct {
	conn    net.Conn
	bw      *bufio.Writer
	br      *bufio.Reader
	framer  *http2.Framer
	hbuf    *bytes.Buffer
	henc    *hpack.Encoder
	hdec    *hpack.Decoder
	request *http.Request

	dialer    *net.Dialer
	tlsConfig *tls.Config

	rawurl string
	method string
	addr   string

	timeout time.Duration
	ctx     context.Context
	reCh    chan struct{}
}

func Newh2Client(ctx context.Context) *H2Client {
	h2 := &H2Client{
		dialer: &net.Dialer{},
		tlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"h2", "h2-16"},
		},
		timeout: 0,
		ctx:     ctx,
		reCh:    make(chan struct{}, 1000),
	}
	h2.request, _ = http.NewRequest("", "", nil)

	return h2
}

func (h2 *H2Client) init() error {
	urlData, err := url.Parse(h2.rawurl)
	if err != nil {
		return err
	}

	addr := urlData.Host + ":"
	if urlData.Scheme == "https" {
		addr = addr + "443"
	} else {
		addr = addr + "80"
	}

	h2.addr = addr

	//create tcp connection
	h2.conn, err = tls.DialWithDialer(h2.dialer, "tcp", h2.addr, h2.tlsConfig)
	if err != nil {
		log.Error("tls.DialWithDialer() error = ", err.Error())
		return TLSDialErr
	}

	h2.hdec = hpack.NewDecoder(4096, nil)
	h2.hbuf = bytes.NewBuffer(nil)

	h2.bw = bufio.NewWriter(h2.conn)
	h2.br = bufio.NewReader(h2.conn)

	h2.request.Method = h2.method
	h2.request.URL = urlData

	h2.framer = http2.NewFramer(h2.bw, h2.br)
	//h2.framer.ReadMetaHeaders = h2.hdec
	h2.henc = hpack.NewEncoder(h2.hbuf)

	return nil
}

func (h2 *H2Client) SetHeader(key, value string) {
	h2.request.Header.Set(key, value)
}

func (h2 *H2Client) SetAliHeader(appKey, appSecret string) {
	request := h2.request
	random := rand.Int63()
	signContent := "random=" + strconv.FormatInt(random, 10)

	hc := hmac.New(sha256.New, []byte(appSecret))
	hc.Write([]byte(signContent))
	sign := hc.Sum([]byte{})

	//request.Header.Set(":path", "/message/ack")
	request.Header.Set("x-auth-name", "appkey")
	request.Header.Set("x-auth-param-app-key", appKey)
	request.Header.Set("x-auth-name", "appkey")
	request.Header.Set("x-auth-param-sign-method", "SHA256")
	request.Header.Set("x-auth-param-random", strconv.FormatInt(random, 10))
	request.Header.Set("x-auth-param-sign", hex.EncodeToString(sign))
	request.Header.Set("x-clear-session", "0")
	request.Header.Set("content-length", "0")
}

func (h2 *H2Client) Do(method, rawurl string) error {
	h2.rawurl = rawurl
	h2.method = method

	if err := h2.do(); err != nil {
		log.Error("H2Client.do() error = ", err)
		return err
	}

	return nil
}

func (h2 *H2Client) Get(rawurl string) error {
	return h2.Do("GET", rawurl)
}

func (h2 *H2Client) Post(rawurl string) error {
	return h2.Do("POST", rawurl)
}

func (h2 *H2Client) do() error {

	if err := h2.init(); err != nil {
		log.Error("H2Client.init() error = ", err)
		return err
	}

	h2.sendPreface()
	h2.sendWriteSettings()
	h2.framer.WriteWindowUpdate(0, 1<<30)
	h2.bw.Flush()

	if err := h2.sendHeadersFrame(); err != nil {
		log.Error("H2Client.sendHeadersFrame() error = ", err)
		return err
	}

	go h2.ctxDetect()
	go h2.heartBeat()

	if err := h2.readLoop(); err != nil {
		log.Error("H2Client.readLoop() error = ", err)
		return err
	}

	return nil
}

func (h2 *H2Client) Close() error {
	var err error
	//if err = h2.hdec.Close(); err != nil {
	//	log.Error("h2.hdec.Close() error = ", err)
	//	return err
	//}
	h2.hbuf.Reset()
	if err = h2.bw.Flush(); err != nil {
		log.Error("h2.bw.Flush() error = ", err)
		return err
	}

	if err = h2.conn.Close(); err != nil {
		log.Error("h2.conn.Close() error = ", err)
	}

	log.Debug("h2client close")

	return nil
}

func (h2 *H2Client) sendPreface() {
	p := []byte(ClientPreface)
	_, err := h2.bw.Write(p)
	if err != nil {
		log.Error("h2.bw.Write() error = ", err)
	}
	//h2.bw.Flush()
}

func (h2 *H2Client) sendWriteSettings() {
	settings := []http2.Setting{
		http2.Setting{
			ID:  http2.SettingEnablePush,
			Val: 1,
		},
		http2.Setting{
			ID:  http2.SettingInitialWindowSize,
			Val: 4194304,
		},
		http2.Setting{
			ID:  http2.SettingMaxHeaderListSize,
			Val: 10485760,
		},
	}

	if err := h2.framer.WriteSettings(settings...); err != nil {
		log.Error("sendWriteSettings() error = ", err)
	}
}

func (h2 *H2Client) sendSettingAck() error {
	fmt.Println("sendSettingAck() start...")
	if err := h2.framer.WriteSettingsAck(); err != nil {
		log.Error("sendSettingAck() error = ", err)
		return err
	}
	h2.bw.Flush()
	return nil
}

func (h2 *H2Client) sendHeadersFrame() error {
	hdrs, err := h2.encodeHeaders(h2.request, true, "", actualContentLength(h2.request))
	if err != nil {
		log.Info("encodeHeaders() error = ", err)
		return err
	}

	err = h2.framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: hdrs,
		EndStream:     true,
		EndHeaders:    true,
	})
	if err != nil {
		log.Error("WriteHeaders() error = ", err)
		return err
	}
	h2.bw.Flush()
	return nil
}

func (h2 *H2Client) heartBeat() {
	for {
		select {
		case <-h2.ctx.Done():
			log.Info("H2Client heartBeat exit")
			return
		case <-h2.reCh:
			continue
		case <-time.After(time.Second * 30):

			if err := h2.framer.WritePing(false, [8]byte{'h', 'e', 'l', 'l', 'o', 'h', '2', 's'}); err != nil {
				log.Error("heartBeat() error = ", err)
				//log.Info("H2Client reConn...")
				h2.Close()
				return // 退出
			} else {
				//log.Info("heartBeat ...")
				h2.bw.Flush()
			}
		}
	}
}

func (h2 *H2Client) readLoop() error {
	//log.Info("H2Client readLoop() start...")

	frameProc := NewFrameHandler(h2)
	for {
		fra, err := h2.framer.ReadFrame()
		h2.reCh <- struct{}{}

		if err != nil {
			log.Error("ReadFrame() error = ", err)
			h2.Close()
			return ReadFrameErr
		}
		//帧处理
		switch f := fra.(type) {
		case *http2.HeadersFrame:
			//log.Info("Receive HeadersFrame: ", f.Header().String())
			frameProc.HandleHeadersFrame(f)
			h2.bw.Flush()

		case *http2.DataFrame:
			//log.Info("Receive DataFrame: ", string(f.Data()))
			frameProc.HandleDataFrame(f)
			//h2.bw.Flush()

		case *http2.SettingsFrame:
			//log.Info("Receive SettingsFrame:", f.String())
			//for i := 0; i < f.NumSettings(); i++ {
			//	log.Debugf("setting: %s\n", f.Setting(i).String())
			//}
			if !f.Header().Flags.Has(http2.FlagSettingsAck) {
				h2.sendSettingAck()
				//log.Info("SettingsFrame is ack")
			}

		case *http2.GoAwayFrame:
			log.Info("receive GoAwayFrame:", f.String())
			h2.Close()
			return ConnCloseErr

		case *http2.PushPromiseFrame:
			//log.Info("receive PushPromiseFrame:", f.String())

		case *http2.WindowUpdateFrame:
			//log.Info("receive WindowUpdateFrame:", f.String())

		case *http2.ContinuationFrame:
			//log.Info("receive ContinuationFrame:", f.String())

		case *http2.RSTStreamFrame:
			//log.Info("receive RSTStreamFrame:", f.String())

		case *http2.PingFrame:
			//log.Info("receive PingFrame:", f.String())

		default:
			log.Info("unknown frame")
		}
	}

	return nil
}

func (h2 *H2Client) ctxDetect() {
	select {
	case <-h2.ctx.Done():
		h2.Close()
		return
	}
}

func (cc *H2Client) encodeHeaders(req *http.Request, addGzipHeader bool, trailers string, contentLength int64) ([]byte, error) {

	tmpBuf := bytes.NewBuffer(nil)
	enc := hpack.NewEncoder(tmpBuf)

	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	host, err := httpguts.PunycodeHostPort(host)
	if err != nil {
		return nil, err
	}

	var path string
	if req.Method != "CONNECT" {
		path = req.URL.RequestURI()
	}

	for k, vv := range req.Header {
		if !httpguts.ValidHeaderFieldName(k) {
			return nil, fmt.Errorf("invalid HTTP header name %q", k)
		}
		for _, v := range vv {
			if !httpguts.ValidHeaderFieldValue(v) {
				return nil, fmt.Errorf("invalid HTTP header value %q for header %q", v, k)
			}
		}
	}

	enumerateHeaders := func(f func(name, value string)) {

		f(":authority", host)
		m := req.Method
		if m == "" {
			m = http.MethodGet
		}
		f(":method", m)
		if req.Method != "CONNECT" {
			f(":path", path)
			f(":scheme", req.URL.Scheme)
		}
		if trailers != "" {
			f("trailer", trailers)
		}

		for k, vv := range req.Header {
			if strings.EqualFold(k, "host") || strings.EqualFold(k, "content-length") {
				// Host is :authority, already sent.
				// Content-Length is automatic, set below.
				continue
			} else if strings.EqualFold(k, "connection") || strings.EqualFold(k, "proxy-connection") ||
				strings.EqualFold(k, "transfer-encoding") || strings.EqualFold(k, "upgrade") ||
				strings.EqualFold(k, "keep-alive") {
				// Per 8.1.2.2 Connection-Specific Header
				// Fields, don't send connection-specific
				// fields. We have already checked if any
				// are error-worthy so just ignore the rest.
				continue
			} else if strings.EqualFold(k, "user-agent") {
				// Match Go's http1 behavior: at most one
				// User-Agent. If set to nil or empty string,
				// then omit it. Otherwise if not mentioned,
				// include the default (below).
				if len(vv) < 1 {
					continue
				}
				vv = vv[:1]
				if vv[0] == "" {
					continue
				}

			}

			for _, v := range vv {
				f(k, v)
			}
		}
		if shouldSendReqContentLength(req.Method, contentLength) {
			f("content-length", strconv.FormatInt(contentLength, 10))
		}
		if addGzipHeader {
			f("accept-encoding", "gzip")
		}
	}

	// Do a first pass over the headers counting bytes to ensure
	// we don't exceed cc.peerMaxHeaderListSize. This is done as a
	// separate pass before encoding the headers to prevent
	// modifying the hpack state.
	hlSize := uint64(0)
	enumerateHeaders(func(name, value string) {
		hf := hpack.HeaderField{Name: name, Value: value}
		hlSize += uint64(hf.Size())
	})

	trace := httptrace.ContextClientTrace(req.Context())
	traceHeaders := traceHasWroteHeaderField(trace)

	// Header list size is ok. Write the headers.
	enumerateHeaders(func(name, value string) {
		name = strings.ToLower(name)
		//cc.writeHeader(name, value)
		enc.WriteField(hpack.HeaderField{Name: name, Value: value})
		if traceHeaders {
			traceWroteHeaderField(trace, name, value)
		}
	})

	//res := make([]byte, cc.hbuf.Len())
	//copy(res, cc.hbuf.Bytes())
	//return res, nil
	return tmpBuf.Bytes(), nil
}

func (cc *H2Client) writeHeader(name, value string) {
	cc.henc.WriteField(hpack.HeaderField{Name: name, Value: value})
}

func (cc *H2Client) encoderSimpleHeaders(req *http.Request) []byte {
	tmpBuf := bytes.NewBuffer(nil)
	enc := hpack.NewEncoder(tmpBuf)

	req.Header.Set(":status", "200")

	for k, vv := range req.Header {
		v := strings.Join(vv, ",")
		log.Debugf("%s : %s", k, v)
		enc.WriteField(hpack.HeaderField{Name: k, Value: v})
	}

	return tmpBuf.Bytes()
}
