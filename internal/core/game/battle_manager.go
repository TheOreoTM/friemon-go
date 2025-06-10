package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
)

type BattleManager struct {
	battles        map[uuid.UUID]*Battle
	playerBattles  map[snowflake.ID]uuid.UUID // Maps player ID to their current battle ID
	channelBattles map[snowflake.ID]uuid.UUID // Maps channel ID to battle ID
	challenges     map[snowflake.ID]*Challenge
	mutex          sync.RWMutex
}

type Challenge struct {
	ID         uuid.UUID    `json:"id"`
	Challenger snowflake.ID `json:"challenger"`
	Challenged snowflake.ID `json:"challenged"`
	ChannelID  snowflake.ID `json:"channel_id"`
	CreatedAt  time.Time    `json:"created_at"`
	ExpiresAt  time.Time    `json:"expires_at"`
	Settings   GameSettings `json:"settings"`
}

func NewBattleManager() *BattleManager {
	return &BattleManager{
		battles:        make(map[uuid.UUID]*Battle),
		playerBattles:  make(map[snowflake.ID]uuid.UUID),
		channelBattles: make(map[snowflake.ID]uuid.UUID),
		challenges:     make(map[snowflake.ID]*Challenge),
	}
}

func (bm *BattleManager) CreateChallenge(challenger, challenged, channelID snowflake.ID, settings GameSettings) (*Challenge, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Check if challenger is already in a battle
	if _, exists := bm.playerBattles[challenger]; exists {
		return nil, fmt.Errorf("challenger is already in a battle")
	}

	// Check if challenged is already in a battle
	if _, exists := bm.playerBattles[challenged]; exists {
		return nil, fmt.Errorf("challenged player is already in a battle")
	}

	// Check if challenger has already challenged this player
	if existingChallenge, exists := bm.challenges[challenged]; exists {
		if existingChallenge.Challenger == challenger {
			return nil, fmt.Errorf("challenge already pending")
		}
	}

	challenge := &Challenge{
		ID:         uuid.New(),
		Challenger: challenger,
		Challenged: challenged,
		ChannelID:  channelID,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute), // 5 minute expiry
		Settings:   settings,
	}

	bm.challenges[challenged] = challenge
	return challenge, nil
}

func (bm *BattleManager) AcceptChallenge(challenged snowflake.ID) (*Battle, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	challenge, exists := bm.challenges[challenged]
	if !exists {
		return nil, fmt.Errorf("no pending challenge found")
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		delete(bm.challenges, challenged)
		return nil, fmt.Errorf("challenge has expired")
	}

	// Check if challenger is still available
	if _, exists := bm.playerBattles[challenge.Challenger]; exists {
		delete(bm.challenges, challenged)
		return nil, fmt.Errorf("challenger is no longer available")
	}

	// Create battle
	battle := NewBattle(challenge.ChannelID, challenge.Challenger, challenge.Challenged)
	battle.Settings = challenge.Settings
	battle.State = BattleStateTeamSelection

	// Register battle
	bm.battles[battle.ID] = battle
	bm.playerBattles[challenge.Challenger] = battle.ID
	bm.playerBattles[challenged] = battle.ID
	bm.channelBattles[challenge.ChannelID] = battle.ID

	// Remove challenge
	delete(bm.challenges, challenged)

	return battle, nil
}

func (bm *BattleManager) DeclineChallenge(challenged snowflake.ID) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if _, exists := bm.challenges[challenged]; !exists {
		return fmt.Errorf("no pending challenge found")
	}

	delete(bm.challenges, challenged)
	return nil
}

func (bm *BattleManager) GetPlayerBattle(playerID snowflake.ID) (*Battle, bool) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	battleID, exists := bm.playerBattles[playerID]
	if !exists {
		return nil, false
	}

	battle, exists := bm.battles[battleID]
	return battle, exists
}

func (bm *BattleManager) GetChannelBattle(channelID snowflake.ID) (*Battle, bool) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	battleID, exists := bm.channelBattles[channelID]
	if !exists {
		return nil, false
	}

	battle, exists := bm.battles[battleID]
	return battle, exists
}

func (bm *BattleManager) GetBattle(battleID uuid.UUID) (*Battle, bool) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	battle, exists := bm.battles[battleID]
	return battle, exists
}

func (bm *BattleManager) GetChallenge(challenged snowflake.ID) (*Challenge, bool) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	challenge, exists := bm.challenges[challenged]
	return challenge, exists
}

func (bm *BattleManager) AddCharacterToTeam(playerID snowflake.ID, character *Character) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	battleID, exists := bm.playerBattles[playerID]
	if !exists {
		return fmt.Errorf("player is not in a battle")
	}

	battle, exists := bm.battles[battleID]
	if !exists {
		return fmt.Errorf("battle not found")
	}

	if battle.State != BattleStateTeamSelection {
		return fmt.Errorf("battle is not in team selection phase")
	}

	player := battle.GetPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found in battle")
	}

	// Check team size limit
	if len(player.Team) >= battle.Settings.TeamSize {
		return fmt.Errorf("team is already full")
	}

	// Check for duplicates if not allowed
	if !battle.Settings.AllowDuplicates {
		for _, teamChar := range player.Team {
			if teamChar.CharacterID == character.CharacterID {
				return fmt.Errorf("duplicate characters not allowed")
			}
		}
	}

	// Check level cap
	if battle.Settings.LevelCap > 0 && character.Level > battle.Settings.LevelCap {
		return fmt.Errorf("character level exceeds cap of %d", battle.Settings.LevelCap)
	}

	// Add character to team
	charCopy := *character // Create a copy to avoid modifying original
	player.Team = append(player.Team, &charCopy)

	return nil
}

func (bm *BattleManager) StartBattle(battleID uuid.UUID) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	battle, exists := bm.battles[battleID]
	if !exists {
		return fmt.Errorf("battle not found")
	}

	return battle.Start()
}

func (bm *BattleManager) SubmitAction(playerID snowflake.ID, action PlayerAction) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	battleID, exists := bm.playerBattles[playerID]
	if !exists {
		return fmt.Errorf("player is not in a battle")
	}

	battle, exists := bm.battles[battleID]
	if !exists {
		return fmt.Errorf("battle not found")
	}

	return battle.AddAction(playerID, action)
}

func (bm *BattleManager) ProcessBattleTurn(battleID uuid.UUID) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	battle, exists := bm.battles[battleID]
	if !exists {
		return fmt.Errorf("battle not found")
	}

	return battle.ProcessTurn()
}

func (bm *BattleManager) EndBattle(battleID uuid.UUID) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	battle, exists := bm.battles[battleID]
	if !exists {
		return fmt.Errorf("battle not found")
	}

	// Clean up battle references
	delete(bm.playerBattles, battle.Player1.ID)
	delete(bm.playerBattles, battle.Player2.ID)
	delete(bm.channelBattles, battle.ChannelID)

	// Mark battle as finished
	battle.State = BattleStateFinished
	now := time.Now()
	battle.FinishedAt = &now

	// Keep battle in memory for a while for viewing results
	// In production, you might want to save to database here

	return nil
}

func (bm *BattleManager) CleanupExpiredChallenges() {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	now := time.Now()
	for challenged, challenge := range bm.challenges {
		if now.After(challenge.ExpiresAt) {
			delete(bm.challenges, challenged)
		}
	}
}

func (bm *BattleManager) GetActiveBattlesCount() int {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	count := 0
	for _, battle := range bm.battles {
		if battle.State == BattleStateInProgress {
			count++
		}
	}
	return count
}

func (bm *BattleManager) GetPendingChallengesCount() int {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	return len(bm.challenges)
}

// Cleanup finished battles (call periodically)
func (bm *BattleManager) CleanupFinishedBattles(maxAge time.Duration) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	now := time.Now()
	for id, battle := range bm.battles {
		if battle.State == BattleStateFinished && battle.FinishedAt != nil {
			if now.Sub(*battle.FinishedAt) > maxAge {
				delete(bm.battles, id)
			}
		}
	}
}
