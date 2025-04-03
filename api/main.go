package api
import (
	"net/http"
	"ysptp/m3u"
	"ysptp/live"
	"github.com/gin-gonic/gin"
)
var tvM3uObj m3u.Tvm3u
var ysptpObj live.Ysptp
var btimeObj live.Btime
var m1905Obj live.M1905

func Register(r *gin.Engine) {
	r.NoRoute(ErrRouter)

	r.HEAD("/", func(c *gin.Context) {
		c.String(http.StatusOK, "请求成功！")
	})

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "请求成功！")
	})
	r.GET("/m1905/cctv6.m3u8", func(c *gin.Context) {
		m1905Obj.HandleMainRequest(c)
	})
	r.GET("/btime/:rid", func(c *gin.Context) {
		rid := c.Param("rid")
		btimeObj.HandleMainRequest(c, rid)
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

/*
	r.GET("/:path/:rid", func(c *gin.Context) {
		enableTV := true
		//path := c.Param("path")
		rid := c.Param("rid")
		ts := c.Query("ts")
		
		if enableTV {
			itvobj := &liveurls.Itv{}
			cdn := c.Query("cdn")
			if ts == "" {
				itvobj.HandleMainRequest(c, cdn, rid)
			} else {
				itvobj.HandleTsRequest(c, ts)
			}
		} else {
			c.String(http.StatusForbidden, "公共服务不提供TV直播")
		}	
	})
 

	app.GET("/ping", handler.Ping)
        app.GET("/:path/:rid", handler.Feiyang)
	route := app.Group("/api")
	{
		route.GET("/hello/:name", handler.Hello)
		//route.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
 */
}

func ErrRouter(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"errors": "this page could not be found",
	})
}



var (
	app *gin.Engine
)

// @title Golang Vercel Deployment
// @description API Documentation for Golang deployment in vercel serverless environment
// @version 1.0

// @schemes https http
// @host golang-vercel.vercel.app
func init() {
	app = gin.New()
	Register(app)
}

// Entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
