package v2

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go-login/tools"
	"net/http"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	timeoutThreshold = 500 * time.Millisecond
	extraTimeout     = 50 * time.Millisecond
	missedMaxError   = errors.New("密码错误次数过多")
	pwdError         = errors.New("密码错误")
)

func Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Millisecond)
	defer cancel()

	var cred User
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	// 4. 多级缓存，简化为内存缓存
	// 5. 优化逻辑：先缓存后数据库
	mCache := tools.GetCache()
	if mCache.Has(cred.Username, cred.Password) {
		c.JSON(http.StatusOK, gin.H{"status": "处理成功"})
		return
	}
	// 7. 错误用户优化处理
	isMissed, err := cred.ValidateMissed()
	if isMissed {
		// 7.1 计数器拦截
		if errors.Is(err, missedMaxError) {
			c.JSON(http.StatusForbidden, gin.H{"error": "密码错误次数过多"})
			return
		}
		// 7.2 缓存拦截
		if errors.Is(err, pwdError) {
			c.JSON(http.StatusForbidden, gin.H{"error": "密码错误"})
			return
		}
	}

	// 2. 创建带缓冲的结果通道
	resultCh := make(chan error, 1)

	// 3. 双路监听，启动异步处理
	go func() {
		// 1. 创建带超时的处理上下文
		processCtx, processCancel := context.WithTimeout(ctx, timeoutThreshold)
		defer processCancel()

		defer close(resultCh)
		resultCh <- tools.ProcessAndCache(processCtx, tools.User(cred))
		go tools.StartWorkersOnce()
		//fmt.Printf("用户%s插入队列完成\n", cred.Username)
	}()

	select {
	case <-ctx.Done(): // 优先级1：硬超时绝对优先
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求处理超时"})

		// 异步资源回收
		go func() {
			select {
			case <-resultCh: // 正常结束
			case <-time.After(extraTimeout): // 快速放弃
				cancel() // 传播取消
			}
		}()
		return // 关键：立即退出
	case err := <-resultCh:
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "处理成功"})
		//case <-processCtx.Done():
		//	// 6. 本来要后置模拟入库，懒得写了直接返回
		//	c.JSON(http.StatusAccepted, gin.H{
		//		"message": "请求已进入后台队列处理",
		//	})
		//
		//	go tools.StartWorkersOnce()
		//
		//	// 异步入队操作避免阻塞
		//	go func() {
		//		user := tools.User{
		//			Username: cred.Username,
		//			Password: cred.Password,
		//		}
		//		select {
		//		case tools.GetQueue() <- user:
		//		default:
		//			// 处理队列满的情况（如持久化存储）
		//			log.Println("队列已满，请求被丢弃")
		//		}
		//	}()
	}
}

func (u *User) ValidateMissed() (bool, error) {
	mUser, exists := tools.GetMissCache().GetMissUserInfo(u.Username, u.Password)
	if exists {
		//1. 计数器拦截
		if mUser.MissCount >= 3 {
			return true, missedMaxError
		}
		//2. 缓存拦截
		if mUser.Password == u.Password {
			go mUser.SaveMissUser() //加入计数器
			return true, pwdError
		}
	}
	return false, nil
}
