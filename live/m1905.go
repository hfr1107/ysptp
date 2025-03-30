package live

import (
	"bytes"
	"compress/gzip"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type M1905 struct {
}

var m1905M3uCache sync.Map

type m1905M3uCacheItem struct {
	liveUrl    string
	Expiration int64
}

var (
	tripleDESKey = []byte{
		105, 117, 102, 108, 101, 115, 56, 55, 56, 55, 114, 101, 119, 106, 107,
		49, 113, 107, 113, 57, 100, 106, 55, 54,
	}
	initializationVector = []byte{118, 115, 48, 108, 100, 55, 119, 51}
	hexDigits            = []byte("0123456789abcdef")
)

func (y *M1905) HandleMainRequest(c *gin.Context) {
	if liveUrl, found := getM1905M3uCache("cctv6"); found {
		LogInfo("1905 命中缓存")
		c.Redirect(302, liveUrl)
		return
	}
	liveUrl := m1905GetM3u8()
	if liveUrl == "" {
		c.String(http.StatusNotFound, "获取视频地址错误！")
		return
	}
	setM1905M3uCache("cctv6", liveUrl)
	c.Redirect(302, liveUrl)
}
func m1905GetM3u8() string {
	deviceID := UIDsData[0].UID
	req, _ := http.NewRequest("GET", "http://mapps.m1905.cn/cctv6/indexweek", nil)
	query := req.URL.Query()
	query.Add("userid", "")
	req.URL.RawQuery = query.Encode()

	req.Header = http.Header{
		"pid":             []string{"233"},
		"ver":             []string{"100/139/2016020901"},
		"Did":             []string{deviceID},
		"osv":             []string{"31"},
		"key":             []string{generateKey(deviceID)},
		"User-Agent":      []string{"Android 12; Language/zh-cn"},
		"App-Conf-Code":   []string{"13"},
		"Accept-Encoding": []string{"gzip"},
	}

	resp, err := Client.Do(req)
	if err != nil {
		LogError("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	// 检查 Content-Encoding 是否是 gzip
	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			LogError("Error creating gzip reader:", err)
			return ""
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = resp.Body
	}

	// 读取解压后的数据
	var body strings.Builder
	_, _ = io.Copy(&body, reader)

	decodedData, _ := base64.StdEncoding.DecodeString(body.String())

	decryptedResult, err := decryptCBC(decodedData)
	if err != nil {
		LogError(err)
		return ""
	}

	resultText := string(decryptedResult)
	//LogDebug("Decrypted content:", resultText)

	if liveURL := extractStreamURL(resultText); liveURL != "" {
		LogInfo("Extracted stream URL:", liveURL)
		return liveURL
	}
	LogError("解析视频地址失败")
	return ""
}

func decryptCBC(ciphertext []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(tripleDESKey)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext not multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, initializationVector)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return removePadding(plaintext)
}

func removePadding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data for padding removal")
	}
	padSize := int(data[len(data)-1])
	if padSize > len(data) {
		return nil, errors.New("invalid padding size")
	}
	return data[:len(data)-padSize], nil
}

func computeMD5Signature(input string) string {
	hash := md5.Sum([]byte(input))
	var buffer bytes.Buffer
	for _, b := range hash {
		buffer.WriteByte(hexDigits[b>>4])
		buffer.WriteByte(hexDigits[b&0x0F])
	}
	return buffer.String()
}

func generateKey(deviceID string) string {
	return computeMD5Signature(deviceID + "m1905_2014")
}

func extractStreamURL(content string) string {
	pattern := regexp.MustCompile(`"liveurl":"([^"]+)"`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}
	return strings.ReplaceAll(matches[1], "\\", "")
}

func getM1905M3uCache(key string) (string, bool) {
	// 查找缓存
	if item, found := m1905M3uCache.Load(key); found {
		cacheItem := item.(m1905M3uCacheItem)
		// 检查缓存是否过期
		if time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.liveUrl, true
		}
	}
	// 如果没有找到或缓存已过期，返回空
	return "", false
}

func setM1905M3uCache(key, liveUrl string) {
	m1905M3uCache.Store(key, m1905M3uCacheItem{
		liveUrl:    liveUrl,
		Expiration: time.Now().Unix() + 3600,
	})
}
