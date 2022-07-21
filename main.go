package main

import (
	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/memcache"
	_ "github.com/astaxie/beego/session/redis"
	_ "github.com/astaxie/beego/session/redis_cluster"
	_ "github.com/simonblowsnow/mm-wiki-ex/app"
)

func main() {
	beego.Run()
}
