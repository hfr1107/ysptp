package main

import (
	"flag"
	"time"
	"ysptp/live"
	"ysptp/m3u"

	"github.com/gin-gonic/gin"
)

var tvM3uObj m3u.Tvm3u
var ysptpObj live.Ysptp

// 设置路由和处理逻辑
func setupRouter() *gin.Engine {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// // 配置HEAD请求的响应
	// r.HEAD("/", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "请求成功！")
	// })

	// // 配置GET请求的响应
	// r.GET("/", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "请求成功！")
	// })

	// 配置获取tv.m3u文件的路由
	r.GET("/", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=tv.m3u")
		tvM3uObj.GetTvM3u(c)
	})

	// 保留其他路径和对象的逻辑
	r.GET("/ysptp/:rid", func(c *gin.Context) {
		rid := c.Param("rid")

		ts := c.Query("ts")
		if ts == "" {
			ysptpObj.HandleMainRequest(c, rid)
		} else {
			ysptpObj.HandleTsRequest(c, ts, rid, c.Query("wsTime"), c.Query("wsSecret"))
		}

	})

	return r
}

func main() {

	host := flag.String("host", "0.0.0.0", "host")
	port := flag.String("p", "16384", "port")
	flag.BoolVar(&live.DebugMode, "debug", false, "开启调试模式")
	flag.IntVar(&live.UIDCount, "UIDCount", 3, "UID负载均衡的个数，范围3到8")
	flag.BoolVar(&live.EnableCache, "cache", false, "启用预先缓存")
	flag.Parse()
	if live.UIDCount < 3 {
		live.UIDCount = 3
	}
	if live.UIDCount > live.UIDMax {
		live.UIDCount = live.UIDMax
	}
	if live.UIDCount < 6 && live.EnableCache {
		live.UIDCount = 6
	}
	// live.Host = *host
	// live.Port = *port

	// live.GetBaseM3uUrl("Live1717729995180256")
	// live.GetBaseM3uUrl("Live1718261577870260")
	// live.GetBaseM3uUrl("Live1718261955077261")
	// live.GetBaseM3uUrl("Live1718276148119264")
	// live.GetBaseM3uUrl("Live1719474204987287")
	// live.GetBaseM3uUrl("Live1719473996025286")
	// live.GetBaseM3uUrl("Live1718276412224269")
	// live.GetBaseM3uUrl("Live1718276458899270")
	// live.GetBaseM3uUrl("Live1718276503187272")
	// live.GetBaseM3uUrl("Live1718276550002273")
	// live.GetBaseM3uUrl("Live1718276603690275")
	// live.GetBaseM3uUrl("Live1718276623932276")
	// live.GetBaseM3uUrl("Live1718276575708274")
	// live.GetBaseM3uUrl("Live1718276498748271")
	// live.GetBaseM3uUrl("Live1718276319614267")
	// live.GetBaseM3uUrl("Live1718276256572265")
	// live.GetBaseM3uUrl("Live1718276138318263")
	// live.GetBaseM3uUrl("Live1719392219423280")
	// live.GetBaseM3uUrl("Live1719392670442283")
	// live.GetBaseM3uUrl("Live1719392779653284")
	// live.GetBaseM3uUrl("Live1719392885692285")
	// live.GetBaseM3uUrl("Live1719392560433282")
	// live.GetBaseM3uUrl("Live1719392360336281")
	// live.GetBaseM3uUrl("Live1704966749996185")
	// live.GetBaseM3uUrl("Live1704872878572161")
	// live.GetBaseM3uUrl("Live1688400593818102")

	live.GetUIDs()
	live.GetGUIDs()
	live.CheckPlayAuth()
	live.GetAppSecret()
	//live.SetCache("check", "", "", "", "", "")

	// live.LogInfo("开始初始化缓存")
	// live.RefreshM3u8Cache()
	// live.LogInfo("初始化缓存完成")

	// 创建一个通道用于停止定时任务
	done := make(chan bool)

	// 启动定时任务（goroutine）
	go timedFunction(done)

	r := setupRouter()
	live.LogInfo("可通过 -h 查看帮助")
	live.LogInfo("Listen on "+*host+":"+*port, "...")
	live.LogInfo("现在可以使用浏览器访问 http://你的ip:" + *port + " 获取m3u文件")
	r.Run(*host + ":" + *port)

	done <- true // 发送停止信号

}

// 定时执行的函数
func timedFunction(done <-chan bool) {
	// 创建一个定时器，每隔 ? 秒触发一次
	ticker := time.NewTicker(3300 * time.Second)
	defer ticker.Stop() // 确保结束时释放资源

	for {
		select {
		case <-done:
			// 收到停止信号，退出函数
			return
		case <-ticker.C:
			// 这是定时执行的业务逻辑
			live.RefreshM3u8Cache()
			//fmt.Println("定时任务执行")
		}
	}
}
