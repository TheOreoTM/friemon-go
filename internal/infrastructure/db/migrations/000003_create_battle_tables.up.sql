-- Battle settings table
CREATE TABLE game_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(255) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO game_settings (setting_key, setting_value) VALUES
('battle_turn_limit', '25'),
('battle_turn_timeout_seconds', '60'),
('battle_switch_costs_turn', 'false'),
('battle_max_team_size', '3');

-- User ELO ratings
CREATE TABLE user_elo (
    user_id VARCHAR(255) PRIMARY KEY,
    elo_rating INT NOT NULL DEFAULT 1000,
    battles_won INT NOT NULL DEFAULT 0,
    battles_lost INT NOT NULL DEFAULT 0,
    battles_total INT NOT NULL DEFAULT 0,
    highest_elo INT NOT NULL DEFAULT 1000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Battles table
CREATE TABLE battles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenger_id VARCHAR(255) NOT NULL,
    opponent_id VARCHAR(255) NOT NULL,
    winner_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, active, completed, cancelled
    turn_count INT NOT NULL DEFAULT 0,
    current_turn_player VARCHAR(255),
    main_thread_id VARCHAR(255),
    challenger_thread_id VARCHAR(255),
    opponent_thread_id VARCHAR(255),
    battle_settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ
);

-- Battle teams - stores the teams used in each battle
CREATE TABLE battle_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    battle_id UUID NOT NULL REFERENCES battles(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    team_position INT NOT NULL, -- 1, 2, or 3
    character_id UUID NOT NULL,
    character_data JSONB NOT NULL, -- snapshot of character at battle time
    current_hp INT NOT NULL,
    status_effects VARCHAR(255)[] DEFAULT '{}',
    stat_stages JSONB NOT NULL DEFAULT '{}', -- stat stage modifications
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_fainted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Battle turns - stores each action taken
CREATE TABLE battle_turns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    battle_id UUID NOT NULL REFERENCES battles(id) ON DELETE CASCADE,
    turn_number INT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action_type VARCHAR(50) NOT NULL, -- move, switch, forfeit
    action_data JSONB NOT NULL DEFAULT '{}',
    result_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_battles_status ON battles(status);
CREATE INDEX idx_battles_participants ON battles(challenger_id, opponent_id);
CREATE INDEX idx_battle_teams_battle_user ON battle_teams(battle_id, user_id);
CREATE INDEX idx_battle_turns_battle ON battle_turns(battle_id, turn_number);