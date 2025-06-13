package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Character struct {
	gorm.Model
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID          string    `gorm:"type:varchar(255);not null;index" json:"owner_id"`
	ClaimedTimestamp time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"claimed_timestamp"`
	IDX              int32     `gorm:"not null" json:"idx"`
	CharacterID      int32     `gorm:"not null" json:"character_id"`
	Level            int32     `gorm:"not null;default:1" json:"level"`
	XP               int32     `gorm:"not null;default:0" json:"xp"`
	Personality      string    `gorm:"type:varchar(50);not null" json:"personality"`
	Shiny            bool      `gorm:"not null;default:false" json:"shiny"`
	IvHP             int32     `gorm:"not null" json:"iv_hp"`
	IvAtk            int32     `gorm:"not null" json:"iv_atk"`
	IvDef            int32     `gorm:"not null" json:"iv_def"`
	IvSpAtk          int32     `gorm:"not null" json:"iv_sp_atk"`
	IvSpDef          int32     `gorm:"not null" json:"iv_sp_def"`
	IvSpd            int32     `gorm:"not null" json:"iv_spd"`
	IvTotal          float64   `gorm:"not null" json:"iv_total"`
	Nickname         string    `gorm:"type:varchar(255);not null;default:''" json:"nickname"`
	Favourite        bool      `gorm:"not null;default:false" json:"favourite"`
	HeldItem         int32     `gorm:"not null;default:-1" json:"held_item"`
	Moves            []int32   `gorm:"type:integer[]" json:"moves"`
	Color            int32     `gorm:"not null" json:"color"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Relationships
	Owner *User `gorm:"foreignKey:OwnerID;references:ID" json:"owner,omitempty"`
}

func (Character) TableName() string {
	return "characters"
}

type User struct {
	gorm.Model
	ID            string    `gorm:"type:varchar(255);primaryKey" json:"id"`
	Balance       int32     `gorm:"not null;default:0" json:"balance"`
	SelectedID    uuid.UUID `gorm:"type:uuid;default:null" json:"selected_id"`
	OrderBy       int32     `gorm:"not null;default:0" json:"order_by"`
	OrderDesc     bool      `gorm:"not null;default:false" json:"order_desc"`
	ShiniesCaught int32     `gorm:"not null;default:0" json:"shinies_caught"`
	NextIdx       int32     `gorm:"not null;default:1" json:"next_idx"`
	ELO           int32     `gorm:"not null;default:1000" json:"elo"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Relationships
	SelectedCharacter *Character  `json:"selected_character,omitempty"`
	Characters        []Character `json:"characters,omitempty"`
}

func (User) TableName() string {
	return "users"
}
