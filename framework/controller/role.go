package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/dto/resp"
	"github.com/hawthorntrees/cronframework/framework/model"
	"github.com/xuri/excelize/v2"
	"net/url"
)

func GetRoles(c *gin.Context) {
	var total int64 = 0
	dbs.GetDB().Table("sys_role").Count(&total)
	sql := "select * from sys_role order by role_id"
	role := []model.Hawthorn_sys_role{}
	tx := dbs.GetDB().Raw(sql).Scan(&role)
	if tx.Error != nil {
		fmt.Print("sb")
	}

	result := resp.PageResult{
		Data:  &role,
		Total: total,
	}
	resp.Success(c, &result)
}

var roleTitle = []interface{}{"角色编号", "角色名称"}

func ExportRoles(c *gin.Context) {
	//log := logger.GetLogger(c)
	var roles []model.Hawthorn_sys_role
	tx := dbs.GetDB().Table("sys_role").Find(&roles)
	if tx.Error != nil {
		resp.Error(c, fmt.Sprintf("查询角色数据失败：%v", tx.Error))
		return
	}

	file := excelize.NewFile()
	defer func(file *excelize.File) {
		err := file.Close()
		if err != nil {
			resp.Error(c, err.Error())
			return
		}
	}(file)
	sheet1 := file.GetSheetName(0)
	stream, err := file.NewStreamWriter(sheet1)
	if err != nil {
		resp.Error(c, err.Error())
		return
	}
	if err := stream.SetRow("A1", roleTitle); err != nil {
		resp.Error(c, err.Error())
		return
	}

	for i, role := range roles {
		n := fmt.Sprintf("A%d", i+2)
		data := []interface{}{role.Role_id, role.Role_name}
		if err := stream.SetRow(n, data); err != nil {
			resp.Error(c, err.Error())
			return
		}
	}
	if err := stream.Flush(); err != nil {
		resp.Error(c, fmt.Sprintf("刷新流数据失败：%v", err))
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape("角色列表.xlsx")))
	if err := file.Write(c.Writer); err != nil { /* 错误处理 */
		resp.Error(c, err.Error())
	}
	return
}
