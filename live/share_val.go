package live

import (
	"net/http"
	"time"
)

// var Host string = "0.0.0.0"
// var Port string = "16384"
var DebugMode = false

var EnableCache = false

var UIDCount = 3

const UIDMax = 8

type UIDData struct {
	UID     string
	UIDInit bool
	GUID    string
}

var UIDsData []UIDData

var AppSecret = "e7e259bbb0ac4848ba70921c860a1216"

const AppId = "5f39826474a524f95d5f436eacfacfb67457c4a7"

// const AppSecret = "e7e259bbb0ac4848ba70921c860a1216"
const AppVersion = "1.3.4"
const UA = "cctv_app_tv"
const Referer = "api.cctv.cn"
const PubKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC/ZeLwTPPLSU7QGwv6tVgdawz9n7S2CxboIEVQlQ1USAHvBRlWBsU2l7+HuUVMJ5blqGc/5y3AoaUzPGoXPfIm0GnBdFL+iLeRDwOS1KgcQ0fIquvr/2Xzj3fVA1o4Y81wJK5BP8bDTBFYMVOlOoCc1ZzWwdZBYpb4FNxt//5dAwIDAQAB"
const EncryptedAppSecret = "S5eEygInW66uSlPc89aam7vNa8H7Dm9ukq5fBQXaPRNqaHNxzzsY1Xoi2weq43UnVSO3ysIFO0FZROzcIpa29J0BRuzGSVWkvVcedOP2Q4Ksz0osenCbzteqU9EgVvGewZF7gSQ/+XUIAZvHnf1AArUjNBAxpE3IL7dMZQWJVM4="
const UrlCloudwsRegister = "https://ytpcloudws.cctv.cn/cloudps/wssapi/device/v1/register"
const UrlCloudwsGet = "https://ytpcloudws.cctv.cn/cloudps/wssapi/device/v1/get"
const UrlCheckPlayAuth = "https://ytpaddr.cctv.cn/gsnw/play/check/obtain"
const UrlGetBaseM3u8 = "https://ytpaddr.cctv.cn/gsnw/live"
const UrlGetAppSecret = "https://ytpaddr.cctv.cn/gsnw/tpa/sk/obtain"
const UrlGetStream = "https://ytpvdn.cctv.cn/cctvmobileinf/rest/cctv/videoliveUrl/getstream"

var CCTVList = map[string]string{
	"cctv1.m3u8":  "Live1717729995180256",
	"cctv2.m3u8":  "Live1718261577870260",
	"cctv3.m3u8":  "Live1718261955077261",
	"cctv4.m3u8":  "Live1718276148119264",
	"cctv5.m3u8":  "Live1719474204987287",
	"cctv5p.m3u8": "Live1719473996025286",
	//"cctv6.m3u8":      "Live1631782155853158",
	"cctv7.m3u8":      "Live1718276412224269",
	"cctv8.m3u8":      "Live1718276458899270",
	"cctv9.m3u8":      "Live1718276503187272",
	"cctv10.m3u8":     "Live1718276550002273",
	"cctv11.m3u8":     "Live1718276603690275",
	"cctv12.m3u8":     "Live1718276623932276",
	"cctv13.m3u8":     "Live1718276575708274",
	"cctv14.m3u8":     "Live1718276498748271",
	"cctv15.m3u8":     "Live1718276319614267",
	"cctv16.m3u8":     "Live1718276256572265",
	"cctv17.m3u8":     "Live1718276138318263",
	"cgtnen.m3u8":     "Live1719392219423280",
	"cgtnfr.m3u8":     "Live1719392670442283",
	"cgtnru.m3u8":     "Live1719392779653284",
	"cgtnar.m3u8":     "Live1719392885692285",
	"cgtnes.m3u8":     "Live1719392560433282",
	"cgtndoc.m3u8":    "Live1719392360336281",
	"cctv16_4k.m3u8":  "Live1704966749996185",
	"cctv4k.m3u8":     "Live1704872878572161",
	"cctv8k_36m.m3u8": "Live1688400593818102",
}

var DeviceModel = [...]string{
	"Sony-KD-85Z9G",
	"Sony-KD-55A90J",
	"LG-OLED65C1",
	"Samsung-QN65Q90TAFXZA",
	"Samsung-QN85QN900BFXZA",
}

var Client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		ForceAttemptHTTP2:   false,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	},
}
