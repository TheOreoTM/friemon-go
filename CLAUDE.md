# Friemon: project guide for Claude

A Pokétwo-style Discord bot where players catch characters from *Frieren: Beyond Journey's End*,
plus a free (no paid currency) Genshin-style relic system. This is a **fresh start**: do not reuse
code from the old repos TheOreoTM/friemon (TypeScript) or TheOreoTM/friemon-go.

## About the developer
- Knows TypeScript, is **learning Go**. Prefer simple, idiomatic Go and briefly explain Go-specific idioms.
- Do not over-engineer. No Redis, job queues or Docker/CI until they are actually needed.
- Runs in WSL; the project lives in the WSL filesystem (not /mnt/c).

## Stack
- Go, disgo (github.com/disgoorg/disgo), with disgoorg/paginator for paged lists
- Postgres on Neon, via pgx (`DATABASE_URL`, the pooled connection string). Migrations are numbered
  `.sql` files in `internal/store/migrations`, applied on startup; never edit one that has already run.
- Store tests use `TEST_DATABASE_URL` (a separate Neon branch) and are skipped when it is unset.
- Static game data (characters, moves, relics, items) lives in **Go code as structs**, not JSON
- Per-player data lives in the database: users, balances, each caught copy (own level, XP, stats, moves), relics, trades, listings

## Layout
```
cmd/bot/main.go     entry point
internal/game/      pure logic: spawn, battle, relics (NO Discord imports, unit-tested)
internal/data/      characters, moves, relics, items
internal/store/     database access
internal/bot/       commands, handlers, cooldowns
```

## Design

### Layer 1: Pokétwo-familiar (new players should feel at home)
- Every message bumps a per-server counter (with per-user and per-server cooldowns); at a threshold
  a random character spawns in a configured channel.
- `catch <name>` guesses the name (case- and accent-insensitive, with aliases). `hint` reveals letters.
- Commands: collection list, select, info, favorite, trade, market, same names as Pokétwo where possible.
- Each caught copy has hidden stat rolls (like IVs), plus a level and XP gained from chatting.
- Rarity tiers: Fern and Stark common, Himmel rare, Frieren legendary.

### Moves
- Characters reference moves by ID. A character's learnset is a list of {level, moveID}.
- On catch: assign up to 4 random moves from those learnable at its starting level.
- `learn <move>`: allowed if the learnset permits and the level is met; if the character already has 4,
  ask which move to replace (select menu).
- Item moves (e.g. a Grimoire) can teach moves outside the learnset to compatible characters.

### Battling
- Goal: Pokémon-level depth. Party vs party (up to 4 each) with switching, turn-based with buttons.
  Battle state is in memory.
- Stats: base + IVs + EVs (earned from battles) + level + nature (+10%/-10%), Pokémon formula.
- Moves are data plus effects: priority, crit stage, flags, multi-hit, recoil, drain, and a list of
  effects with chance and target. Rare one-off moves may get custom code.
- Statuses and field effects use Pokémon mechanics with a mix of names: universal ones keep
  Pokémon names (Burn, Sleep, Freeze, Confusion), others are Frieren-themed.
- Abilities, field effects and relic set bonuses plug into the battle engine through hooks.
- Weather (5 turns): Heatwave (Destruction x1.5, Nature x0.5, no freezing, stronger heals),
  Storm (Nature x1.5, Destruction x0.5, no burns, Lightning Magic never misses),
  Sandstorm (1/16 chip except Martial; Martial x1.5 Magic Defense),
  Blizzard (1/16 chip except Nature; Nature x1.5 Defense; Frost Magic never misses).
- Terrain (5 turns): Mana Field (Destruction x1.3, no sleep), Sacred Ground (Holy x1.3, Curse x0.5,
  blocks priority moves), Overgrowth (Nature x1.3, heals 1/16, ground moves x0.5),
  Mist (no new conditions or confusion, Curse x0.5). No "grounded" rule.
- Side conditions: Physical/Magic Ward (screens), Tailwind, Reversal Field (Trick Room),
  hazards (Mana Mines, Caltrops, Curse Seals). Plus Protect, Substitute, charge and recharge
  moves, trapping, Taunt, Encore and pivot moves.
- Types: Destruction, Barrier, Holy, Martial, Curse, Nature. Moves are Physical (Attack/Defense),
  Magical (Magic Attack/Magic Defense) or Status.

### Relics (free gacha, Genshin-artifact style)
- 5 slots: Staff, Robe, Amulet, Ring, Grimoire. Rarity 1-5 stars, a set, a main stat, 3-4 substats.
- Upgrade to +20 using Mana and fodder; substat rolls at +4/8/12/16/20.
- Set bonuses at 2 and 4 pieces.
- Sources: dungeon runs (main source, with a pity counter for a guaranteed 5 star), and a small
  chance (~3-5%) of a drop on a wild catch, biased toward 1-3 stars; rarer characters raise the odds.
  Pity applies to dungeons only. Wild drops use the same anti-spam caps as catching.
- Mana is earned by playing, with a daily cap (like resin). Regenerate it lazily from timestamps.
- Keep it easy for newcomers: an "auto-equip best" button and a simple relic score.

### Frieren-specific features (later)
- **Journeys:** send a character away for hours; store an `endsAt` timestamp and compute results when checked.
- **Hero Party:** `/party` holds up to 4 characters and is also the battle team. Slot 1 is the
  leader (`/select` sets it). Chatting gives the leader full XP and the others 50%.
  Party bonuses come later.
- **Mana Suppression:** some spawns hide their true rarity until caught.
- **First-Class Mage Exam:** seasonal ranked ladder.

## Build order
1. Bot connects and answers one slash command
2. Spawn + catch + collection
3. Levels, XP, moves
4. Battles
5. Relics + dungeons
6. Journeys, market, seasons

## Rules
- Slash commands are the primary interface (prefix commands need the Message Content intent, which needs approval at 100+ servers).
- No official Frieren art; use text embeds or original art. Character names and descriptions are fine.
- Do not use "Poké" in names (Nintendo trademark).
- Pokétwo's source is GPL-3.0: reimplement ideas, never copy its code.
- Validate all game data (every move ID in a learnset exists) with a test that runs in CI.