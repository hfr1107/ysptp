package live

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
)

func GetUIDStatus() (string, bool) {
	dataJson := ReadJsonFile("./data.json")
	if dataJson == (Data{}) {
		LogInfo("data.json 为空，重新生成")
		newUID := GenerateAndroidID()
		var newDataJson Data
		newDataJson.UID = newUID
		newDataJson.Init = false
		WriteJsonFile(newDataJson, "./data.json")
		UID = newUID
		return newUID, false
	} else {
		UID = dataJson.UID
		UIDInit = dataJson.Init
		LogInfo("UID 读取成功：", UID)
		LogInfo("Init 读取成功：", UIDInit)
		return UID, UIDInit
	}
}

func GetGUID() error {
	if UID == "" {
		LogInfo("UID 为空，重新获取")
		GetUIDStatus()
	}

	encrypredUID, _ := EncryptByPublicKey(UID, PubKey)
	// 构造 JSON 数据
	requestBody := map[string]string{
		"device_name": "央视频电视投屏助手",
		"device_id":   encrypredUID,
	}
	// 转换为 JSON 字符串
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		LogError("Error marshalling JSON:", err)
		return err
	}

	// 创建请求主体
	reqBody := bytes.NewBuffer([]byte(jsonData))
	url := UrlCloudwsRegister
	if UIDInit {
		url = UrlCloudwsGet
	}
	LogDebug("UrlCloudws：", url)
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		LogError("Error creating request:", err)
		return err
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("UID", UID)
	req.Header.Set("Referer", Referer)
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// 执行请求并读取响应
	//client := &http.Client{}
	resp, err := Client.Do(req)
	if err != nil {
		LogError("请求失败：", err)
	}
	defer resp.Body.Close()
	var body strings.Builder
	_, _ = io.Copy(&body, resp.Body)
	LogDebug("UrlCloudws结果：", body.String())
	// 解析 JSON 响应
	var result map[string]interface{}
	e2 := json.Unmarshal([]byte(body.String()), &result)
	if e2 != nil {
		return e2
	}
	if result["result"] == 0.0 {
		data := result["data"].(map[string]interface{})
		GUID = data["guid"].(string)
		if !UIDInit {
			dataJson := ReadJsonFile("./data.json")
			dataJson.Init = true
			WriteJsonFile(dataJson, "./data.json")
		}
	} else if result["result"] == 604.0 {
		dataJson := ReadJsonFile("./data.json")
		dataJson.Init = true
		WriteJsonFile(dataJson, "./data.json")
		GetGUID()
	} else if result["result"] == 605.0 {
		dataJson := ReadJsonFile("./data.json")
		dataJson.Init = false
		WriteJsonFile(dataJson, "./data.json")
		GetGUID()
	} else {
		LogError("GetGUID 未知错误：", result["result"])
	}

	return nil

}

func CheckPlayAuth() bool {
	// 构造 JSON 数据
	requestBody := map[string]string{
		"guid": GUID,
	}
	// 转换为 JSON 字符串
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		LogError("Error marshalling JSON:", err)
	}

	// 创建请求主体
	reqBody := bytes.NewBuffer([]byte(jsonData))
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", UrlCheckPlayAuth, reqBody)
	if err != nil {
		LogError("Error creating request:", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("UID", UID)
	req.Header.Set("Referer", Referer)
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// 执行请求并读取响应
	//client := &http.Client{}
	resp, err := Client.Do(req)
	if err != nil {
		LogError("请求失败：", err)
		return false
	}
	defer resp.Body.Close()
	var body strings.Builder
	_, _ = io.Copy(&body, resp.Body)
	LogDebug("CheckPlayAuth结果：", body.String())
	// 解析 JSON 响应
	var result map[string]interface{}
	e2 := json.Unmarshal([]byte(body.String()), &result)
	if e2 != nil {
		return false
	}
	if result["message"].(string) == "SUCCESS" {
		LogInfo("播放授权成功")
		return true
	} else {
		return false
	}
}

func GetBaseM3uUrl(liveID string) string {
	// 使用 crypto/rand 生成一个范围内的随机数
	max := big.NewInt(int64(len(DeviceModel))) // 设置最大范围为 len(DeviceModele)
	randomIndex, _ := rand.Int(rand.Reader, max)
	// 构造 JSON 数据
	requestBody := map[string]interface{}{
		"rate":       "",
		"systemType": "android",
		"model":      DeviceModel[randomIndex.Int64()],
		"id":         liveID,
		"userId":     "",
		"clientSign": "cctvVideo",
		"deviceId": map[string]string{
			"serial":     "",
			"imei":       "",
			"android_id": UID,
		},
	}

	// 将结构体序列化为 JSON
	jsonData, err := json.MarshalIndent(requestBody, "", "  ")
	if err != nil {
		LogError("Error marshaling JSON:", err)
		return ""
	}
	// 创建请求主体
	reqBody := bytes.NewBuffer([]byte(jsonData))
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", UrlGetBaseM3u8, reqBody)
	if err != nil {
		LogError("Error creating request:", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("UID", UID)
	req.Header.Set("Referer", Referer)
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// 执行请求并读取响应
	//client := &http.Client{}
	resp, err := Client.Do(req)
	if err != nil {
		LogError("请求失败：", err)
		return ""
	}
	defer resp.Body.Close()
	var body strings.Builder
	_, _ = io.Copy(&body, resp.Body)
	LogDebug("GetBaseM3uUrl结果：", body.String())
	// 解析 JSON 响应
	var result map[string]interface{}
	e2 := json.Unmarshal([]byte(body.String()), &result)
	if e2 != nil {
		return ""
	}
	if result["message"].(string) != "SUCCESS" {
		LogError("GetBaseM3uUrl 未知错误：", result["message"])
		return ""
	}
	data := result["data"].(map[string]interface{})
	videoList := data["videoList"].([]interface{})

	// 获取 videoList[0] 的 url
	if len(videoList) > 0 {
		video := videoList[0].(map[string]interface{})
		url := video["url"].(string)
		LogDebug("Video URL:", url)
		return url
	} else {
		LogError("No videos available.")
		return ""
	}
}

func GetAppSecret() bool {
	if GUID == "" {
		LogInfo("GUID 为空，重新获取")
		GetGUID()
	}
	encryptedGUID, _ := EncryptByPublicKey(GUID, PubKey)
	// 构造 JSON 数据
	requestBody := map[string]string{
		"guid": encryptedGUID,
	}
	// 转换为 JSON 字符串
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		LogError("Error marshalling JSON:", err)
	}

	// 创建请求主体
	reqBody := bytes.NewBuffer([]byte(jsonData))
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", UrlGetAppSecret, reqBody)
	if err != nil {
		LogError("Error creating request:", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("UID", UID)
	req.Header.Set("Referer", Referer)
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// 执行请求并读取响应
	//client := &http.Client{}
	resp, err := Client.Do(req)
	if err != nil {
		LogError("请求失败：", err)
		return false
	}
	defer resp.Body.Close()
	var body strings.Builder
	_, _ = io.Copy(&body, resp.Body)
	LogDebug("GetAppSecret结果：", body.String())
	// 解析 JSON 响应
	var result map[string]interface{}
	e2 := json.Unmarshal([]byte(body.String()), &result)
	if e2 != nil {
		return false
	}
	if result["message"].(string) == "SUCCESS" {
		data := result["data"].(map[string]interface{})
		decryptedAppSecret, e := DecryptByPublicKey(data["appSecret"].(string), PubKey)
		if e != nil {
			LogError("AppSecret解密失败：", e)
			return false
		}
		LogDebug("AppSecret：" + decryptedAppSecret)
		AppSecret = decryptedAppSecret
		return true
	}
	return false
}
