package live

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Ysptp struct {
}

var basem3uCache sync.Map
var m3uCache sync.Map

type BaseM3uCacheItem struct {
	baseUrl    string
	Expiration int64
}

type M3uCacheItem struct {
	uid          string
	playUrl      string
	Expiration   int64
	appRandomStr string
	appSign      string
	urlPath      string
}

func (y *Ysptp) HandleMainRequest(c *gin.Context, vid string) {

	// 检查全局变量 cctvList 中是否包含指定的 ID
	// 如果找不到对应的 ID，返回 404 错误并终止函数
	if _, ok := CCTVList[vid]; !ok {
		c.String(http.StatusNotFound, "vid not found!") // 返回 404 状态码和错误信息
		return
	}

	// 调用自定义函数 getURL，根据 ID、基础 URL、用户 UID 和 path 获取视频数据
	header, data, urlPath := getURL(vid, CCTVList[vid])
	for key, values := range header {
		// 删除Gin自动设置的Content-Type
		if key == "Content-Type" {
			c.Writer.Header().Del(key)
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Writer.Header().Del("Content-Length")

	// 构建当前请求的主机和路径信息，作为 URL 前缀
	// 示例: 如果请求地址为 http://example.com/path，则 golang = "http://example.com/path"
	golang := "http://" + c.Request.Host + c.Request.URL.Path

	// 使用正则表达式匹配视频数据中的 TS 文件链接
	// `(?i)` 表示忽略大小写，匹配 .ts 文件的路径
	re := regexp.MustCompile(`((?i).*?\.ts)`)

	// 将匹配到的 TS 文件路径替换为新的路径，格式为:
	// 当前请求地址 + "?ts=" + 附加参数 + TS 文件路径
	data = re.ReplaceAllString(data, golang+"?ts="+urlPath+"$1")
	//fmt.Println("data", data)

	// 设置 HTTP 响应头，用于指定文件下载的名称
	// c.Header("Content-Disposition", "attachment;filename="+vid)

	// 返回 HTTP 响应状态码 200 和处理后的视频数据
	c.Data(http.StatusOK, header.Get("Content-Type"), []byte(data))
}

// 处理 TS 请求，返回 TS 视频流数据
func (y *Ysptp) HandleTsRequest(c *gin.Context, ts, vid string, wsTime string, wsSecret string) {

	// 构建请求数据
	wT := ""
	wS := ""
	if wsTime != "" {
		wT = "&wsTime=" + wsTime
	}
	if wsSecret != "" {
		wS = "&wsSecret=" + wsSecret
	}
	data := ts + wT + wS
	uid, _, appSign, appRandomStr, _, found := GetCache(vid)
	if !found {
		LogError("未知的ts", ts)
		c.String(http.StatusNotFound, "ts not found!")
	}

	// 设置响应头为视频流类型
	c.Header("Content-Type", "video/MP2T")
	// 返回视频数据
	c.String(http.StatusOK, getTs(data, uid, appSign, appRandomStr))
}

// 获取视频 URL，若缓存中存在则直接返回，否则发起请求获取
func getURL(vid, liveID string) (http.Header, string, string) {

	// 查找缓存
	if uid, playURL, appSign, appRandomStr, urlPath, found := GetCache(vid); found {
		// 如果缓存中有，返回缓存中的数据
		//fmt.Println("命中缓存", cacheKey)
		header, data := fetchData(playURL, uid, appSign, appRandomStr, urlPath)
		return header, data, urlPath
	}

	// 初始化随机数种子（使用纳秒时间戳确保每次运行结果不同）
	rand.Seed(time.Now().UnixNano())

	// 生成随机整数
	uidIndex := rand.Intn(UIDCount)
	uid := UIDsData[uidIndex].UID
	LogDebug(vid, " 使用UID ", uidIndex)

	baseM3u8Url, found := GetBaseM3uCache(vid)
	if !found {
		baseM3u8Url = GetBaseM3uUrl(liveID, uidIndex)
		if baseM3u8Url == "" {
			LogError("获取base m3u8地址失败")
			//return "", ""
			//panic("获取base m3u8地址失败")
			return nil, "", ""
		}
		SetBaseM3uCache(vid, baseM3u8Url)
	}

	// POST 数据
	postData := map[string]string{
		"appcommon": `{"adid":"` + uid + `","av":"` + AppVersion + `","an":"央视视频电视投屏助手","ap":"cctv_app_tv"}`,
		"url":       baseM3u8Url,
	}
	// postData := map[string]string{
	// 	"appcommon": `{"adid":"123456","av":"1.3.4","an":"央视视频电视投屏助手","ap":"cctv_app_tv"}`,
	// 	"url":       "http://live-tpgq.cctv.cn/live/3e1b6788736d5a9507c7f9f627ff04f8.m3u8",
	// }

	retry := 2
REPEAT:
	appRandomStr := uuid.New().String()
	appSignStr := AppId + AppSecret + appRandomStr
	appSign := Md5Encrypt(appSignStr)

	// 创建 POST 请求
	req, _ := http.NewRequest("POST", UrlGetStream, strings.NewReader(EncodeFormData(postData)))
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Referer", Referer)
	req.Header.Set("UID", uid)
	req.Header.Set("APPID", AppId)
	req.Header.Set("APPSIGN", appSign)
	req.Header.Set("APPRANDOMSTR", appRandomStr)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 执行请求并读取响应
	//client := &http.Client{}
	resp, err := Client.Do(req)
	if err != nil {
		LogError("请求失败：", err)
		//return "", ""
		//panic(err)
		return nil, "", ""
	}
	defer resp.Body.Close()
	var body strings.Builder
	_, _ = io.Copy(&body, resp.Body)

	LogDebug("getstream结果：", body.String())

	// 解析 JSON 响应
	var result map[string]interface{}
	e := json.Unmarshal([]byte(body.String()), &result)
	if e != nil {
		LogError(e)
		return nil, "", ""
	}

	if result["succeed"].(float64) != 1.0 {
		LogError("getstream 错误")
		if retry > 0 {
			GetAppSecret()
			LogError("剩余", retry, "次重试")
			uidIndex = (uidIndex + 1) % UIDCount
			retry = retry - 1
			goto REPEAT
		}
		return nil, "", ""
	}
	playURL := result["url"].(string)
	urlPath := ExtractUrlPath(playURL)
	LogDebug("playURL:", playURL)

	// 将结果缓存起来
	SetCache(vid, uid, playURL, appRandomStr, appSign, urlPath)

	header, data := fetchData(playURL, uid, appSign, appRandomStr, urlPath)
	// 返回获取的数据
	return header, data, urlPath
}

// 从指定的播放 URL 获取数据
func fetchData(playURL string, uid string, appSign string, appRandomStr, urlPath string) (http.Header, string) {

	// 创建一个 HTTP 客户端，用于发起请求
	// client := &http.Client{
	// 	Transport: &http.Transport{
	// 		ForceAttemptHTTP2: false,
	// 		//TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // 测试时可跳过证书验证
	// 	},
	// 	Timeout: 3 * time.Second,
	// }

	// 无限循环，直到函数返回数据为止
	for {
		// 构造一个 HTTP GET 请求
		// 使用传入的 playURL 作为请求地址，nil 表示不需要请求体
		req, _ := http.NewRequest("GET", playURL, nil)

		// 设置请求头字段，模拟请求来源
		req.Header["Accept"] = []string{"*/*"}
		//req.Header["Range"] = []string{"bytes=0-"}
		//req.Header["Connection"] = []string{"close"}
		req.Header["User-Agent"] = []string{UA}
		req.Header["Referer"] = []string{Referer}
		req.Header["UID"] = []string{uid}
		req.Header["APPID"] = []string{AppId}
		req.Header["APPSIGN"] = []string{appSign}
		req.Header["APPRANDOMSTR"] = []string{appRandomStr}
		req.Header["Icy-MetaData"] = []string{"1"}

		// 执行请求并获取响应
		resp, err := Client.Do(req)
		if err != nil {
			LogError("请求失败：", err)
			//return ""
			//panic(err)
			return nil, ""
		}
		// 确保响应体在函数返回之前被正确关闭以释放资源
		defer resp.Body.Close()

		// 使用 strings.Builder 构建响应数据
		var body strings.Builder
		// 将响应体的内容复制到 body 中（忽略错误处理）
		_, _ = io.Copy(&body, resp.Body)

		// 将响应数据转换为字符串
		data := body.String()
		//fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
		//fmt.Printf("响应头: %+v\n", resp.Header)
		//fmt.Printf("响应内容长度: %d bytes\n", len(body))

		//LogDebug("fetchData结果：", body.String())

		// 使用正则表达式匹配返回数据中的 m3u8 播放链接
		re := regexp.MustCompile(`(.*\.m3u8\?.*)`) // 匹配带有 `.m3u8` 文件及其查询参数的字符串
		matches := re.FindStringSubmatch(data)     // 查找匹配的结果

		// 如果匹配到 m3u8 文件链接
		if len(matches) > 0 {
			// 将 playURL 更新为拼接了 path 和匹配到的链接的新 URL
			playURL = urlPath + matches[0]
		} else {
			// 如果没有匹配到 m3u8 文件链接，直接返回响应数据
			return resp.Header, data
		}
	}
}

// 获取 TS 视频流数据
func getTs(url string, uid string, appSign string, appRandomStr string) string {
	// 创建 GET 请求
	req, _ := http.NewRequest("GET", url, nil)
	req.Header["Accept"] = []string{"*/*"}
	req.Header["Connection"] = []string{"keep-alive"}
	req.Header["User-Agent"] = []string{UA}
	req.Header["Referer"] = []string{Referer}
	req.Header["UID"] = []string{uid}
	req.Header["APPID"] = []string{AppId}
	req.Header["APPSIGN"] = []string{appSign}
	req.Header["APPRANDOMSTR"] = []string{appRandomStr}
	req.Header["Icy-MetaData"] = []string{"1"}

	// 执行请求并读取响应
	// client := &http.Client{
	// 	Transport: &http.Transport{
	// 		ForceAttemptHTTP2: false,
	// 		//TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // 测试时可跳过证书验证
	// 	},
	// 	Timeout: 3 * time.Second,
	// }
	resp, err := Client.Do(req)
	if err != nil {
		LogError("请求失败：", err)
		return ""
	}
	defer resp.Body.Close()
	var body strings.Builder
	_, _ = io.Copy(&body, resp.Body)

	// 返回响应内容
	return body.String()
}

// 从缓存中获取数据
func GetCache(key string) (string, string, string, string, string, bool) {
	// 查找缓存
	if item, found := m3uCache.Load(key); found {
		cacheItem := item.(M3uCacheItem)
		// 检查缓存是否过期
		if time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.uid, cacheItem.playUrl, cacheItem.appSign, cacheItem.appRandomStr, cacheItem.urlPath, true
		}
	}
	// 如果没有找到或缓存已过期，返回空
	return "", "", "", "", "", false
}

// 设置缓存数据
func SetCache(key, uid, playUrl, appRandomStr, appSign, urlPath string) {
	m3uCache.Store(key, M3uCacheItem{
		uid:          uid,
		playUrl:      playUrl,
		Expiration:   time.Now().Unix() + 2000,
		appRandomStr: appRandomStr,
		appSign:      appSign,
		urlPath:      urlPath,
	})
}

// 从缓存中获取数据
func GetBaseM3uCache(key string) (string, bool) {
	// 查找缓存
	if item, found := basem3uCache.Load(key); found {
		cacheItem := item.(BaseM3uCacheItem)
		// 检查缓存是否过期
		if time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.baseUrl, true
		}
	}
	// 如果没有找到或缓存已过期，返回空
	return "", false
}

func SetBaseM3uCache(key, baseM3uUrl string) {
	basem3uCache.Store(key, BaseM3uCacheItem{
		baseUrl:    baseM3uUrl,
		Expiration: time.Now().Unix() + 43200,
	})
}

func RefreshM3u8Cache() {

	GetAppSecret()

	if !EnableCache {
		return
	}

	rand.Seed(time.Now().UnixNano())
	uidIndex := rand.Intn(UIDCount)

	var uidCount int
	var uidsData []UIDData

	uidCount = UIDCount
	uidsData = UIDsData

	LogInfo("刷新缓存中...")
	for vid, liveID := range CCTVList {
		retry := 2
	REPEAT:
		LogDebug(vid, " 使用UID ", uidsData[uidIndex].UID)
		baseM3u8Url, found := GetBaseM3uCache(vid)
		if !found {
			baseM3u8Url = GetBaseM3uUrl(liveID, uidIndex)
			if baseM3u8Url == "" {
				LogError("获取base m3u8地址失败")
				// if uidCount > 1 {
				// 	uidCount = uidCount - 1
				// 	uidsData = append(uidsData[:uidIndex], uidsData[uidIndex+1:]...)
				// 	uidIndex = uidIndex % uidCount
				// 	time.Sleep(10 * time.Second)
				// 	goto REPEAT
				// } else {
				// 	break
				// }
				continue

			}
			SetBaseM3uCache(vid, baseM3u8Url)
		}

		// POST 数据
		postData := map[string]string{
			"appcommon": `{"adid":"` + uidsData[uidIndex].UID + `","av":"` + AppVersion + `","an":"央视视频电视投屏助手","ap":"cctv_app_tv"}`,
			"url":       baseM3u8Url,
		}
		// postData := map[string]string{
		// 	"appcommon": `{"adid":"123456","av":"1.3.4","an":"央视视频电视投屏助手","ap":"cctv_app_tv"}`,
		// 	"url":       "http://live-tpgq.cctv.cn/live/3e1b6788736d5a9507c7f9f627ff04f8.m3u8",
		// }

		appRandomStr := uuid.New().String()
		appSignStr := AppId + AppSecret + appRandomStr
		appSign := Md5Encrypt(appSignStr)

		// 创建 POST 请求
		req, _ := http.NewRequest("POST", UrlGetStream, strings.NewReader(EncodeFormData(postData)))
		req.Header.Set("User-Agent", UA)
		req.Header.Set("Referer", Referer)
		req.Header.Set("UID", uidsData[uidIndex].UID)
		req.Header.Set("APPID", AppId)
		req.Header.Set("APPSIGN", appSign)
		req.Header.Set("APPRANDOMSTR", appRandomStr)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// 执行请求并读取响应
		//client := &http.Client{}
		resp, err := Client.Do(req)
		if err != nil {
			LogError("请求失败：", err)
			continue
			//return "", ""
			//panic(err)
		}
		defer resp.Body.Close()
		var body strings.Builder
		_, _ = io.Copy(&body, resp.Body)

		LogDebug("getstream结果：", body.String())

		// 解析 JSON 响应
		var result map[string]interface{}
		e := json.Unmarshal([]byte(body.String()), &result)
		if e != nil {
			LogError("getstream 解析错误")
			continue
		}
		succeed := result["succeed"].(float64)
		if succeed != 1.0 {
			LogError("getstream 错误")
			// if uidCount > 1 {
			// 	uidCount = uidCount - 1
			// 	uidsData = append(uidsData[:uidIndex], uidsData[uidIndex+1:]...)
			// 	uidIndex = uidIndex % uidCount
			// 	time.Sleep(10 * time.Second)
			// 	goto REPEAT
			// }
			// break
			if retry > 0 {
				LogError("剩余", retry, "次重试")
				GetAppSecret() //AppSecret每小时0分刷新
				retry = retry - 1
				uidIndex = (uidIndex + 1) % uidCount
				time.Sleep(10 * time.Second)
				goto REPEAT
			}
			continue
		}

		playURL := result["url"].(string)
		urlPath := ExtractUrlPath(playURL)
		LogDebug("playURL:", playURL)

		// 将结果缓存起来
		SetCache(vid, uidsData[uidIndex].UID, playURL, appRandomStr, appSign, urlPath)
		uidIndex = (uidIndex + 1) % uidCount
		time.Sleep(10 * time.Second)
	}
	LogInfo("缓存刷新完成")

}
