package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/beego/beego/v2/server/web/context"

	"github.com/beego/beego/v2/server/web"
	"github.com/cdle/jd_study/jdc/controllers"
	"github.com/cdle/jd_study/jdc/models"
)

var help = "-p 运行端口\n-qla 青龙登录地址\n-qlu 青龙登录用户名\n-qlp 青龙登录密码\n-v4 配置文件路径"

func main() {
	l := len(os.Args)
	if l == 0 {
		fmt.Println(help)
		return
	}

	for i, arg := range os.Args {
		if i+1 <= l-1 {
			v := os.Args[i+1]
			switch arg {
			case "-h":
				fmt.Println(help)
				return
			case "-p":
				p, _ := strconv.Atoi(v)
				web.BConfig.Listen.HTTPPort = p
			case "-qla":
				vv := regexp.MustCompile(`^(https?://[\.\w]+:?\d*)`).FindStringSubmatch(v)
				if len(vv) == 2 {
					models.QlAddress = vv[1]
				}
			case "-qlu":
				models.QlUserName = v
			case "-qlp":
				models.QlPassword = v
			case "-v4":
				models.V4Config = v
			case "-m":
				models.Master = v
			case "-l":
				models.ListConfig = v
			case "-f":
				models.QrcodeFront = v
			}
		}
	}
	if models.V4Config != "" {
		f, err := os.Open(models.V4Config)
		if err != nil {
			logs.Warn("无法打开V4配置文件，请检查路径是否正确")
			return
		}
		f.Close()
	} else if models.ListConfig != "" {
		f, err := os.Open(models.ListConfig)
		if err != nil {
			logs.Warn("无法打开指定配置文件，请检查路径是否正确")
			return
		}
		f.Close()
	} else {
		if models.QlAddress == "" {
			logs.Warn("未指定青龙登录地址")
			return
		}
		if models.QlUserName == "" {
			logs.Warn("未指定青龙登录用户名")
			return
		}
		if models.QlPassword == "" {
			logs.Warn("未指定青龙登录密码")
			return
		}
		if models.GetToken(); models.Token == "" {
			logs.Warn("JDC无法与青龙面板取得联系，请检查账号")
			return
		} else {
			models.QlVersion, _ = models.GetQlVersion(models.QlAddress)
			logs.Info("JDC成功接入青龙" + models.QlVersion)
		}
	}
	models.Save <- &models.JdCookie{}
	web.Get("/", func(ctx *context.Context) {
		if models.QrcodeFront != "" {
			if strings.Contains(models.QrcodeFront, "http://") {
				s, _ := httplib.Get(models.QrcodeFront).String()
				ctx.WriteString(s)
				return
			} else {
				f, err := os.Open(models.QrcodeFront)
				if err == nil {
					d, _ := ioutil.ReadAll(f)
					ctx.WriteString(string(d))
					return
				}
			}
		}
		ctx.WriteString(models.Qrocde)
	})
	web.Router("/api/login/qrcode", &controllers.LoginController{}, "get:GetQrcode")
	web.Router("/api/login/query", &controllers.LoginController{}, "get:Query")
	web.Router("/api/account", &controllers.AccountController{}, "get:List")
	web.Router("/api/account", &controllers.AccountController{}, "post:CreateOrUpdate")
	web.Router("/admin", &controllers.AccountController{}, "get:Admin")
	web.BConfig.AppName = "jdc"
	web.BConfig.WebConfig.AutoRender = false
	web.BConfig.CopyRequestBody = true
	web.BConfig.WebConfig.Session.SessionOn = true
	web.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600
	web.BConfig.WebConfig.Session.SessionName = "jdc"
	web.Run()
}