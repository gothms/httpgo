package test

import (
	"github.com/gothms/httpgo/framework/gin"
	"net/http"
	"reflect"
	"testing"
)

func TestGET(t *testing.T) {
	r := gin.Default()
	//r.GET("/ping", func(c *gin.Context) {
	//r.GET("/a:a", func(c *gin.Context) {
	//r.GET("/a/*a", func(c *gin.Context) {
	//r.GET("/a:a/", func(c *gin.Context) {
	r.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		c.String(http.StatusOK, "%s is %s", name, action)
	}) // localhost:8080/user/lee/send：lee is /send
	r.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}
func TestPOST(t *testing.T) {
	r := gin.Default()
	r.POST("/form", func(c *gin.Context) {
		message := c.PostForm("message")
		name := c.DefaultPostForm("name", "Lee")
		m := c.PostFormMap("map")
		c.JSON(200, gin.H{
			"message": message,
			"name":    name,
			"map":     m,
		})
	})
	r.Run()
}
func TestLookup(t *testing.T) {
	r := gin.Default()
	f := func(c *gin.Context) {
		//r.GET("/a:a/b:b/abc/*", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
	r.GET("/a:a/", f)
	handlers, params, tsr, fullpath := r.Lookup(http.MethodGet, "/a:a/")

	t.Logf("%p\n", f)
	t.Log(handlers)
	for _, fn := range handlers {
		//t.Logf("%s\n", fn)
		//t.Log(reflect.ValueOf(fn))
		t.Log(reflect.TypeOf(fn).Kind())
	}
	t.Log(*params)
	t.Log(tsr)
	t.Log(fullpath)
	//r.Run()
}
