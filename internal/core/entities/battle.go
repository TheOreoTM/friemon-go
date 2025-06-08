package entities

import (
	"math"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
)

type BattleStatus string

const (
	BattleStatusPending   BattleStatus = "pending"
	BattleStatusActive    BattleStatus = "active"
	BattleStatusCompleted BattleStatus = "completed"
	BattleStatusCancelled BattleStatus = "cancelled"
)

type ActionType string

const (
	ActionMove    ActionType = "move"
	ActionSwitch  ActionType = "switch"
	ActionForfeit ActionType = "forfeit"
)

type Battle struct {
	ID                 uuid.UUID     `json:"id"`
	ChallengerID       snowflake.ID  `json:"challenger_id"`
	OpponentID         snowflake.ID  `json:"opponent_id"`
	WinnerID           *snowflake.ID `json:"winner_id,omitempty"`
	Status             BattleStatus  `json:"status"`
	TurnCount          int           `json:"turn_count"`
	CurrentTurnPlayer  *snowflake.ID `json:"current_turn_player,omitempty"`
	MainThreadID       *snowflake.ID `json:"main_thread_id,omitempty"`
	ChallengerThreadID *snowflake.ID `json:"challenger_thread_id,omitempty"`
	OpponentThreadID   *snowflake.ID `json:"opponent_thread_id,omitempty"`
	Settings           GameSettings  `json:"battle_settings"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	CompletedAt        *time.Time    `json:"completed_at,omitempty"`

	// Runtime data (not stored in DB)
	ChallengerTeam []*BattleTeamMember    `json:"-"`
	OpponentTeam   []*BattleTeamMember    `json:"-"`
	TurnQueue      []BattleTurnAction     `json:"-"`
	WeatherEffect  *WeatherEffect         `json:"-"`
	FieldEffects   map[string]interface{} `json:"-"`
}

type BattleTeamMember struct {
	ID            uuid.UUID                `json:"id"`
	BattleID      uuid.UUID                `json:"battle_id"`
	UserID        snowflake.ID             `json:"user_id"`
	TeamPosition  int                      `json:"team_position"`
	CharacterID   uuid.UUID                `json:"character_id"`
	CharacterData Character                `json:"character_data"`
	CurrentHP     int                      `json:"current_hp"`
	StatusEffects []constants.StatusEffect `json:"status_effects"`
	StatStages    map[string]int           `json:"stat_stages"`
	IsActive      bool                     `json:"is_active"`
	IsFainted     bool                     `json:"is_fainted"`
	TempModifiers map[string]int           `json:"temp_modifiers"`
	LastMoveUsed  *int                     `json:"last_move_used,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
}

type BattleTurn struct {
	ID         uuid.UUID              `json:"id"`
	BattleID   uuid.UUID              `json:"battle_id"`
	TurnNumber int                    `json:"turn_number"`
	UserID     snowflake.ID           `json:"user_id"`
	ActionType ActionType             `json:"action_type"`
	ActionData map[string]interface{} `json:"action_data"`
	ResultData map[string]interface{} `json:"result_data"`
	CreatedAt  time.Time              `json:"created_at"`
}

type BattleTurnAction struct {
	UserID     snowflake.ID           `json:"user_id"`
	ActionType ActionType             `json:"action_type"`
	ActionData map[string]interface{} `json:"action_data"`
	Priority   int                    `json:"priority"`
	Character  *BattleTeamMember      `json:"character"`
}

type GameSettings struct {
	TurnLimit          int  `json:"turn_limit"`
	TurnTimeoutSeconds int  `json:"turn_timeout_seconds"`
	SwitchCostsTurn    bool `json:"switch_costs_turn"`
	MaxTeamSize        int  `json:"max_team_size"`
}

type WeatherEffect struct {
	Type      string `json:"type"`
	Duration  int    `json:"duration"`
	Intensity int    `json:"intensity"`
}

type UserElo struct {
	UserID       snowflake.ID `json:"user_id"`
	EloRating    int          `json:"elo_rating"`
	BattlesWon   int          `json:"battles_won"`
	BattlesLost  int          `json:"battles_lost"`
	BattlesTotal int          `json:"battles_total"`
	HighestElo   int          `json:"highest_elo"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Battle methods
func NewBattle(challengerID, opponentID snowflake.ID, settings GameSettings) *Battle {
	return &Battle{
		ID:             uuid.New(),
		ChallengerID:   challengerID,
		OpponentID:     opponentID,
		Status:         BattleStatusPending,
		TurnCount:      0,
		Settings:       settings,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ChallengerTeam: make([]*BattleTeamMember, 0, settings.MaxTeamSize),
		OpponentTeam:   make([]*BattleTeamMember, 0, settings.MaxTeamSize),
		TurnQueue:      make([]BattleTurnAction, 0),
		FieldEffects:   make(map[string]interface{}),
	}
}

func (b *Battle) GetTeam(userID snowflake.ID) []*BattleTeamMember {
	if userID == b.ChallengerID {
		return b.ChallengerTeam
	}
	return b.OpponentTeam
}

func (b *Battle) GetActiveCharacter(userID snowflake.ID) *BattleTeamMember {
	team := b.GetTeam(userID)
	for _, member := range team {
		if member.IsActive && !member.IsFainted {
			return member
		}
	}
	return nil
}

func (b *Battle) GetOpponentID(userID snowflake.ID) snowflake.ID {
	if userID == b.ChallengerID {
		return b.OpponentID
	}
	return b.ChallengerID
}

func (b *Battle) IsPlayersTurn(userID snowflake.ID) bool {
	return b.CurrentTurnPlayer != nil && *b.CurrentTurnPlayer == userID
}

func (b *Battle) HasTeamFainted(userID snowflake.ID) bool {
	team := b.GetTeam(userID)
	for _, member := range team {
		if !member.IsFainted {
			return false
		}
	}
	return true
}

func (b *Battle) IsComplete() bool {
	return b.Status == BattleStatusCompleted || b.Status == BattleStatusCancelled
}

// BattleTeamMember methods
func NewBattleTeamMember(battleID uuid.UUID, userID snowflake.ID, position int, character *Character) *BattleTeamMember {
	return &BattleTeamMember{
		ID:            uuid.New(),
		BattleID:      battleID,
		UserID:        userID,
		TeamPosition:  position,
		CharacterID:   character.ID,
		CharacterData: *character,
		CurrentHP:     character.MaxHP(),
		StatusEffects: make([]constants.StatusEffect, 0),
		StatStages:    make(map[string]int),
		IsActive:      position == 1, // First character is active by default
		IsFainted:     false,
		TempModifiers: make(map[string]int),
		CreatedAt:     time.Now(),
	}
}

func (btm *BattleTeamMember) ApplyDamage(damage int) {
	btm.CurrentHP = max(0, btm.CurrentHP-damage)
	if btm.CurrentHP == 0 {
		btm.IsFainted = true
		btm.IsActive = false
	}
}

func (btm *BattleTeamMember) Heal(amount int) {
	maxHP := btm.CharacterData.MaxHP()
	btm.CurrentHP = min(maxHP, btm.CurrentHP+amount)
	if btm.CurrentHP > 0 {
		btm.IsFainted = false
	}
}

func (btm *BattleTeamMember) AddStatusEffect(effect constants.StatusEffect) {
	for _, existing := range btm.StatusEffects {
		if existing == effect {
			return // Already has this status
		}
	}
	btm.StatusEffects = append(btm.StatusEffects, effect)
}

func (btm *BattleTeamMember) RemoveStatusEffect(effect constants.StatusEffect) {
	for i, existing := range btm.StatusEffects {
		if existing == effect {
			btm.StatusEffects = append(btm.StatusEffects[:i], btm.StatusEffects[i+1:]...)
			break
		}
	}
}

func (btm *BattleTeamMember) HasStatusEffect(effect constants.StatusEffect) bool {
	for _, existing := range btm.StatusEffects {
		if existing == effect {
			return true
		}
	}
	return false
}

func (btm *BattleTeamMember) ModifyStat(stat string, stages int) {
	if btm.StatStages == nil {
		btm.StatStages = make(map[string]int)
	}
	btm.StatStages[stat] = max(-6, min(6, btm.StatStages[stat]+stages))
}

func (btm *BattleTeamMember) GetStatStage(stat string) int {
	if btm.StatStages == nil {
		return 0
	}
	return btm.StatStages[stat]
}

func (btm *BattleTeamMember) CalculateEffectiveStat(stat string) int {
	baseStat := btm.getBaseStat(stat)
	stage := btm.GetStatStage(stat)

	multiplier := 1.0
	if stage > 0 {
		multiplier = 1.0 + (float64(stage) * 0.5)
	} else if stage < 0 {
		multiplier = 1.0 / (1.0 + (float64(-stage) * 0.5))
	}

	return int(float64(baseStat) * multiplier)
}

func (btm *BattleTeamMember) getBaseStat(stat string) int {
	switch stat {
	case "atk":
		return btm.CharacterData.Atk()
	case "def":
		return btm.CharacterData.Def()
	case "satk":
		return btm.CharacterData.SpAtk()
	case "sdef":
		return btm.CharacterData.SpDef()
	case "spd":
		return btm.CharacterData.Spd()
	default:
		return btm.CharacterData.HP()
	}
}

// ELO calculation functions
func CalculateEloChange(winnerElo, loserElo int) (int, int) {
	const K = 32 // K-factor

	expectedWin := 1.0 / (1.0 + math.Pow(10, float64(loserElo-winnerElo)/400.0))
	expectedLose := 1.0 - expectedWin

	winnerChange := int(K * (1.0 - expectedWin))
	loserChange := int(K * (0.0 - expectedLose))

	return winnerChange, loserChange
}

func GetEloRank(elo int) string {
	switch {
	case elo >= 2000:
		return "Master"
	case elo >= 1800:
		return "Diamond"
	case elo >= 1600:
		return "Platinum"
	case elo >= 1400:
		return "Gold"
	case elo >= 1200:
		return "Silver"
	case elo >= 1000:
		return "Bronze"
	default:
		return "Novice"
	}
}
