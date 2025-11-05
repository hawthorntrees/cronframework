package controller

import "github.com/gin-gonic/gin"

func RegisterRouter(engine *gin.RouterGroup) {
	engine.POST("/addMenu", AddMenu)
	engine.POST("/getAllMenus", GetAllMenus)
	engine.POST("/getRoleMenus", GetRoleMenus)
	engine.POST("/getMenus", GetMenus)
	engine.POST("/login", Login)
	engine.POST("/getRoles", GetRoles)
}
