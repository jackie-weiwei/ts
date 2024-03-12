package ts

import (
	"container/list"

	"github.com/gin-gonic/gin"
)

type FuncRouter func(*gin.RouterGroup)

var (
	routers *list.List
)

func init() {
	routers = list.New()
}

func RegisterRouter(funcRouter FuncRouter) {
	routers.PushBack(funcRouter)
}

func InitRouter(r *gin.RouterGroup) {
	for e := routers.Front(); e != nil; e = e.Next() {
		e.Value.(FuncRouter)(r)
	}
}
