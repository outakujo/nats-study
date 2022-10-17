package main

import (
	"github.com/gin-gonic/gin"
	"nats-study/model"
	"net/http"
)

func main() {
	eng := gin.New()
	eng.POST("add", func(c *gin.Context) {
		var da = struct {
			To   model.Target
			From model.Target
		}{}
		err := c.ShouldBindJSON(&da)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		top, err := model.Liaotian.Add(da.To, da.From)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		c.JSON(http.StatusOK, top)
	})
	eng.DELETE("del", func(c *gin.Context) {
		var da = struct {
			To   string
			From string
		}{}
		err := c.ShouldBindJSON(&da)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		err = model.Liaotian.Del(da.To, da.From)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		c.JSON(http.StatusOK, "ok")
	})
	eng.POST("send", func(c *gin.Context) {
		var da = struct {
			To   string
			From string
			Msg  string
		}{}
		err := c.ShouldBindJSON(&da)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		se, err := model.Liaotian.Get(da.To, da.From)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		err = se.Send(model.Msg{Body: []byte(da.Msg), Type: model.TextMsg})
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}
		c.JSON(http.StatusOK, "ok")
	})
	eng.Run()
}
