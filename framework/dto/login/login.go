package login

type Sys_menu_dto struct {
	Menu_id   int    `gorm:"type:int;comment:'菜单编号';primaryKey" json:"-"`
	Parent_id int    `gorm:"type:int;comment:'父菜单编号'" json:"-"`
	Path      string `gorm:"type:varchar(100);comment:'路由地址'" json:"path"`
	Component string `gorm:"type:varchar(100);comment:'组件路径'" json:"component"`
	Meta      `gorm:"embedded" json:"meta"`
	Children  []*Sys_menu_dto `gorm:"-" json:"children"`
}
type Meta struct {
	Menu_name string `gorm:"type:varchar(50);comment:'菜单名称'" json:"title"`
	Order_num int    `gorm:"type:int;comment:'显示顺序'" json:"order"`
	Is_cache  string `gorm:"type:varchar(2);comment:'是否缓存0-是1-否'" json:"is_cache"`
	Icon      string `gorm:"type:varchar(100);comment:'菜单图标'" json:"icon"`
}
