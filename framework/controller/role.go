package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/dto/resp"
	"github.com/hawthorntrees/cronframework/framework/model"
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
