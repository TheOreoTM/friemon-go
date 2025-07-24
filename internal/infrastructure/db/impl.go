package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/core/game"
)

var _ Store = (*DB)(nil)

func (db *DB) DeleteEverything(ctx context.Context) error {
	tx := db.WithContext(ctx)

	if err := tx.Delete(&Character{}, "1=1").Error; err != nil {
		return err
	}

	if err := tx.Delete(&User{}, "1=1").Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateUser(ctx context.Context, user game.User) (*game.User, error) {
	dbUser := modelUserToDBUser(user)

	result := db.WithContext(ctx).Save(&dbUser)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbUserToModelUser(dbUser), nil
}

func (db *DB) GetSelectedCharacter(ctx context.Context, id snowflake.ID) (*game.Character, error) {
	var user User
	result := db.WithContext(ctx).Preload("SelectedCharacter").First(&user, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %s not found", id.String())
		}
		return nil, result.Error
	}

	// Check if the user has a selected character
	if user.SelectedCharacter == nil {
		return nil, fmt.Errorf("user with ID %s has no selected character", id.String())
	}

	return dbCharToModelChar(*user.SelectedCharacter), nil
}

func (db *DB) CreateUser(ctx context.Context, id snowflake.ID) (*game.User, error) {
	dbUser := User{
		ID:            id.String(),
		Balance:       0,
		OrderBy:       0,
		OrderDesc:     false,
		ShiniesCaught: 0,
		NextIdx:       1,
		ELO:           1000, // Default ELO value
	}

	result := db.WithContext(ctx).Create(&dbUser)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbUserToModelUser(dbUser), nil
}

func (db *DB) GetUser(ctx context.Context, id snowflake.ID) (*game.User, error) {
	var user User
	result := db.WithContext(ctx).First(&user, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return dbUserToModelUser(user), nil
}

func (db *DB) EnsureUser(ctx context.Context, id snowflake.ID) (*game.User, error) {
	var user User
	result := db.WithContext(ctx).First(&user, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return db.CreateUser(ctx, id)
		}
		return nil, result.Error
	}

	return dbUserToModelUser(user), nil
}

func (db *DB) DeleteCharacter(ctx context.Context, id uuid.UUID) (*game.Character, error) {
	var character Character
	result := db.WithContext(ctx).First(&character, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	if err := db.WithContext(ctx).Delete(&character).Error; err != nil {
		return nil, err
	}

	return dbCharToModelChar(character), nil
}

func (db *DB) UpdateCharacter(ctx context.Context, id uuid.UUID, ch *game.Character) (*game.Character, error) {
	dbChar := modelCharToDBChar(ch)
	dbChar.ID = id

	result := db.WithContext(ctx).Save(&dbChar)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbCharToModelChar(dbChar), nil
}

func (db *DB) GetCharacter(ctx context.Context, id uuid.UUID) (*game.Character, error) {
	var character Character
	result := db.WithContext(ctx).First(&character, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return dbCharToModelChar(character), nil
}

func (db *DB) CreateCharacter(ctx context.Context, ownerID snowflake.ID, char *game.Character) (Character, error) {
	// Get the user's next idx
	var user User
	result := db.WithContext(ctx).First(&user, "id = ?", ownerID.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Character{}, fmt.Errorf("user with ID %s not found", ownerID.String())
		}
		return Character{}, result.Error
	}

	// Ensure NextIdx is valid
	if user.NextIdx <= 0 {
		return Character{}, fmt.Errorf("invalid NextIdx for user with ID %s", ownerID.String())
	}

	// Assign IDX and OwnerID to the character
	char.IDX = int(user.NextIdx)
	char.OwnerID = ownerID.String()

	dbChar := modelCharToDBChar(char)

	// Start a transaction
	tx := db.WithContext(ctx).Begin()

	// Create the character
	if err := tx.Create(&dbChar).Error; err != nil {
		tx.Rollback()
		return Character{}, err
	}

	// Update user's next idx
	if err := tx.Model(&user).Update("next_idx", user.NextIdx+1).Error; err != nil {
		tx.Rollback()
		return Character{}, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return Character{}, err
	}

	// Assign the generated ID back to the character
	char.ID = dbChar.ID
	return dbChar, nil
}

func (db *DB) GetCharactersForUser(ctx context.Context, userID snowflake.ID) ([]game.Character, error) {
	var characters []Character
	err := db.WithContext(ctx).Where("owner_id = ?", userID).Find(&characters).Error

	modelChars := make([]game.Character, len(characters))
	for i, char := range characters {
		modelChars[i] = *dbCharToModelChar(char)
	}

	return modelChars, err
}

func (db *DB) Tx(ctx context.Context, fn func(Store) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDB := &DB{DB: tx}
		return fn(txDB)
	})
}

// Conversion functions
func dbUserToModelUser(dbUser User) *game.User {
	orderBy := game.OrderBy(dbUser.OrderBy)

	return &game.User{
		ID:         snowflake.MustParse(dbUser.ID),
		Balance:    int(dbUser.Balance),
		SelectedID: dbUser.SelectedID,
		Order: game.OrderOptions{
			OrderBy: orderBy,
			Desc:    dbUser.OrderDesc,
		},
		NextIdx:       int(dbUser.NextIdx),
		ShiniesCaught: int(dbUser.ShiniesCaught),
	}
}

func modelUserToDBUser(user game.User) User {
	return User{
		ID:            user.ID.String(),
		Balance:       int32(user.Balance),
		SelectedID:    user.SelectedID,
		OrderBy:       int32(user.Order.OrderBy),
		OrderDesc:     user.Order.Desc,
		NextIdx:       int32(user.NextIdx),
		ShiniesCaught: int32(user.ShiniesCaught),
	}
}

func modelCharToDBChar(ch *game.Character) Character {
	return Character{
		ID:               ch.ID,
		OwnerID:          ch.OwnerID,
		ClaimedTimestamp: ch.ClaimedTimestamp,
		IDX:              int32(ch.IDX),
		CharacterID:      int32(ch.CharacterID),
		Level:            int32(ch.Level),
		XP:               int32(ch.XP),
		Personality:      ch.Personality.String(),
		Shiny:            ch.Shiny,
		IvHP:             int32(ch.IvHP),
		IvAtk:            int32(ch.IvAtk),
		IvDef:            int32(ch.IvDef),
		IvSpAtk:          int32(ch.IvSpAtk),
		IvSpDef:          int32(ch.IvSpDef),
		IvSpd:            int32(ch.IvSpd),
		IvTotal:          ch.IvTotal,
		Nickname:         ch.Nickname,
		Favourite:        ch.Favourite,
		HeldItem:         int32(ch.HeldItem),
		Moves:            ch.Moves,
		Color:            ch.Color,
	}
}

func dbCharToModelChar(dbch Character) *game.Character {
	return &game.Character{
		ID:               dbch.ID,
		OwnerID:          dbch.OwnerID,
		ClaimedTimestamp: dbch.ClaimedTimestamp,
		IDX:              int(dbch.IDX),
		CharacterID:      int(dbch.CharacterID),
		Level:            int(dbch.Level),
		XP:               int(dbch.XP),
		Personality:      stringToPersonality(dbch.Personality),
		Shiny:            dbch.Shiny,
		IvHP:             int(dbch.IvHP),
		IvAtk:            int(dbch.IvAtk),
		IvDef:            int(dbch.IvDef),
		IvSpAtk:          int(dbch.IvSpAtk),
		IvSpDef:          int(dbch.IvSpDef),
		IvSpd:            int(dbch.IvSpd),
		IvTotal:          dbch.IvTotal,
		Nickname:         dbch.Nickname,
		Favourite:        dbch.Favourite,
		HeldItem:         int(dbch.HeldItem),
		Moves:            dbch.Moves,
		Color:            dbch.Color,
	}
}

func stringToPersonality(s string) constants.Personality {
	switch s {
	case "Aloof":
		return constants.PersonalityAloof
	case "Stoic":
		return constants.PersonalityStoic
	case "Merry":
		return constants.PersonalityMerry
	case "Resolute":
		return constants.PersonalityResolute
	case "Skeptical":
		return constants.PersonalitySkeptical
	case "Brooding":
		return constants.PersonalityBrooding
	case "Brave":
		return constants.PersonalityBrave
	case "Insightful":
		return constants.PersonalityInsightful
	case "Playful":
		return constants.PersonalityPlayful
	case "Rash":
		return constants.PersonalityRash
	default:
		return constants.PersonalityAloof
	}
}
