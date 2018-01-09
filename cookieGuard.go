package main

//import "fmt"
import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"strings"
	"github.com/op/go-logging"
	"os"
	"bytes"
	"net/url"
)

var log = logging.MustGetLogger("fly")
var format = logging.MustStringFormatter(
	`%{time:2006-01-02 15:04:05 } %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
)

type Cookie struct {
	Domain   string `json:"domain"`
	HostOnly bool   `json:"hostOnly"`
	HTTPOnly bool   `json:"httpOnly"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	SameSite string `json:"sameSite"`
	Secure   bool   `json:"secure"`
	Session  bool   `json:"session"`
	StoreID  string `json:"storeId"`
	Value    string `json:"value"`
}

func initLog(){
	logFile, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil{
		panic(err.Error())
	}
	// For demo purposes, create two backend for os.Stderr.
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend3 := logging.NewLogBackend(logFile, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	backend3Formatter := logging.NewBackendFormatter(backend3, format)
	backend3Leveled := logging.AddModuleLevel(backend3Formatter)
	backend3Leveled.SetLevel(logging.DEBUG, "")


	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter, backend3Leveled)
}

type Cache struct{
	cookies map[string][]*Cookie
}

var CACHE *Cache = &Cache{}

func initConfig(){
	f, err := os.Open("config.json")
	if err != nil{
		panic(err.Error())
	}
	bts := make([]byte, 1024, 1024)
	_, err = f.Read(bts)
	if err != nil{
		panic(err.Error())
	}
}

func getFilterDomains(host string)[]string{
	blocks := strings.Split(host, ".")
	rst := []string{host}
	if len(blocks) <= 2{
		return rst
	}
	for i:=1; i<=len(blocks)-2; i++{
		s := "." + strings.Join(blocks[i:len(blocks)], ".")
		rst = append(rst, s)
	}
	return rst
}

func getCookies(u string)([]*Cookie, error){
	var rst []*Cookie
	url, err := url.Parse(u)
	if err != nil{
		return rst, err
	}
	host := url.Host
	filters := getFilterDomains(host)
	fmt.Println("filters", filters)
	for _, f := range filters{
		rst = append(rst, CACHE.cookies[f]...)
	}
	return rst, nil
}


type ProxyRequest struct{
	Url string `json: url`
	Method string `json: method`
	Body string `json: body`
	ContentType string `json: contentType`
}



func proxyHttpRequest(proxyReq *ProxyRequest)(*http.Response, error){
	req, err := http.NewRequest(proxyReq.Method, proxyReq.Url, strings.NewReader(proxyReq.Body))
	if err != nil{
		fmt.Println(err.Error())
		return nil, err
	}

	cookies, err := getCookies(proxyReq.Url)

	if err != nil{
		return nil, err
	}

	for i:=0; i<len(cookies); i++{
		c := cookies[i]
		ck := http.Cookie{
			Name: c.Name,
			Value: c.Value,
			Path: c.Path,
			Domain: c.Domain,
			Secure: c.Secure,
			HttpOnly: c.HTTPOnly,
		}
		req.AddCookie(&ck)
	}
	fmt.Println(proxyReq)
	req.Header.Add("Content-Type", proxyReq.ContentType)
	fmt.Println(req.Header)
	client := http.Client{}
	resp, err := client.Do(req)
	return resp, err
}

func main() {
	initLog()
	//initConfig()


	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, CACHE.cookies)
	})

	router.GET("/cookies", func(c *gin.Context) {
		q := c.Query("q")
		rst, _ := getCookies(q)
		c.JSON(http.StatusOK, rst)
	})

	router.POST("/cookies", func(c *gin.Context) {
		cookies := make([]*Cookie, 0, 10)
		err := c.BindJSON(&cookies)
		m := make(map[string][]*Cookie)
		for i:=0; i<len(cookies); i++{
			ck := cookies[i]
			m[ck.Domain] = append(m[ck.Domain], ck)
		}
		CACHE.cookies = m
		fmt.Println(CACHE)
		c.JSON(http.StatusOK, gin.H{"success": err == nil, "cookies": cookies})
	})

	router.POST("/proxy", func(c *gin.Context) {
		req := &ProxyRequest{}
		err := c.BindJSON(req)
		resp, err := proxyHttpRequest(req)
		buf := new(bytes.Buffer)
		msg := ""
		if(err != nil){
			fmt.Println(err.Error())
			msg = err.Error()
		}else{
			buf.ReadFrom(resp.Body)
			fmt.Println("response==>", buf.String())
		}
		c.JSON(http.StatusOK, gin.H{"success": err == nil,
									"proxy": req,
									"msg": msg,
									"resp": buf.String()})
	})

	router.Run(":9090")

}