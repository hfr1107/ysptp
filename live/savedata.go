package live

import (
	"encoding/json"
	"os"
)

type Data struct {
	UID  string `json:"UID"`
	Init bool   `json:"Init"`
}

func ReadJsonFile(file string) Data {
	// 读取配置文件
	data, err := os.ReadFile(file)
	if err != nil {
		//log.Fatal("Error reading config file:", err)
		LogError("Error reading config file:", err)
		return Data{}
	}

	// 创建一个配置对象
	var dataCache Data

	// 将 JSON 数据解析到结构体中
	err = json.Unmarshal(data, &dataCache)
	if err != nil {
		//log.Fatal("Error unmarshalling config data:", err)
		LogError("Error unmarshalling config data:", err)
		return Data{}
	}
	return dataCache
}
func WriteJsonFile(data Data, file string) {
	// 将配置对象转换为 JSON 字符串
	jsonData, err := json.Marshal(data)
	if err != nil {
		LogError("Error marshalling config data:", err)
	}

	// 写入配置文件
	err = os.WriteFile(file, jsonData, 0644)
	if err != nil {
		LogError("Error writing config file:", err)
	}
}
