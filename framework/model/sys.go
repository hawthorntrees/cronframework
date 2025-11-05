package model

type Hawthorn_sys_menu struct {
	Menu_id      int                  `gorm:"type:int;comment:'菜单编号';primaryKey" json:"menu_id"`
	Menu_name    string               `gorm:"type:varchar(50);comment:'菜单名称'" json:"menu_name"`
	Parent_id    int                  `gorm:"type:int;comment:'父菜单编号'" json:"parent_id"`
	Order_num    int                  `gorm:"type:int;comment:'显示顺序'" json:"order_num"`
	Path         string               `gorm:"type:varchar(100);comment:'路由地址'" json:"path"`
	Component    string               `gorm:"type:varchar(100);comment:'组件路径'" json:"component"`
	Is_cache     string               `gorm:"type:varchar(2);comment:'是否缓存0-是1-否'" json:"is_cache"`
	Menu_type    string               `gorm:"type:varchar(2);comment:'菜单类型M-菜单P-页面A权限'" json:"menu_type"`
	Status       string               `gorm:"type:varchar(2);comment:'菜单状态0-正常1-停用'" json:"status"`
	Perms        string               `gorm:"type:varchar(100);comment:'权限标识冒号分割'" json:"perms"`
	Url          string               `gorm:"type:varchar(100);comment:'url'" json:"url"`
	Icon         string               `gorm:"type:varchar(100);comment:'菜单图标'" json:"icon"`
	Create_by    string               `gorm:"type:varchar(50);comment:'创建者'" json:"create_by"`
	Create_time  string               `gorm:"type:varchar(17);comment:'创建时间'" json:"create_time"`
	Update_by    string               `gorm:"type:varchar(20);comment:'更新者'" json:"update_by"`
	Update_time  string               `gorm:"type:varchar(20);comment:'更新时间'" json:"update_time"`
	Remark       string               `gorm:"type:varchar(17);comment:'备注'" json:"remark"`
	Children     []*Hawthorn_sys_menu `gorm:"-" json:"children"`
	Has_Children bool                 `gorm:"-" json:"hasChildren"`
}
type Hawthorn_sys_role_menus struct {
	Role_id string `gorm:"type:varchar(10);comment:'角色编号';primaryKey" json:"role_id"`
	Menu_id int    `gorm:"type:int;comment:'菜单编号';primaryKey" json:"menu_id"`
}
type Hawthorn_sys_role_perms struct {
	Role_id string `gorm:"type:varchar(10);comment:'角色编号';primaryKey" json:"role_id"`
	Perms   string `gorm:"type:varchar(50);comment:'权限标识';primaryKey" json:"perms"`
}
type Hawthorn_sys_perms_url struct {
	Perms string `gorm:"type:varchar(50);comment:'权限标识';primaryKey" json:"perms"`
	Url   string `gorm:"type:varchar(50);comment:'url';primaryKey" json:"url"`
}

type Hawthorn_sys_role struct {
	Role_id   string `gorm:"type:varchar(10);comment:'角色编号';primaryKey" json:"role_id"`
	Role_name string `gorm:"type:varchar(50);comment:'角色名称'" json:"role_name"`
}

type Hawthorn_sys_user struct {
	User_id       string `gorm:"type:varchar(20);comment:'用户编号';primaryKey" json:"user_id"`
	User_name     string `gorm:"type:varchar(50);comment:'用户名称'" json:"user_name"`
	User_password string `gorm:"type:varchar(50);comment:'用户密码'" json:"user_password"`
}
