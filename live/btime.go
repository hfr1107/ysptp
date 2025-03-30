package live

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Btime struct {
}

var btimeM3uCache sync.Map

type btimeM3uCacheItem struct {
	streamUrl  string
	Expiration int64
}

var BtimeList = map[string]string{
	"bjws4k.m3u8":    "5755n511tbk8flo40l4c71l0sdf", //北京卫视4K超高清
	"brtv_wy.m3u8":   "54db6gi5vfj8r8q1e6r89imd64s", //BRTV文艺
	"brtv_jskj.m3u8": "53bn9rlalq08lmb8nf8iadoph0b", //BRTV纪实科教
	"brtv_ys.m3u8":   "50mqo8t4n4e8gtarqr3orj9l93v", //BRTV影视
	"brtv_cj.m3u8":   "50e335k9dq488lb7jo44olp71f5", //BRTV财经
	"brtv_tyxx.m3u8": "54hv0f3pq079d4oiil2k12dkvsc", //BRTV体育休闲
	"btrv_ish.m3u8":  "50j015rjrei9vmp3h8upblr41jf", //BRTV i生活
	"brtv_xw.m3u8":   "53gpt1ephlp86eor6ahtkg5b2hf", //BRTV新闻
	"kqsr.m3u8":      "55skfjq618b9kcq9tfjr5qllb7r", //卡酷少儿

}

func (y *Btime) HandleMainRequest(c *gin.Context, vid string) {
	LogInfo("当前vid ", vid)
	if _, ok := BtimeList[vid]; !ok {
		c.String(http.StatusNotFound, "vid not found!") // 返回 404 状态码和错误信息
		return
	}
	if streamUrl, found := getBtimeM3uCache(vid); found {
		LogInfo(vid, "命中缓存")
		c.Redirect(302, streamUrl)
		return
	}
	streamUrl := btimeGetM3u8(BtimeList[vid])
	if streamUrl == "" {
		c.String(http.StatusNotFound, "获取视频地址错误！")
		return
	}
	setBtimeM3uCache(vid, streamUrl)
	c.Redirect(302, streamUrl)
}

func btimeGetM3u8(gid string) string {
	// 目标URL
	baseURL := "https://pc.api.btime.com/video/play"
	timestamp := time.Now().Unix()
	sign := getSign(gid, 151, timestamp)
	// 构造查询参数
	params := url.Values{}
	params.Add("from", "pc")
	params.Add("id", gid)
	params.Add("type_id", "151")
	params.Add("timestamp", strconv.FormatInt(timestamp, 10))
	params.Add("sign", sign)

	// 构造完整URL
	fullURL := baseURL + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		LogError("Error creating request:", err)
		return ""
	}

	// 设置请求头
	req.Header = http.Header{
		"Accept":          []string{"*/*"},
		"Accept-Encoding": []string{"gzip, deflate"},
		"Accept-Language": []string{"en-US,en-GB;q=0.9,en;q=0.8,zh-CN;q=0.7,zh;q=0.6,ja;q=0.5"},
		"Referer":         []string{"https://www.btime.com/"},
		"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"},
	}

	// 发送请求
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

	LogInfo("Status Code:", resp.Status)
	LogDebug("Headers:", resp.Header)
	LogInfo("Body:", body.String())

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(body.String()), &result); err != nil {
		LogError("Error parsing JSON:", err)
		return ""
	}
	streamURL, ok := result["data"].(map[string]interface{})["video_stream"].([]interface{})[0].(map[string]interface{})["stream_url"].(string)
	if !ok {
		LogError("Failed to extract stream_url")
		return ""
	}
	if streamURL[:4] == "http" {
		LogInfo("streamURL:", streamURL)
		return streamURL
	}
	finalStreamURL, err := getStreamURL(streamURL)
	if err != nil {
		LogError("解析视频地址失败: ", err)
		return ""
	}
	LogInfo("streamURL:", finalStreamURL)
	return finalStreamURL
}
func getBtimeM3uCache(key string) (string, bool) {
	// 查找缓存
	if item, found := btimeM3uCache.Load(key); found {
		cacheItem := item.(btimeM3uCacheItem)
		// 检查缓存是否过期
		if time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.streamUrl, true
		}
	}
	// 如果没有找到或缓存已过期，返回空
	return "", false
}

func setBtimeM3uCache(key, streamUrl string) {
	btimeM3uCache.Store(key, btimeM3uCacheItem{
		streamUrl:  streamUrl,
		Expiration: time.Now().Unix() + 3600,
	})
}

func base64Decode(strData string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(strData)
	if err != nil {
		return "", err
	}
	decodedStr := string(decodedBytes)
	escaped := url.QueryEscape(decodedStr)
	unescaped, err := url.QueryUnescape(escaped)
	if err != nil {
		return "", err
	}
	return unescaped, nil
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func getStreamURL(encodedStr string) (string, error) {
	reversedStr := reverseString(encodedStr)
	firstDecode, err := base64Decode(reversedStr)
	if err != nil {
		return "", err
	}
	secondDecode, err := base64Decode(firstDecode)
	if err != nil {
		return "", err
	}
	return secondDecode, nil
}

func getSign(id string, typeID int, timestamp int64) string {
	signStr := fmt.Sprintf("%s%d%dTtJSg@2g*$K4PjUH", id, typeID, timestamp)
	hasher := md5.New()
	hasher.Write([]byte(signStr))
	md5Hash := hex.EncodeToString(hasher.Sum(nil))
	return md5Hash[:8]
}
