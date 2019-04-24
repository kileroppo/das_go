package upgrade

import (
	"../core/httpgo"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"../core/log"
	"strings"
	"net/http"
	"os"
	"io"
	"fmt"
	"path"
	"net/url"
	"time"
	"../core/constant"
	"bytes"
	"encoding/hex"
	"../core/util"
	"../core/entity"
)

type Msg struct {
	Pus PusMsg				`json:"PUS"`
}

type PusMsg struct {
	Head Header				`json:"header"`
	Body PusDowloadBody		`json:"body"`
}

type Header struct {
	Api_version string		`json:"api_version"`
	Return_string string	`json:"return_string"`
	Seq_id string 			`json:"seq_id"`
	Http_code string 		`json:"http_code"`
	Message_type string 	`json:"message_type"`
}

type PusDowloadBody struct {
	New_version string		`json:"new_version"`
	Endpoint_type string	`json:"endpoint_type"`
	Force_upgrade string 	`json:"force_upgrade"`
	Md5 string				`json:"md5"`
	Vendor_name string		`json:"vendor_name"`
	Url string 				`json:"url"`
	Platform string			`json:"platform"`
	Part string				`json:"part"`
	Readme string 			`json:"readme"`
}

type QueryPkgInfo struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	SeqId int			`json:"seqId"`

	Part int			`json:"part"`
	FileName string		`json:"fileName"`
	FileSize int64		`json:"fileSize"`
	MD5 string			`json:"MD5"`
}

type TransferPkgData struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	SeqId int			`json:"seqId"`

	Part int			`json:"part"`
	Offset int64		`json:"offset"`
	FileData string		`json:"fileData"`
}

func GetUpgradeFileInfo(devId string, devType string, seqId int, partId int) {
	body1, err1:= httpgo.Http2WonlyUpgrade(devType)
	if nil != err1 {
		log.Error("get upgrade file from wonly pus failed, err: ", err1)
		return
	}
	log.Debug("body:", string(body1), ", error: ", err1)

	var data Msg
	if err := json.Unmarshal(body1, &data); err != nil {
		log.Error("Msg json.Unmarshal, err=", err)
		return
	}
	if "200" != data.Pus.Head.Http_code {
		log.Error("get upgrade file failed, http_code: ", data.Pus.Head.Http_code)
		return
	}

	part := strings.Split(data.Pus.Body.Part, ",")
	log.Debug("len:", len(part), ", part[0]:", part[0], ", part[1]:", part[1])

	md5 := strings.Split(data.Pus.Body.Md5, ",")
	log.Debug("len:", len(md5), ", md5[0]:", md5[0], ", md5[1]:", md5[1])

	fileUrl := strings.Split(data.Pus.Body.Url, ",")
	log.Debug("len:", len(fileUrl), ", url[0]:", fileUrl[0], ", url[1]:", fileUrl[1])

	var pkgUrl string
	var pkgMd5 string
	pkgMd5 = md5[partId]
	pkgUrl = fileUrl[partId]

	for i, _ := range part {
		if partId == i {
			pkgMd5 = md5[i]
			pkgUrl = fileUrl[i]
			break
		}
	}

	log.Debug("pkgMd5:", pkgMd5, ", pkgUrl:", pkgUrl)
	if "" == pkgUrl && "" == pkgMd5 {
		log.Error("GetUpgradeFileInfo pkgUrl pkgMd5 is null, check the partId, ", partId)
		return
	}

	fileName, fileSize, err := Download(pkgUrl)
	if nil != err {
		log.Error("download file err: ", err)
		return
	}
	var fileInfo QueryPkgInfo
	fileInfo.Cmd = constant.Get_Upgrade_FileInfo
	fileInfo.Ack = 1
	fileInfo.DevType = devType
	fileInfo.DevId = devId
	fileInfo.SeqId = seqId
	fileInfo.Part = partId
	fileInfo.FileName = fileName
	fileInfo.FileSize = fileSize
	fileInfo.MD5 = pkgMd5
	if toDevice_fileInfo, err := json.Marshal(fileInfo); err == nil {
		log.Info("constant.Get_Upgrade_FileInfo, resp to device, ", string(toDevice_fileInfo))

		// myKey := util.MD52Bytes(fileInfo.DevId)

		var strToDevData string
		var toDevHead entity.MyHeader
		toDevHead.ApiVersion = constant.API_VERSION
		toDevHead.ServiceType = constant.SERVICE_TYPE_UPGRADE_NOPASS

		/*if strToDevData, err = util.ECBEncrypt(toDevice_fileInfo, myKey); err == nil {
			toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
			toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, toDevHead)
			strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
		}*/

		toDevHead.CheckSum = util.CheckSum(toDevice_fileInfo)
		toDevHead.MsgLen =  (uint16)(strings.Count(string(toDevice_fileInfo),"") - 1)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, toDevHead)
		strToDevData = hex.EncodeToString(buf.Bytes()) + string(toDevice_fileInfo)

		httpgo.Http2OneNET_write(devId, strToDevData)
	} else {
		log.Error("toDevice_str json.Marshal, err=", err)
	}
}

// 实现单个文件的下载
func Download(fileUrl string) (fileName string, fileSize int64, err error) {
	log.Debug("To download %s\n", fileUrl)

	uri, err0 := url.ParseRequestURI(fileUrl)
	if err0 != nil {
		log.Error("fileUrl is error, err:", err0.Error())
		return "", 0, err0
	}
	filename := path.Base(uri.Path)

	exist, err1 := PathExists("logs/")
	if err1 != nil {
		log.Error("get dir error![%v]\n", err1)
		return "", 0, err1
	}
	if exist {
		log.Debug("has dir: ", "logs/")
	} else {
		log.Debug("no dir: ", "logs/")
		// 创建文件夹
		err2 := os.Mkdir("logs/", os.ModePerm)
		if err != nil {
			log.Error("mkdir failed, ", err2)
			return "", 0, err2
		} else {
			log.Debug("mkdir success!")
		}
	}

	fpath := fmt.Sprintf("logs/%s", filename)
	newFile, err3 := os.Create(fpath)
	if err3 != nil {
		log.Error("process failed for ", filename, ", err:", err3.Error())
		return "", 0, err3
	}

	defer newFile.Close()

	client := http.Client{ Timeout: 900 * time.Second }
	resp, err4 := client.Get(fileUrl)
	if err4 != nil {
		// panic(err4)
		log.Error("http get upgrade file failed, ", filename, ", err:", err4.Error())
		return "", 0, err4
	}
	log.Debug("fileName", filename, ", fileSize:", resp.ContentLength)

	defer resp.Body.Close()

	_, err5 := io.Copy(newFile, resp.Body)
	if err5 != nil {
		log.Error("save file error, err:", err5.Error())
		return "", 0, err5
	}

	return filename, resp.ContentLength, nil
}

func TransferFileData(devId string, devType string, seqId int, offset int64, fileName string, partId int) {
	log.Debug("TransferFileData %s to device.", fileName)
	fpath := fmt.Sprintf("logs/%s", fileName)
	file, err0 := os.OpenFile(fpath, os.O_RDONLY, os.ModePerm)
	if err0 != nil {
		defer file.Close()
		log.Error("os.OpenFile failed, err=", err0)
		return
	}

	file.Seek(constant.FILE_MAX*offset, 0)

	var buffer bytes.Buffer
	io.CopyN(&buffer, file, constant.FILE_MAX)
	_bytes := buffer.Bytes()

	// encodedStr := hex.EncodeToString(_bytes)
	encodedStr := base64.StdEncoding.EncodeToString(_bytes)

	var fileData TransferPkgData
	fileData.Cmd = constant.Download_Upgrade_File
	fileData.Ack = 1
	fileData.DevType = devType
	fileData.DevId = devId
	fileData.SeqId = seqId
	fileData.Part = partId
	fileData.Offset = offset
	fileData.FileData = encodedStr
	if toDevice_fileData, err := json.Marshal(fileData); err == nil {
		log.Info("constant.Download_Upgrade_File, resp to device, ", string(toDevice_fileData))
		httpgo.Http2OneNET_write(devId, string(toDevice_fileData))

		// myKey := util.MD52Bytes(fileData.DevId)

		var strToDevData string
		var toDevHead entity.MyHeader
		toDevHead.ApiVersion = constant.API_VERSION
		toDevHead.ServiceType = constant.SERVICE_TYPE_UPGRADE_NOPASS

		/*if strToDevData, err = util.ECBEncrypt(toDevice_fileData, myKey); err == nil {
			toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
			toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, toDevHead)
			strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
		}*/

		toDevHead.CheckSum = util.CheckSum(toDevice_fileData)
		toDevHead.MsgLen =  (uint16)(strings.Count(string(toDevice_fileData),"") - 1)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, toDevHead)
		strToDevData = hex.EncodeToString(buf.Bytes()) + string(toDevice_fileData)

		httpgo.Http2OneNET_write(devId, strToDevData)

	} else {
		log.Error("toDevice_fileData json.Marshal, err=", err)
	}
}

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
