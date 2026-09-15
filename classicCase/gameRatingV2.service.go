package services

import (
	"context"
	"database/sql"
	"fmt"
	"game-queue-workers/models"
	"game-queue-workers/pkg/logger"
	"math"
	"time"

	"gorm.io/gorm"
)

// update game_rating_config set last_bet_id = 0 , laset_exec_at = null;
// truncate table game_rating restart IDENTITY;
//
//	drop table game_rating_day_v2;
//	drop table game_rating_batch_v2;
//	drop table game_rating_dirty_user_v2;
//
// GameRatingV2 使用“按天明细 + 最终汇总表”的方式重构游戏评分任务。
//
// 设计目标：
// 1. 首轮全量时，允许直接扫描 game_bet 的海量数据并分批汇总。
// 2. 增量阶段继续按 id 水位推进，只处理上次成功之后的新数据。
// 3. 游戏天数（play_cnt）必须按自然日去重，不能因为批次切分或任务重跑而重复累计。
// 4. 批次级别幂等：任务中途失败后，下一次重跑不会把已经成功提交的批次再次累加。
// 5. 评分重算尽量走 SQL 集合计算，避免把大量 game_rating 记录拉回 Go 进程。
//
// 关键辅助表：
//
//   - game_rating_day_v2
//     按 (user_id, game_id, bet_date) 唯一存放用户在某游戏某一天的聚合结果。
//     这张表保证“游戏天数”天然幂等。
//
//   - game_rating_batch_v2
//     记录已经成功处理的 id range 批次。任务失败后重跑时会自动跳过已完成批次。
//
//   - game_rating_dirty_user_v2
//     记录本轮受影响的 user_id。批量更新评分时只重算这些用户，避免全表扫描。
//
// 注意：
// - 这里假定 game_rating 表已经存在 (user_id, game_id) 唯一约束；旧逻辑的 upsert 也依赖该约束。
// - 时区延续旧逻辑，按 UTC-6 计算 bet_date / recent / register_days。
type GameRatingV2 struct {
	ServiceBase
	ctx context.Context
}

func NewGameRatingV2(ctx context.Context) *GameRatingV2 {
	return &GameRatingV2{
		ServiceBase: ServiceBase{},
		ctx:         ctx,
	}
}

const (
	// 每批扫描的 game_bet id 范围。
	// 千万级表上不建议一次吞整表，分批能显著降低长事务与排序压力。
	GameRatingV2BatchSize int64 = 500000

	// advisory lock key，保证同一时刻只有一个实例在跑 V2。
	// 这里使用固定 bigint key 即可。
	gameRatingV2AdvisoryLockKey int64 = 2026040201
)

// Run 执行 V2 版本的游戏评分任务。
func (g *GameRatingV2) Run() error {
	dbMain := db.psql.Use("d5")

	unlock, err := g.acquireRunLock(dbMain)
	if err != nil {
		return err
	}
	defer unlock()

	if err := g.ensureSupportTables(dbMain); err != nil {
		return err
	}

	config, err := g.getConfig()
	if err != nil {
		return err
	}

	// 未到配置间隔时直接跳过。
	if calcDiffDays(config.LasetExecAt) < config.Interval {
		logger.Info(fmt.Sprintf("GenerateGameRatingV2 skip, last execAt: %s, interval=%d", config.LasetExecAt, config.Interval))
		return nil
	}

	endID, err := g.getMaxBetId()
	if err != nil {
		return err
	}

	startID := config.LasetBetId + 1
	logger.Info(fmt.Sprintf("GenerateGameRatingV2 window: %d -> %d", startID, endID))

	// 首轮全量重建时，从干净状态开始：
	// - 清空 game_rating，避免在历史脏数据上继续叠加
	// - 清空辅助表，避免旧批次 checkpoint 或旧日明细污染本轮重建结果
	if startID == 1 {
		if err := g.resetForFullRebuild(dbMain); err != nil {
			return err
		}
	}

	// 即便当前没有新 bet，也仍然尝试清理 dirty users，兼容上次批次已成功但评分刷新失败的场景。
	if startID <= endID {
		for batchStart := startID; batchStart <= endID; batchStart += GameRatingV2BatchSize {
			batchEnd := int64(math.Min(float64(batchStart+GameRatingV2BatchSize-1), float64(endID)))
			if err := g.processBatch(dbMain, batchStart, batchEnd); err != nil {
				return err
			}
		}
	} else {
		logger.Info("GenerateGameRatingV2 no new game_bet rows in current window")
	}

	if err := g.refreshDirtyUsers(dbMain, config); err != nil {
		return err
	}

	// 只有全部批次与评分刷新都成功后，才推进主水位。
	if err := g.setLastBetId(config.Id, endID); err != nil {
		return err
	}

	if err := g.cleanupBatchCheckpoint(dbMain, endID); err != nil {
		return err
	}

	return nil
}

// acquireRunLock 通过 PG advisory lock 保证任务单实例运行。
//
// 这里专门保留一个 sql.Conn 持锁，直到任务结束再释放。
// 业务查询仍然走 gorm 的连接池，不要求所有语句都复用同一连接。
func (g *GameRatingV2) acquireRunLock(dbMain *gorm.DB) (func(), error) {
	sqlDB, err := dbMain.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB failed: %w", err)
	}

	conn, err := sqlDB.Conn(g.ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire sql.Conn failed: %w", err)
	}

	var locked bool
	if err := conn.QueryRowContext(g.ctx, "SELECT pg_try_advisory_lock($1)", gameRatingV2AdvisoryLockKey).Scan(&locked); err != nil {
		conn.Close()
		return nil, fmt.Errorf("acquire advisory lock failed: %w", err)
	}

	if !locked {
		conn.Close()
		return nil, fmt.Errorf("game rating v2 job is already running")
	}

	logger.Info("GenerateGameRatingV2 advisory lock acquired")

	return func() {
		if err := releaseAdvisoryLock(g.ctx, conn, gameRatingV2AdvisoryLockKey); err != nil {
			logger.Warn("GenerateGameRatingV2 release advisory lock failed", err)
		}
	}, nil
}

func releaseAdvisoryLock(ctx context.Context, conn *sql.Conn, lockKey int64) error {
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
		return err
	}
	return nil
}

// ensureSupportTables 创建 V2 运行所需的辅助表。
//
// 说明：
// - game_rating_day_v2 是核心幂等表。只要这张表的主键存在，同一 user/game/date 就不会被重复计成多天。
// - game_rating_batch_v2 是批次断点表。批次成功提交后记一条记录，失败回滚则不会留下痕迹。
// - game_rating_dirty_user_v2 用来缩小后续 metadata / score 刷新的影响范围。
func (g *GameRatingV2) ensureSupportTables(dbMain *gorm.DB) error {
	sql := `
CREATE TABLE IF NOT EXISTS game_rating_day_v2 (
	user_id BIGINT NOT NULL,
	game_id BIGINT NOT NULL,
	bet_date DATE NOT NULL,
	bet_amount NUMERIC(20,5) NOT NULL DEFAULT 0,
	bet_cnt BIGINT NOT NULL DEFAULT 0,
	win_amount NUMERIC(20,5) NOT NULL DEFAULT 0,
	win_cnt BIGINT NOT NULL DEFAULT 0,
	last_bet_at TIMESTAMP NULL,
	created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	PRIMARY KEY (user_id, game_id, bet_date)
);

CREATE INDEX IF NOT EXISTS idx_game_rating_day_v2_last_bet_at
	ON game_rating_day_v2 (last_bet_at);

CREATE TABLE IF NOT EXISTS game_rating_batch_v2 (
	batch_start_id BIGINT NOT NULL,
	batch_end_id BIGINT NOT NULL,
	processed_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	PRIMARY KEY (batch_start_id, batch_end_id)
);

CREATE TABLE IF NOT EXISTS game_rating_dirty_user_v2 (
	user_id BIGINT NOT NULL PRIMARY KEY,
	updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
);`

	if err := dbMain.Exec(sql).Error; err != nil {
		return fmt.Errorf("ensure game rating v2 support tables failed: %w", err)
	}
	return nil
}

// resetForFullRebuild 在全量首轮执行前清空主表与辅助表。
//
// 触发条件：
// - game_rating_config.last_bet_id = 0
// - 因而本轮窗口从 id=1 开始
//
// 这样做的目的是把旧版本残留结果、失败批次残留状态全部清掉，
// 保证全量重建一定从一致的初始状态开始。
func (g *GameRatingV2) resetForFullRebuild(dbMain *gorm.DB) error {
	logger.Info("GenerateGameRatingV2 reset tables before full rebuild")

	sql := `
TRUNCATE TABLE game_rating RESTART IDENTITY;
TRUNCATE TABLE game_rating_day_v2;
TRUNCATE TABLE game_rating_batch_v2;
TRUNCATE TABLE game_rating_dirty_user_v2;`

	if err := dbMain.Exec(sql).Error; err != nil {
		return fmt.Errorf("reset tables before full rebuild failed: %w", err)
	}
	return nil
}

// processBatch 处理一个确定的 id 区间。
//
// 幂等保证依赖两个层面：
// 1. 批次成功后在 game_rating_batch_v2 留痕；重跑时先检查，已完成则直接跳过。
// 2. 游戏天数通过 game_rating_day_v2 的 (user_id, game_id, bet_date) 主键保证唯一。
//
// 事务内步骤：
// 1. 把当前 batch 的 game_bet 聚合到临时表 tmp_game_rating_v2_batch（按 user/game/date）。
// 2. 更新已存在的 day 明细。
// 3. 插入本批次第一次出现的新 day，并产出 tmp_game_rating_v2_new_days。
// 4. 将批次增量累计到 game_rating，其中 play_cnt 只累加“新插入的天数”。
// 5. 标记 dirty users，供后续统一重算评分。
// 6. 记录 batch checkpoint。
func (g *GameRatingV2) processBatch(dbMain *gorm.DB, batchStart, batchEnd int64) error {
	done, err := g.batchProcessed(dbMain, batchStart, batchEnd)
	if err != nil {
		return err
	}
	if done {
		logger.Info(fmt.Sprintf("GenerateGameRatingV2 skip processed batch: %d -> %d", batchStart, batchEnd))
		return nil
	}

	logger.Info(fmt.Sprintf("GenerateGameRatingV2 process batch: %d -> %d", batchStart, batchEnd))

	return dbMain.Transaction(func(tx *gorm.DB) error {
		createBatchTempSQL := `
CREATE TEMP TABLE tmp_game_rating_v2_batch ON COMMIT DROP AS
SELECT
	user_id,
	game_id,
	((created_at AT TIME ZONE 'utc') - INTERVAL '6 hour')::date AS bet_date,
	SUM(bet_amount) AS bet_amount,
	COUNT(1) AS bet_cnt,
	SUM(winning_amount) AS win_amount,
	SUM(CASE WHEN winning_amount > bet_amount THEN 1 ELSE 0 END) AS win_cnt,
	MAX(created_at AT TIME ZONE 'utc') AS last_bet_at
FROM game_bet
WHERE id BETWEEN ? AND ?
	AND transaction_type = 1
	AND status = 1
GROUP BY user_id, game_id, bet_date;`
		if err := tx.Exec(createBatchTempSQL, batchStart, batchEnd).Error; err != nil {
			return fmt.Errorf("create tmp batch table failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		// 空批次仍然记 checkpoint，避免之后重复扫描同一段空窗口。
		var rowCount int64
		if err := tx.Raw("SELECT COUNT(1) FROM tmp_game_rating_v2_batch").Scan(&rowCount).Error; err != nil {
			return fmt.Errorf("count tmp batch rows failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		if rowCount == 0 {
			if err := tx.Exec(
				"INSERT INTO game_rating_batch_v2(batch_start_id, batch_end_id, processed_at) VALUES (?, ?, NOW() AT TIME ZONE 'utc')",
				batchStart,
				batchEnd,
			).Error; err != nil {
				return fmt.Errorf("insert empty batch checkpoint failed [%d-%d]: %w", batchStart, batchEnd, err)
			}
			return nil
		}

		updateDaySQL := `
UPDATE game_rating_day_v2 AS d
SET
	bet_amount = d.bet_amount + t.bet_amount,
	bet_cnt = d.bet_cnt + t.bet_cnt,
	win_amount = d.win_amount + t.win_amount,
	win_cnt = d.win_cnt + t.win_cnt,
	last_bet_at = CASE
		WHEN d.last_bet_at IS NULL THEN t.last_bet_at
		ELSE GREATEST(d.last_bet_at, t.last_bet_at)
	END,
	updated_at = NOW() AT TIME ZONE 'utc'
FROM tmp_game_rating_v2_batch AS t
WHERE d.user_id = t.user_id
	AND d.game_id = t.game_id
	AND d.bet_date = t.bet_date;`
		if err := tx.Exec(updateDaySQL).Error; err != nil {
			return fmt.Errorf("update game_rating_day_v2 failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		createNewDayTempSQL := `
CREATE TEMP TABLE tmp_game_rating_v2_new_days ON COMMIT DROP AS
WITH inserted AS (
	INSERT INTO game_rating_day_v2 (
		user_id,
		game_id,
		bet_date,
		bet_amount,
		bet_cnt,
		win_amount,
		win_cnt,
		last_bet_at,
		created_at,
		updated_at
	)
	SELECT
		t.user_id,
		t.game_id,
		t.bet_date,
		t.bet_amount,
		t.bet_cnt,
			t.win_amount,
			t.win_cnt,
			t.last_bet_at,
			NOW() AT TIME ZONE 'utc',
			NOW() AT TIME ZONE 'utc'
		FROM tmp_game_rating_v2_batch AS t
		LEFT JOIN game_rating_day_v2 AS d
		ON d.user_id = t.user_id
		AND d.game_id = t.game_id
		AND d.bet_date = t.bet_date
	WHERE d.user_id IS NULL
	RETURNING user_id, game_id
)
SELECT
	user_id,
	game_id,
	COUNT(1) AS play_cnt_delta
	FROM inserted
	GROUP BY user_id, game_id;`
		if err := tx.Exec(createNewDayTempSQL).Error; err != nil {
			return fmt.Errorf("insert game_rating_day_v2 failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		upsertRatingSQL := `
WITH rating_delta AS (
	SELECT
		user_id,
		game_id,
		SUM(bet_amount) AS bet_amount,
		SUM(bet_cnt) AS bet_cnt,
		SUM(win_amount) AS win_amount,
		SUM(win_cnt) AS win_cnt,
		MAX(last_bet_at) AS last_bet_at
	FROM tmp_game_rating_v2_batch
	GROUP BY user_id, game_id
)
INSERT INTO game_rating (
	user_id,
	game_id,
	score,
	is_like,
	is_recent,
	last_bet_at,
	bet_amount,
	bet_cnt,
	win_amount,
	win_cnt,
	play_cnt,
	all_bet_amount,
	collectioned,
	register_at,
	status,
	created_at,
	updated_at
)
SELECT
	rd.user_id,
	rd.game_id,
	0,
	false,
	false,
	rd.last_bet_at,
	rd.bet_amount,
	rd.bet_cnt,
	rd.win_amount,
	rd.win_cnt,
	COALESCE(nd.play_cnt_delta, 0),
	0,
	false,
	NULL,
	1,
	NOW() AT TIME ZONE 'utc',
	NOW() AT TIME ZONE 'utc'
FROM rating_delta AS rd
LEFT JOIN tmp_game_rating_v2_new_days AS nd
	ON nd.user_id = rd.user_id
	AND nd.game_id = rd.game_id
ON CONFLICT (user_id, game_id) DO UPDATE
SET
	bet_amount = game_rating.bet_amount + EXCLUDED.bet_amount,
	bet_cnt = game_rating.bet_cnt + EXCLUDED.bet_cnt,
	win_amount = game_rating.win_amount + EXCLUDED.win_amount,
	win_cnt = game_rating.win_cnt + EXCLUDED.win_cnt,
	play_cnt = game_rating.play_cnt + EXCLUDED.play_cnt,
	last_bet_at = CASE
		WHEN game_rating.last_bet_at IS NULL THEN EXCLUDED.last_bet_at
		ELSE GREATEST(game_rating.last_bet_at, EXCLUDED.last_bet_at)
	END,
	updated_at = NOW() AT TIME ZONE 'utc';`
		if err := tx.Exec(upsertRatingSQL).Error; err != nil {
			return fmt.Errorf("upsert game_rating failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		markDirtyUsersSQL := `
INSERT INTO game_rating_dirty_user_v2(user_id, updated_at)
SELECT DISTINCT user_id, NOW() AT TIME ZONE 'utc'
FROM tmp_game_rating_v2_batch
ON CONFLICT (user_id) DO UPDATE
SET updated_at = EXCLUDED.updated_at;`
		if err := tx.Exec(markDirtyUsersSQL).Error; err != nil {
			return fmt.Errorf("mark dirty users failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		if err := tx.Exec(
			"INSERT INTO game_rating_batch_v2(batch_start_id, batch_end_id, processed_at) VALUES (?, ?, NOW() AT TIME ZONE 'utc')",
			batchStart,
			batchEnd,
		).Error; err != nil {
			return fmt.Errorf("insert batch checkpoint failed [%d-%d]: %w", batchStart, batchEnd, err)
		}

		return nil
	})
}

func (g *GameRatingV2) batchProcessed(dbMain *gorm.DB, batchStart, batchEnd int64) (bool, error) {
	var count int64
	err := dbMain.Raw(
		"SELECT COUNT(1) FROM game_rating_batch_v2 WHERE batch_start_id = ? AND batch_end_id = ?",
		batchStart,
		batchEnd,
	).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("query batch checkpoint failed [%d-%d]: %w", batchStart, batchEnd, err)
	}
	return count > 0, nil
}

// refreshDirtyUsers 只对受影响用户重算 metadata 与 score。
//
// 为什么按 user 维度刷新：
// - score 中的 bet_amount / all_bet_amount 依赖“用户全量游戏”的投注占比。
// - 只要某个用户任意一个游戏的 bet_amount 增加，该用户所有 game_rating 的分母都可能变化。
//
// 因此 dirty 维度必须至少是 user，而不是单个 (user, game)。
func (g *GameRatingV2) refreshDirtyUsers(dbMain *gorm.DB, conf models.GameRatingConfig) error {
	hasDirty, err := g.hasDirtyUsers(dbMain)
	if err != nil {
		return err
	}
	if !hasDirty {
		logger.Info("GenerateGameRatingV2 no dirty users to refresh")
		return nil
	}

	logger.Info("GenerateGameRatingV2 refresh dirty users score")

	refreshSQL := `
WITH dirty_users AS (
	SELECT user_id
	FROM game_rating_dirty_user_v2
),
user_totals AS (
	SELECT
		gr.user_id,
		SUM(gr.bet_amount) AS all_bet_amount
	FROM game_rating AS gr
	INNER JOIN dirty_users AS du
		ON du.user_id = gr.user_id
	GROUP BY gr.user_id
),
day_counts AS (
	SELECT
		d.user_id,
		d.game_id,
		COUNT(1) AS play_cnt
	FROM game_rating_day_v2 AS d
	INNER JOIN dirty_users AS du
		ON du.user_id = d.user_id
	GROUP BY d.user_id, d.game_id
),
recent_games AS (
	SELECT
		d.user_id,
		d.game_id,
		true AS is_recent
	FROM game_rating_day_v2 AS d
	INNER JOIN dirty_users AS du
		ON du.user_id = d.user_id
	WHERE ((NOW() AT TIME ZONE 'utc' AT TIME ZONE '-6')::date - d.bet_date) BETWEEN 1 AND 2
	GROUP BY d.user_id, d.game_id
),
calc AS (
	SELECT
		gr.id,
		COALESCE(ut.all_bet_amount, 0) AS all_bet_amount,
		COALESCE(dc.play_cnt, 0) AS play_cnt,
		u.created_at AS register_at,
		CASE WHEN fg.status = 1 THEN true ELSE false END AS collectioned,
		COALESCE(rg.is_recent, false) AS is_recent,
		(
			CASE
				WHEN COALESCE(ut.all_bet_amount, 0) > 0
				THEN TRUNC(((gr.bet_amount / ut.all_bet_amount) * ?)::numeric, 4)
				ELSE 0
			END
			+
			CASE
				WHEN u.created_at IS NOT NULL
					AND ((NOW() AT TIME ZONE '-6')::date - (u.created_at AT TIME ZONE 'utc' AT TIME ZONE '-6')::date) > 0
				THEN TRUNC((
					TRUNC((
						COALESCE(dc.play_cnt, 0)::numeric /
						((NOW() AT TIME ZONE '-6')::date - (u.created_at AT TIME ZONE 'utc' AT TIME ZONE '-6')::date)::numeric
					), 4) * ?
				)::numeric, 4)
				ELSE 0
			END
			+
			CASE WHEN fg.status = 1 THEN ? ELSE 0 END
			+
			CASE WHEN COALESCE(rg.is_recent, false) THEN ? ELSE 0 END
		) AS score,
		(
			(
				CASE
					WHEN COALESCE(ut.all_bet_amount, 0) > 0
					THEN TRUNC(((gr.bet_amount / ut.all_bet_amount) * ?)::numeric, 4)
					ELSE 0
				END
				+
				CASE
					WHEN u.created_at IS NOT NULL
						AND ((NOW() AT TIME ZONE '-6')::date - (u.created_at AT TIME ZONE 'utc' AT TIME ZONE '-6')::date) > 0
					THEN TRUNC((
						TRUNC((
							COALESCE(dc.play_cnt, 0)::numeric /
							((NOW() AT TIME ZONE '-6')::date - (u.created_at AT TIME ZONE 'utc' AT TIME ZONE '-6')::date)::numeric
						), 4) * ?
					)::numeric, 4)
					ELSE 0
				END
				+
				CASE WHEN fg.status = 1 THEN ? ELSE 0 END
				+
				CASE WHEN COALESCE(rg.is_recent, false) THEN ? ELSE 0 END
			) >= ?
			AND (gr.bet_cnt >= ? OR gr.bet_amount >= ?)
		) AS is_like
	FROM game_rating AS gr
	INNER JOIN dirty_users AS du
		ON du.user_id = gr.user_id
	LEFT JOIN user_totals AS ut
		ON ut.user_id = gr.user_id
	LEFT JOIN day_counts AS dc
		ON dc.user_id = gr.user_id
		AND dc.game_id = gr.game_id
	LEFT JOIN users AS u
		ON u.id = gr.user_id
	LEFT JOIN favourite_games AS fg
		ON fg.user_id = gr.user_id
		AND fg.game_id = gr.game_id
	LEFT JOIN recent_games AS rg
		ON rg.user_id = gr.user_id
		AND rg.game_id = gr.game_id
)
UPDATE game_rating AS gr
SET
	all_bet_amount = calc.all_bet_amount,
	play_cnt = calc.play_cnt,
	register_at = calc.register_at,
	collectioned = calc.collectioned,
	is_recent = calc.is_recent,
	score = calc.score,
	is_like = calc.is_like,
	updated_at = NOW()
FROM calc
WHERE gr.id = calc.id;`

	args := []interface{}{
		conf.BetRatio,
		conf.ActivityRatio,
		conf.Collectioned,
		conf.RecentPlayed,
		conf.BetRatio,
		conf.ActivityRatio,
		conf.Collectioned,
		conf.RecentPlayed,
		conf.AdmitScore,
		conf.BetCnt,
		conf.BetAmount,
	}
	if err := dbMain.Exec(refreshSQL, args...).Error; err != nil {
		return fmt.Errorf("refresh dirty users score failed: %w", err)
	}

	if err := dbMain.Exec("DELETE FROM game_rating_dirty_user_v2").Error; err != nil {
		return fmt.Errorf("clear dirty users failed: %w", err)
	}

	return nil
}

func (g *GameRatingV2) hasDirtyUsers(dbMain *gorm.DB) (bool, error) {
	var count int64
	if err := dbMain.Raw("SELECT COUNT(1) FROM game_rating_dirty_user_v2").Scan(&count).Error; err != nil {
		return false, fmt.Errorf("query dirty users failed: %w", err)
	}
	return count > 0, nil
}

// cleanupBatchCheckpoint 清理已经被主水位覆盖掉的旧批次记录，避免 checkpoint 表无限膨胀。
func (g *GameRatingV2) cleanupBatchCheckpoint(dbMain *gorm.DB, lastBetID int64) error {
	if err := dbMain.Exec("DELETE FROM game_rating_batch_v2 WHERE batch_end_id <= ?", lastBetID).Error; err != nil {
		return fmt.Errorf("cleanup batch checkpoint failed: %w", err)
	}
	return nil
}

func (g *GameRatingV2) getConfig() (config models.GameRatingConfig, err error) {
	err = db.psql.Use("d5").Where("status = ?", 1).First(&config).Error
	if err != nil {
		return config, fmt.Errorf("get game_rating_config failed: %w", err)
	}
	return config, nil
}

func (g *GameRatingV2) getMaxBetId() (maxID int64, err error) {
	err = db.psql.Use("d5").
		Raw("SELECT COALESCE(MAX(id), 0) FROM game_bet WHERE status = 1").
		Scan(&maxID).Error
	if err != nil {
		return 0, fmt.Errorf("get max game_bet id failed: %w", err)
	}
	return maxID, nil
}

func (g *GameRatingV2) setLastBetId(id, lastBetID int64) error {
	return db.psql.Use("d5").Where("id = ?", id).Updates(&models.GameRatingConfig{
		LasetBetId:  lastBetID,
		LasetExecAt: time.Now(),
	}).Error
}

func calcDiffDays(t time.Time) int {
	loc := time.FixedZone("UTC-6", -6*60*60)

	now := time.Now().In(loc)
	t = t.In(loc)

	tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
	nDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return int(nDay.Sub(tDay).Hours() / 24)
}
