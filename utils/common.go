package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"io"
	"time"
)

const SessionPrefix = "sess_" // session 前缀

// 生成 64 位 随机 id
func GetSnowflakeId() string { // TODO: 这个随机 id 的各部分如何设置
	//default node id eq 1,this can modify to different serverId node
	node, _ := snowflake.NewNode(1)
	// Generate a snowflake ID.
	id := node.Generate().String()
	return id
}

// 返回 sha1 计算的哈希值
func Sha1(s string) (str string) {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// 返回格式化时间
func GetNowDateTime() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
}

// 获取随机 token
func GetRandomToken(length int) string { // TODO: 这个还没看懂是什么原理
	r := make([]byte, length)
	io.ReadFull(rand.Reader, r)
	return base64.URLEncoding.EncodeToString(r)
}

// 返回 sessionId，其实就是把前缀加上了
func CreateSessionId(sessionId string) string {
	return SessionPrefix + sessionId
}

// 通过 userId 生成 sessionId，也是有规律的，加了前缀
func GetSessionIdByUserId(userId int) string {
	return fmt.Sprintf("sess_map_%d", userId)
}

// 通过 sessionId 获取 session，返回 SessionPrefix + sessionId
func GetSessionName(sessionId string) string {
	return SessionPrefix + sessionId
}
