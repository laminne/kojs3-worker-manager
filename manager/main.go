package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/laminne/kemomimi-ojs/manager"
	"net/http"
	"time"
)

type RunningStatus struct {
	TaskID string       // 問題ID
	Status []TestStatus // テストケースごと
}

type TestStatus struct {
	TestID     string // テストケースのID
	ExitStatus int    // 実行時のステータスコード
	Duration   int    // 実行秒数
	Status     string // WA RE などのステータス
}

type Running struct {
	StartTime time.Time
	Status    string
}

func main() {
	engine := gin.Default()
	engine.POST("/run", handler)
	engine.Run(":3000")
}

func handler(c *gin.Context) {
	var code manager.Code
	err := c.BindJSON(&code)
	if err != nil {
		return
	}
	res := manager.Start(code)
	if res == "Error" {
		c.PureJSON(http.StatusOK, `{"TaskID":"`+code.TaskID+`","Status":[{"TestID":"","ExitStatus":-1,"Duration":11000,"Status":"TLE"},{"TestID":"","ExitStatus":-1,"Duration":11000,"Status":"TLE"},{"TestID":"","ExitStatus":-1,"Duration":11000,"Status":"TLE"}]}`)
		return
	}
	var responseData RunningStatus

	err = json.Unmarshal([]byte(res), &responseData)
	if err != nil {
		return
	}
	fmt.Println(res)
	c.JSON(http.StatusOK, responseData)
}
