package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/model"
)

func AddMenu(ctx *gin.Context) {
	menu := model.Hawthorn_sys_menu{}
	ctx.ShouldBindJSON(&menu)
	//db.DB.Session(&gorm.Session{}).Raw("select max(menu_id)+1 as menu_id from sys_menu", &menu.Menu_id)
	dbs.GetDB().Create(&menu)
}
func UpdMenu(ctx *gin.Context) {
	menu := model.Hawthorn_sys_menu{}
	ctx.ShouldBindJSON(&menu)
	//db.DB.Session(&gorm.Session{}).Raw("select max(menu_id)+1 as menu_id from sys_menu", &menu.Menu_id)
	//fmt.Println(menu.Menu_id)
	dbs.GetDB().Create(&menu)
}

func GetRoleMenus(c *gin.Context) {
	var menus []model.Hawthorn_sys_menu
	if err := dbs.GetDB().Where("menu_type != ?", "A").Find(&menus).Error; err != nil {
		c.JSON(500, gin.H{"error": "数据库查询失败: " + err.Error()})
		return
	}
	idMap := make(map[int]*model.Hawthorn_sys_menu)
	for i := range menus {
		menu := &menus[i]
		idMap[menu.Menu_id] = menu
	}

	var menusA []model.Hawthorn_sys_menu
	if err := dbs.GetDB().Where("menu_type = ?", "A").Find(&menusA).Error; err != nil {
		c.JSON(500, gin.H{"error": "数据库查询失败: " + err.Error()})
		return
	}
	//for index := range menusA {
	//	parentId := menusA[index].Parent_id
	//	append(idMap[parentId].Children, &menusA[index])
	//}

	//idMap := make(map[int]*mapper.Hawthorn_sys_menu)
	childrenMap := make(map[int][]*model.Hawthorn_sys_menu) // 父ID -> 子节点列表
	for i := range menus {
		menu := &menus[i]
		idMap[menu.Menu_id] = menu
		childrenMap[menu.Parent_id] = append(childrenMap[menu.Parent_id], menu)
	}
	for _, menu := range idMap {
		if children, exists := childrenMap[menu.Menu_id]; exists {
			menu.Children = children
			menu.Has_Children = true
		} else {
			menu.Children = []*model.Hawthorn_sys_menu{} // 确保空数组而非null
			menu.Has_Children = false
		}
	}
	fmt.Println(childrenMap[0])
	topMenus := childrenMap[0]
	if topMenus == nil {
		topMenus = []*model.Hawthorn_sys_menu{}
	}
	c.JSON(200, topMenus)

}

func GetAllMenus(c *gin.Context) {
	var menus []model.Hawthorn_sys_menu
	if err := dbs.GetDB().Find(&menus).Error; err != nil {
		c.JSON(500, gin.H{"error": "数据库查询失败: " + err.Error()})
		return
	}
	idMap := make(map[int]*model.Hawthorn_sys_menu)
	childrenMap := make(map[int][]*model.Hawthorn_sys_menu) // 父ID -> 子节点列表
	for i := range menus {
		menu := &menus[i]

		idMap[menu.Menu_id] = menu
		childrenMap[menu.Parent_id] = append(childrenMap[menu.Parent_id], menu)
	}
	for _, menu := range idMap {
		if children, exists := childrenMap[menu.Menu_id]; exists {
			menu.Children = children
			menu.Has_Children = true
		} else {
			menu.Children = []*model.Hawthorn_sys_menu{} // 确保空数组而非null
			menu.Has_Children = false
		}
	}
	fmt.Println(childrenMap[0])
	topMenus := childrenMap[0]
	if topMenus == nil {
		topMenus = []*model.Hawthorn_sys_menu{}
	}
	c.JSON(200, topMenus)
}
