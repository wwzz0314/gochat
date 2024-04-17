package router

import (
	"github.com/gin-gonic/gin"
	"gochat/api/handler"
	"gochat/api/rpc"
	"gochat/logic/dao"
	"gochat/proto"
	"gochat/tools"
	"net/http"
	"strings"
)

func Register() *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware())
	initUserRouter(r)
	initPushRouter(r)
	r.NoRoute(func(c *gin.Context) {
		tools.FailWithMsg(c, "please check request url !")
	})
	return r
}

func initUserRouter(r *gin.Engine) {
	userGroup := r.Group("/user")
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/register", handler.Register)
	userGroup.Use(CheckSessionId())
	{
		userGroup.GET("/checkAuth", handler.CheckAuth)
		userGroup.POST("/logout", handler.Logout)
	}
}

func initPushRouter(r *gin.Engine) {
	pushGroup := r.Group("/push")
	pushGroup.Use(CheckSessionId())
	{
		pushGroup.POST("/push", handler.Push)
		pushGroup.POST("/pushRoom", handler.PushRoom)
		pushGroup.POST("/count", handler.Count)
		pushGroup.POST("/getRoomInfo", handler.GetRoomInfo)
	}
}

type FormCheckSessionId struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

func CheckSessionId() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Request.Header.Get("Authorization")
		authToken = strings.Trim(authToken, "Bearer")
		authToken = strings.TrimSpace(authToken)
		c.Request.Header.Set("Authorization", authToken)
		if authToken == "" {
			c.Abort()
			tools.ResponseWithCode(c, tools.CodeSessionError, "UnAuthorized", nil)
			return
		}
		req := &proto.CheckAuthRequest{
			AuthToken: authToken,
		}
		code, userId, userName := rpc.RpcLogicObj.CheckAuth(req)
		if code == tools.CodeFail || userId <= 0 || userName == "" {
			c.Abort()
			tools.ResponseWithCode(c, tools.CodeSessionError, "UnAuthorized", nil)
			return
		} else {
			user := dao.User{
				Id:       userId,
				UserName: userName,
			}
			c.Set("user", user)
		}
		c.Next()
		return
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		var openCorsFlag = true
		if openCorsFlag {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET,OPTIONS,POST,PUT,DELETE")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, nil)
		}
		c.Next()
	}
}
