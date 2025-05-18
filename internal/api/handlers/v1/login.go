package v1

import (
	"github.com/gin-gonic/gin"
	"go-login/internal/common"
	"go-login/tools"
	"net/http"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	//1. 获取用户名和密码
	var cred User
	if err := c.BindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if cred.Username == "" || cred.Password == "" {
		c.JSON(400, gin.H{"error": "用户名或密码不能为空"})
		return
	}
	//2. 模拟第三方等待时间
	tools.FakeToWait()
	//3. 模拟第三方返回结果
	commonUser := common.User{
		Username: cred.Username,
		Password: cred.Password,
	}
	hasUser, err := commonUser.Login()
	if err != nil || !hasUser {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
	}
	//4. 模拟mysql入库
	tools.FakeToMysql()
	//5. 模拟redis入库
	tools.FakeToRedis()
	//6. 返回结果
	c.String(http.StatusOK, "success")
}
