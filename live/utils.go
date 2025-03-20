package live

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
)

// 加密函数
func EncryptByPublicKey(data string, pubKeyStr string) (string, error) {
	// 解码 Base64 编码的公钥
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %v", err)
	}
	// 解析公钥为 PKIX 格式
	pubKey, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}
	// 类型断言，将公钥转换为 rsa.PublicKey 类型
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not of type *rsa.PublicKey")
	}
	// 使用公钥加密数据
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, []byte(data))
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}
	// 将加密后的字节数据进行 Base64 编码
	encryptedStr := base64.StdEncoding.EncodeToString(ciphertext)
	//encryptedStr := url.QueryEscape(string(ciphertext))
	return encryptedStr, nil
}
func rsaPublicDecrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}
func DecryptByPublicKey(encryptedStr string, pubKeyStr string) (string, error) {
	// 解码 Base64 编码的公钥
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %v", err)
	}
	// 解析公钥为 PKIX 格式
	pubKey, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}
	// 类型断言，将公钥转换为 rsa.PublicKey 类型
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not of type *rsa.PublicKey")
	}
	// 使用公钥解密数据
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}
	decryptedData := []byte{}
	i := 0
	len := len(ciphertext)
	for {
		j := len - i
		if j <= 0 {
			break
		}
		if j > 128 {
			decryptedData = append(decryptedData, rsaPublicDecrypt(rsaPubKey, ciphertext[i:i+128])...)
		} else {
			decryptedData = append(decryptedData, rsaPublicDecrypt(rsaPubKey, ciphertext[i:])...)
			break
		}
		i += 128
	}
	return string(decryptedData), nil
}

// Md5Encrypt 生成字符串的 MD5 值
func Md5Encrypt(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

// EncodeFormData 将 map[string]string 编码为 application/x-www-form-urlencoded 格式
func EncodeFormData(data map[string]string) string {
	formData := ""
	for key, value := range data {
		if formData != "" {
			formData += "&"
		}
		formData += key + "=" + value
	}
	return formData
}

func GenerateAndroidID() string {
	// Android ID 是 8 字节的随机值
	bytes := make([]byte, 8) // 8 字节 = 64 位
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err) // 确保读取随机数时没有出错
	}
	return hex.EncodeToString(bytes)
}

func ExtractUrlPath(url string) string {
	re := regexp.MustCompile(`(https?://.+/).+?\.m3u8.*`)
	return re.ReplaceAllString(url, "$1")
}

// infoLogger 输出到 stdout
var infoLogger = log.New(os.Stdout, "[INFO] ", log.LstdFlags)

func LogInfo(msgs ...interface{}) {
	infoLogger.Println(msgs...)
}

var debugLogger = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)

func LogDebug(msgs ...interface{}) {
	if DebugMode {
		debugLogger.Println(msgs...)
	}
}

// 创建自定义 Logger，启用行号和文件名输出
var errorLogger = log.New(
	os.Stderr,
	"[ERROR] ",
	log.LstdFlags|log.Lshortfile, // 包含日期、时间、短文件名和行号
)

func LogError(msgs ...interface{}) {
	// 调用 Output 方法，calldepth=2 表示跳过 Logger 自身的调用栈
	errorLogger.Output(2, fmt.Sprintln(msgs...))
}

// func LogDebug(msgs ...interface{}) {
// 	if DebugMode {
// 		log.Println("[DEBUG]", msgs)
// 	}
// }

// func LogInfo(msgs ...interface{}) {
// 	log.Println("[INFO]", msgs)
// }
// func LogError(msgs ...interface{}) {
// 	log.Println("[ERROR]", msgs)
// }
