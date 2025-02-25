package user

type User struct {
	ID         int64  `gorm:"primaryKey"`
	Username   string `gorm:"column:user_name;unique"`
	Password   string
	History    string         `gorm:"type:text"`
	HistoryMap map[string]int `gorm:"-"`
}
