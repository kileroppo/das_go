package h2client

import (
	"net/http"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"

	"../log"
	"../entity"
)

var (
	AliDataCh = make(chan entity.AliRawData, 1000)
)

type FrameHandler interface {
	HandleDataFrame(f *http2.DataFrame)
	HandleHeadersFrame(f *http2.HeadersFrame)
}

func NewFrameHandler(cli *H2Client) FrameHandler {
	return AliDataHandle{
		cli:     cli,
		dataMap: make(map[uint32]aliFrameData),
	}
}

type aliFrameData struct {
	streamId uint32
	msgId    string
	qos      string
	topic    string
	rawData []byte
}

type AliDataHandle struct {
	cli     *H2Client
	dataMap map[uint32]aliFrameData
}

func (a AliDataHandle) HandleHeadersFrame(f *http2.HeadersFrame) {
	//dec := hpack.NewDecoder(4086,nil)

	headers, err := a.cli.hdec.DecodeFull(f.HeaderBlockFragment())
	if err != nil {
		log.Error("HeadersFrame DecodeFull() error = ", err)
		return
	}

	streamId := f.StreamID
	var xQos, msgId, topic string
	for _, h := range headers {
		log.Debugf("Header -> %s : %s\n", h.Name, h.Value)
		if strings.EqualFold("x-qos", h.Name) {
			xQos = h.Value
		} else if strings.EqualFold("x-message-id", h.Name) {
			msgId = h.Value
		} else if strings.EqualFold("x-topic", h.Name) {
			topic = h.Value
		}
	}
	data := aliFrameData{
		msgId:    msgId,
		qos:      xQos,
		streamId: streamId,
		topic:    topic,
	}

	a.dataMap[data.streamId] = data
}

func (a AliDataHandle) HandleDataFrame(f *http2.DataFrame) {
	log.Debug("HandleDataFrame() start...")
	data, ok := a.dataMap[f.StreamID]

	msg := make([]byte, len(f.Data()))
	copy(msg, f.Data())

	//log.Debug("receive data: ", string(msg))

	if ok {
		if f.StreamEnded() {
			if data.qos == "1" || data.qos == "2" {
				a.writeAck(data)
			}
			delete(a.dataMap, f.StreamID)
			if len(data.topic) > 0 {
				AliDataCh <- entity.AliRawData{
					RawData:msg,
					Topic:data.topic,
				}
			}

		} else {
			data.rawData = append(data.rawData, msg...)
			a.dataMap[f.StreamID] = data
		}
	} else {
		log.Warningf("DataFrame %d did not match any HeadersFrame", f.StreamID)
	}

	//log.Debug("dataMap len: ", len(a.dataMap))
}

func (a AliDataHandle) writeAck(data aliFrameData) {
	//urlData,_ := url.Parse(a.cli.rawurl)
	request, _ := http.NewRequest("GET", a.cli.rawurl, nil)
	request.Header.Set("x-message-id", data.msgId)
	//request.Header.Set("x-sdk-version", "1.1.4")
	//request.Header.Set("x-sdk-version-name", "v1.1.4")
	//request.Header.Set("x-sdk-platform", "java")

	var err error
	log.Infof("Send Ack for streamId %d", data.streamId)
	blockFragment, err := a.cli.encodeHeaders(request, false, "", actualContentLength(request))
	if err != nil {
		log.Warning("writeAck encodeHeaders() error = ", err)
		a.cli.bw.Flush()
		return
	}
	//blockFragment := a.cli.encoderSimpleHeaders(request)

	bf := make([]byte, len(blockFragment))
	copy(bf, blockFragment)

	a.readHeader(bf)

	a.cli.bw.Flush()
	err = a.cli.framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      data.streamId+1,
		BlockFragment: blockFragment,
		EndStream:     true,
		EndHeaders:    true,
	})

	if err != nil {
		log.Error("WriteHeaders() error = ", err)
	} else {
		log.Info("Send Ack Success")
	}
	a.cli.bw.Flush()
}

func (a AliDataHandle) readHeader(bf []byte) {

	dec := hpack.NewDecoder(4096, nil)
	hds, err := dec.DecodeFull(bf)
	if err != nil {
		log.Error("decoder.DecodeFull() error = ", err)
		return
	}

	for _, h := range hds {
		log.Debugf("ACK -> %s:%s", h.Name, h.Value)
	}
}
