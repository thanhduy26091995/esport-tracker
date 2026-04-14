package model

type Config struct {
	Key         string `gorm:"primary_key;type:varchar(50)" json:"key"`
	Value       string `gorm:"type:text;not null" json:"value"`
	Description string `gorm:"type:text" json:"description"`
}

func (Config) TableName() string {
	return "config"
}
