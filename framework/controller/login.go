package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/dto/login"
	"github.com/hawthorntrees/cronframework/framework/dto/resp"
	"github.com/hawthorntrees/cronframework/framework/model"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"gorm.io/gorm"
	"strconv"
)

func Login(ctx *gin.Context) {
	loginInfo := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{}
	ctx.ShouldBindJSON(&loginInfo)
	user := model.Hawthorn_sys_user{}
	db := dbs.GetDB()

	result := db.Session(&gorm.Session{}).First(&user, "user_id=?", loginInfo.Name)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			resp.Error(ctx, "用户不存在")
			return
		} else {
			resp.Error(ctx, result.Error.Error())
			return
		}
	}
	//2. 验证密码
	token, err := utils.GenerateToken(loginInfo.Name, loginInfo.Password)
	if err != nil {
		ctx.JSON(200, err)
	} else {
		resp.Success(ctx, resp.RespJson{"token": token})
	}
}

// 这里主要是返回要显示的菜单即可，不需要带按钮
//func GetMenus_bak(ctx *gin.Context) {
//	sql := "select menu.* from sys_menu menu ,sys_role_menus menus where menus.role_id='AA1' and menu.menu_id = menus.menu_id and menu.menu_type != 'A'"
//	menus := []*mapper.Hawthorn_sys_menu{}
//	dbs.GetSession().Raw(sql).Scan(&menus)
//	idMap := make(map[int]*mapper.Hawthorn_sys_menu)
//	childMap := make(map[int][]*mapper.Hawthorn_sys_menu)
//	for _, menu := range menus {
//		idMap[menu.Menu_id] = menu
//		childMap[menu.Parent_id] = append(childMap[menu.Parent_id], menu)
//	}
//	for _, menu := range idMap {
//		if children, exists := childMap[menu.Menu_id]; exists {
//			menu.Children = children
//		} else {
//			menu.Children = []*mapper.Hawthorn_sys_menu{}
//		}
//	}
//	topMenu := childMap[0]
//	ctx.JSON(200, topMenu)
//}

func GetMenus(ctx *gin.Context) {
	sql := "select menu.* from sys_menu menu ,sys_role_menus menus where menus.role_id='AA1' and menu.menu_id = menus.menu_id and menu.menu_type != 'A'"
	menus := []*login.Sys_menu_dto{}
	dbs.GetDB().Session(&gorm.Session{}).Raw(sql).Scan(&menus)
	idMap := make(map[int]*login.Sys_menu_dto)
	childMap := make(map[int][]*login.Sys_menu_dto)
	for _, menu := range menus {
		idMap[menu.Menu_id] = menu
		if menu.Path == "" {
			menu.Path = strconv.Itoa(menu.Menu_id)
		}
		childMap[menu.Parent_id] = append(childMap[menu.Parent_id], menu)
	}
	for _, menu := range idMap {
		if children, exists := childMap[menu.Menu_id]; exists {
			menu.Children = children
		} else {
			menu.Children = nil
		}
	}
	topMenu := childMap[0]
	ctx.JSON(200, topMenu)
}
