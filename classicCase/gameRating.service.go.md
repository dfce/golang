```go
package services

import (
	"context"
	"fmt"
	"game-queue-workers/models"
	"game-queue-workers/pkg/logger"
	"game-queue-workers/pkg/utils"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GameRating struct {
	ServiceBase
	models.GameRating
	ctx context.Context
}

func NewGameRating(ctx context.Context) *GameRating {
	return &GameRating{
		ServiceBase: ServiceBase{},
		ctx:         ctx,
	}
}

// ============= GameBets 聚合结果 =================
type GameBets struct {
	UserId     int64     `json:"user_id"`
	GameId     int64     `json:"game_id"`
	BetAmount  float64   `json:"bet_amount"`
	BetCnt     int64     `json:"bet_cnt"`
	WinAmount  float64   `json:"win_amount"`
	WinCnt     int64     `json:"win_cnt"`
	ActiveDays int64     `json:"active_days"`
	LastBetAt  time.Time `json:"last_bet_at"`
	IsRecent   bool      `json:"is_recent"`
}

// ============= 配置（按需调优） =================
const (
	ShardCount       = 128    // 分片数（ 64/128/256，取决于并发与 DB 连接）
	BatchIDSize      = 100000 // 读取 id range 大小（按机器 I/O 调整）
	ReadConcurrency  = 8      // 并发读取批次数（不要超过 DB 能力）
	UpsertBatchSize  = 1000   // 每个 shard 一次 upsert 的条数（100-2000 之间测试）
	MaxUpsertRetries = 4      // 死锁重试次数
	ShardChanBuf     = 4096   // 每个 shard channel 缓冲（过大占内存，过小容易阻塞）
	QueryRetry       = 2      // 单个 range 查询失败重试次数
)

// ============= 类型 ================
type ShardedProcessor struct {
	ctx    context.Context
	cancel context.CancelFunc

	totalInserted int64 // 全局计数
	totalFailed   int64
	startID       int64
	endID         int64
}

// create shard channels and worker stats
type shardStat struct {
	flushedBatches int64
	flushedRows    int64
	failures       int64
}

func NewShardedProcessor(ctx context.Context, startID, endID int64) *ShardedProcessor {
	cctx, cancel := context.WithCancel(ctx)
	return &ShardedProcessor{
		ctx:     cctx,
		cancel:  cancel,
		startID: startID,
		endID:   endID,
	}
}

/**
1. 按照 user_id % shardCount 把每条聚合的结果分发到固定的 shard
	(同一user的数据总是由同一个worker处理; ⚠️ “大R” 用户会导致该分片处理缓慢; 待优化)

2. 读取数据按照 bet-id-range 批次, 并发读取受 readConcurrency 控制
3. 每个shard用有界 channel 做队列, backpressure 保证数据不丢失
*/

func (g *GameRating) RunSharded() error {
	// ============= 初始值 ================
	config, err := g.getConfig()
	if err != nil {
		return err
	}

	startId := config.LasetBetId + 1
	endId, err := g.getMaxBetId()
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("GameRating run start: %d --->>  %d", startId, endId))

	// ============= create processor ================
	proc := NewShardedProcessor(g.ctx, startId, endId)
	if err = proc.Run(g); err != nil {
		logger.Error("proc.Run error", err)
		return err
	}

	// 全部完成后再统计、计算分数、更新 last id 等
	if err := g.statisGameRating(); err != nil {
		logger.Error("statisGameRating error", err)
		return err
	}

	if err := g.GenerateGameScore(config); err != nil {
		logger.Error("GenerateGameScore error", err)
		return err
	}

	// 更新 last processed id
	if err := g.setLastBetId(config.Id, endId); err != nil {
		logger.Warn("setLastBetId failed", err)
		return err
	}
	return nil

}

// Run 执行分片处理
func (p *ShardedProcessor) Run(gr *GameRating) error {
	logger.Info(fmt.Sprintf("====== sharded run start id %d -> %d, shards=%d ======", p.startID, p.endID, ShardCount))

	shards := make([]chan GameBets, ShardCount)
	stats := make([]*shardStat, ShardCount)

	for i := 0; i < ShardCount; i++ {
		shards[i] = make(chan GameBets, ShardChanBuf)
		stats[i] = &shardStat{}
	}

	// start shard workers
	var wg sync.WaitGroup
	for i := 0; i < ShardCount; i++ {
		wg.Add(1)
		go func(shard int) {
			defer wg.Done()
			p.runShardWorker(gr, shard, shards[shard], stats[shard])
		}(i)

	}

	// produce: iterate id ranges, query aggregated rows, dispathc to shards
	var readWg sync.WaitGroup
	readSem := make(chan struct{}, ReadConcurrency)

	for s := p.startID; s <= p.endID; s += int64(BatchIDSize) {
		e := int64(math.Min(float64(s)+BatchIDSize-1, float64(p.endID)))
		readWg.Add(1)
		readSem <- struct{}{}

		go func(s, e int64) {
			defer readWg.Done()
			defer func() { <-readSem }()

			rangeID := fmt.Sprintf("%d-%d", s, e)
			var rows []GameBets
			var err error

			for attempt := 0; attempt <= QueryRetry; attempt++ {
				rows, err = p.queryAggrateBets(s, e)
				if err == nil {
					break
				}
				logger.Warn(fmt.Sprintf("⚠️⚠️⚠️[RANGE %s] query attempt %d failed: %v", rangeID, attempt+1, err), err)
				time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
			}

			if err != nil {
				logger.Error(fmt.Sprintf("❌☠️❌☠️❌☠️ [RANGE %s] give up after retries: %v", rangeID, err))
				atomic.AddInt64(&p.totalFailed, 1)
				return
			}

			logger.Debug(fmt.Sprintf("[RANGE %s] returned %d aggregated rows", rangeID, len(rows)))

			// dispatch per user shard (will block if shard channels is full— backpressure)
			for _, r := range rows {
				select {
				case shards[int(r.UserId%int64(ShardCount))] <- r:
					// dispathced
				case <-gr.ctx.Done():
					return
				}
			}
		}(s, e)
	}
	// wite readers finish then close shard channels

	readWg.Wait()
	for i := 0; i < ShardCount; i++ {
		close(shards[i])
	}

	// wite workers finish
	wg.Wait()

	// aggrate stats
	var totalRows, totalBatches, totalFails int64
	for i := 0; i < ShardCount; i++ {
		totalRows += atomic.LoadInt64(&stats[i].flushedRows)
		totalBatches += atomic.LoadInt64(&stats[i].flushedBatches)
		totalFails += atomic.LoadInt64(&stats[i].failures)
	}

	logger.Info(fmt.Sprintf("sharded run done: shards=%d totalRows=%d totalBatches=%d shardFailures=%d procFailures=%d",
		ShardCount, totalRows, totalBatches, totalFails, atomic.LoadInt64(&p.totalFailed)))

	return nil
}

func (p *ShardedProcessor) runShardWorker(gr *GameRating, shard int, ch <-chan GameBets, stat *shardStat) {

	logger.Info(fmt.Sprintf("[SHARD %03d] worker started", shard))

	buf := make([]GameBets, 0, UpsertBatchSize)
	flushTicker := time.NewTicker(1 * time.Second)

	defer flushTicker.Stop()

	flush := func() {
		if len(buf) == 0 {
			// logger.Debug("no buf to flush")
			return
		}

		// 合并重复数据
		merged := gr.deduplicateBatch(buf)
		// 排序减少死锁
		sort.Slice(merged, func(i, j int) bool {
			if merged[i].UserId == merged[j].UserId {
				return merged[i].GameId < merged[j].GameId
			}
			return merged[i].UserId < merged[j].UserId
		})

		// try upsert with retries
		if err := p.upsertWithRetry(merged); err != nil {
			logger.Error(fmt.Sprintf("[SHARD %03d] upsert failed: %v", shard, err))
			atomic.AddInt64(&stat.failures, 1)
			atomic.AddInt64(&p.totalFailed, 1)
		} else {
			atomic.AddInt64(&stat.flushedBatches, 1)
			atomic.AddInt64(&stat.flushedRows, int64(len(merged)))
			atomic.AddInt64(&p.totalInserted, int64(len(merged)))
			logger.Debug(fmt.Sprintf("[SHARD %03d] flushed batch rows=%d", shard, len(merged)))
		}
		buf = buf[:0]

	}

	//
	for {
		select {
		case item, ok := <-ch:
			if !ok {
				// channel closed - flush remaining and exit
				flush()
				logger.Info(fmt.Sprintf("[SHARD %03d] worker exiting (channel closed)", shard))
				return
			}
			buf = append(buf, item)
			if len(buf) >= UpsertBatchSize {
				flush()
			}

		case <-flushTicker.C:
			flush()

		case <-gr.ctx.Done():
			logger.Warn(fmt.Sprintf("[SHARD %03d] worker exiting ctx.Done", shard), gr.ctx.Err())
			flush()
			return
		}
	}

}

// 获取 gameRatingConfig
func (g *GameRating) getConfig() (config models.GameRatingConfig, err error) {
	err = db.psql.Use("d5").Where("status = ?", 1).First(&config).Error
	if err != nil {
		logger.Error("GameRating getConfig err ----- > ", err)
		return
	}
	return
}

// 获取当前 gameBet 最后一条数据 id
func (g *GameRating) getMaxBetId() (maxId int64, err error) {
	err = db.psql.Use("d5").Raw("select max(id) from game_bet where status = 1").Scan(&maxId).Error
	// err := db.psql.Use("d5").Table("game_bet").Select("max(id)").Where("status = 1").Scan(&maxId).Error
	if err != nil {
		logger.Error("GameRating getMaxBetId err ----- > ", err)
		return
	}
	return
}

// deduplicateBatch：把 batch 中相同 (user_id,game_id) 合并（sum/count/max lastBet）
func (g *GameRating) deduplicateBatch(batch []GameBets) []GameBets {
	// m := make(map[string]GameBets, len(batch))
	m := make(map[int64]GameBets, len(batch))
	for _, s := range batch {
		// key := fmt.Sprintf("%d_%d", s.UserId, s.GameId)
		key := s.UserId<<32 | s.GameId // 如果 gameId <= 2^32
		if ex, ok := m[key]; ok {
			ex.BetAmount += s.BetAmount
			ex.BetCnt += s.BetCnt
			ex.WinAmount += s.WinAmount
			ex.WinCnt += s.WinCnt
			ex.ActiveDays += s.ActiveDays
			ex.IsRecent = s.IsRecent
			if s.LastBetAt.After(ex.LastBetAt) {
				ex.LastBetAt = s.LastBetAt
			}

			m[key] = ex
		} else {
			m[key] = s
		}
	}

	out := make([]GameBets, 0, len(batch))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func (p *ShardedProcessor) queryAggrateBets(start, end int64) ([]GameBets, error) {
	// sql := `with daily as (
	// 					select
	// 					user_id, game_id, sum(bet_amount) bet_amount,
	// 						count(1) bet_cnt, sum(winning_amount) win_amount,
	// 						sum(case when winning_amount > bet_amount then 1 else 0 end) win_cnt,
	// 						(created_at at time zone 'utc' at time zone '-6')::date bet_date
	// 					from game_bet
	// 					where id between ? and ? and transaction_type = 1 and status = 1
	// 					group by user_id, game_id, bet_date
	// 				)
	// 				select
	// 					user_id, game_id, sum(bet_amount) bet_amount,
	// 					sum(bet_cnt) bet_cnt, sum(win_amount) win_amount,
	// 					sum(win_cnt) win_cnt, count(1) AS active_days,
	// 					max(bet_date) last_bet_at
	// 				from daily
	// 				group by user_id, game_id`

	sql := `with daily as (	
						select 
							user_id, game_id, sum(bet_amount) bet_amount, 
								count(1) bet_cnt, sum(winning_amount) win_amount, 
								sum(case when winning_amount > bet_amount then 1 else 0 end) win_cnt,
								max(created_at) last_bet_at,
												(created_at at time zone 'utc' at time zone '-6')::date bet_date,
								BOOL_OR ((created_at at time zone '-6')::date >= ((now() at time zone '-6') - interval '2 d')::date and 
								(created_at at time zone '-6')::date < (now() at time zone '-6')::date) is_recent
							from game_bet
							where id between ? and ? and transaction_type = 1 and status = 1
							group by user_id, game_id, bet_date
					)
					select
						user_id, game_id, sum(bet_amount) bet_amount,
						sum(bet_cnt) bet_cnt, sum(win_amount) win_amount,
						sum(win_cnt) win_cnt, count(1) AS active_days,
						max(last_bet_at) last_bet_at, is_recent
					from daily
					group by user_id, game_id, is_recent`

	var rows []GameBets
	if err := db.psql.Use("d5").Raw(sql, start, end).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (p *ShardedProcessor) upsertWithRetry(merged []GameBets) error {
	if len(merged) == 0 {
		logger.Debug("no merged to upsert")
		return nil
	}

	//
	now := time.Now().UTC()
	rows := make([]*models.GameRating, 0, len(merged))
	for _, s := range merged {
		rows = append(rows, &models.GameRating{
			UserId:    s.UserId,
			GameId:    s.GameId,
			BetAmount: s.BetAmount,
			BetCnt:    s.BetCnt,
			WinAmount: s.WinAmount,
			WinCnt:    s.WinCnt,
			PlayCnt:   s.ActiveDays,
			LastBetAt: &s.LastBetAt,
			IsRecent:  s.IsRecent,
			CreatedAt: &now,
			UpdatedAt: &now,
			Score:     0.0,
			IsLike:    false,
		})
	}

	var err error
	for attempt := 1; attempt <= MaxUpsertRetries; attempt++ {
		err = db.psql.Use("d5").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "game_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"bet_amount":  gorm.Expr("game_rating.bet_amount + EXCLUDED.bet_amount"),
				"bet_cnt":     gorm.Expr("game_rating.bet_cnt + EXCLUDED.bet_cnt"),
				"win_amount":  gorm.Expr("game_rating.win_amount + EXCLUDED.win_amount"),
				"win_cnt":     gorm.Expr("game_rating.win_cnt + EXCLUDED.win_cnt"),
				"play_cnt":    gorm.Expr("game_rating.play_cnt + EXCLUDED.play_cnt"),
				"last_bet_at": gorm.Expr("GREATEST(game_rating.last_bet_at, EXCLUDED.last_bet_at)"),
				"updated_at":  now,
			}),
		}).Create(&rows).Error

		if err == nil {
			return nil
		}
		// detect deadlock
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "40P01" {
			// random backoff
			sleep := time.Duration(100*attempt+rand.Intn(200)) * time.Millisecond
			logger.Warn(fmt.Sprintf("upsert deadlock, attempt=%d sleep=%v", attempt, sleep))
			time.Sleep(sleep)
			continue
		}
		// other error: return immediately (caller will count)
		return err
	}

	return err
}

// 汇总

func (*GameRating) statisGameRating() (err error) {
	sql := `with bet as (
					select 
					sum(bet_amount) bet_amount, user_id
					from game_rating
					group by user_id
				),
				calc as(
					select 
						r.id, r.user_id, r.game_id, r.score, r.bet_amount, r.last_bet_at,
						b.bet_amount all_bet_amount, u.created_at register_at, 
						case f.status when 1 then true else false end as favourite
					from game_rating r
					left join bet b on b.user_id = r.user_id 
					left join users u on u.id = r.user_id
					left join favourite_games f on f.user_id = r.user_id and f.game_id = r.game_id
				)
				update game_rating a
				set 
					all_bet_amount = calc.all_bet_amount,
					collectioned = calc.favourite, 
					register_at = calc.register_at
				from calc
				where a.id = calc.id`
	return db.psql.Use("d5").Exec(sql).Error
}

// 多线程批量更新
func (*GameRating) GenerateGameScore(conf models.GameRatingConfig) error {
	const (
		batchSize   = 5000 // 读取行数
		workerCount = 8    //  8    // 并发worker数量
	)

	type TmpGameScore struct {
		Id     int64   `gorm:"column:id;primaryKey"`
		Score  float64 `gorm:"type:decimal(16,5);column:score"`
		IsLike bool    `gorm:"column:is_like"`
	}

	dbMain := db.psql.Use("d5")

	// 总条数
	// 获取总数和最大ID
	var total int64
	if err := dbMain.Model(&models.GameRating{}).Count(&total).Error; err != nil {
		return fmt.Errorf("count failed: %w", err)
	}

	var maxID int64
	if err := dbMain.Model(&models.GameRating{}).Select("MAX(id)").Scan(&maxID).Error; err != nil {
		return fmt.Errorf("max id failed: %w", err)
	}

	logger.Info(fmt.Sprintf("Total records: %d, Max ID: %d", total, maxID))

	rowsPerWorker := (maxID / int64(workerCount)) + 1
	wg := sync.WaitGroup{}
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			startID := int64(workerID) * rowsPerWorker
			endID := startID + rowsPerWorker - 1

			tmpTable := fmt.Sprintf("tmp_game_score_%d", workerID)
			dbWorker := db.psql.Use("d5").Session(&gorm.Session{NewDB: true})

			// 创建普通表（非 TEMP）
			createSQL := fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id BIGINT PRIMARY KEY,
					score NUMERIC(12,6),
					is_like BOOLEAN
				);
				TRUNCATE TABLE %s;
			`, tmpTable, tmpTable)

			if err := dbWorker.Exec(createSQL).Error; err != nil {
				errCh <- fmt.Errorf("worker %d create table failed: %w", workerID, err)
				return
			}
			logger.Info(fmt.Sprintf("worker %d processing range %d - %d", workerID, startID, endID))

			var lastID int64 = startID
			for {
				var records []models.GameRating
				err := dbWorker.Model(&models.GameRating{}).
					Where("id BETWEEN ? AND ?", lastID, endID).
					Order("id").
					Limit(batchSize).
					Find(&records).Error

				if err != nil {
					errCh <- fmt.Errorf("worker %d read failed: %w", workerID, err)
					return
				}
				if len(records) == 0 {
					break
				}

				tmpRows := make([]TmpGameScore, 0, len(records))
				for _, r := range records {
					score := calcScore(conf, r)
					tmpRows = append(tmpRows, TmpGameScore{
						Id:     r.Id,
						Score:  score,
						IsLike: score >= conf.AdmitScore && (r.BetCnt >= conf.BetCnt || r.BetAmount >= conf.BetAmount),
					})
				}

				if err := dbWorker.Table(tmpTable).CreateInBatches(tmpRows, batchSize).Error; err != nil {
					errCh <- fmt.Errorf("worker %d insert failed: %w", workerID, err)
					return
				}

				lastID = records[len(records)-1].Id + 1
				if len(records) < batchSize {
					break
				}
			}

			logger.Info(fmt.Sprintf("worker %d finished range %d - %d", workerID, startID, endID))
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	// 主线程合并更新
	logger.Info("Merging temporary tables...")
	unionParts := make([]string, workerCount)
	for i := 0; i < workerCount; i++ {
		unionParts[i] = fmt.Sprintf("SELECT * FROM tmp_game_score_%d", i)
	}
	mergeSQL := fmt.Sprintf(`
		UPDATE game_rating AS gr
		SET score = t.score,
		    is_like = t.is_like,
		    updated_at = NOW()
		FROM (%s) AS t
		WHERE gr.id = t.id;
	`, strings.Join(unionParts, " UNION ALL "))

	if err := dbMain.Exec(mergeSQL).Error; err != nil {
		return fmt.Errorf("merge update failed: %w", err)
	}

	// 清理
	for i := 0; i < workerCount; i++ {
		dbMain.Exec(fmt.Sprintf("DROP TABLE IF EXISTS tmp_game_score_%d;", i))
	}

	logger.Info("✅ GenerateGameScore completed successfully")
	return nil
}

func (g *GameRating) setLastBetId(id, LasetBetId int64) error {
	return db.psql.Use("d5").Where("id = ? ", id).Updates(&models.GameRatingConfig{LasetBetId: LasetBetId}).Error
}

func calcScore(conf models.GameRatingConfig, r models.GameRating) float64 {
	baseScore := float64(0)

	baseScore += utils.TruncDecimal4((r.BetAmount / r.AllBetAmount) * conf.BetRatio)
	registerDays := int64(calcDiffDays(*r.RegisterAt))

	if registerDays > 0 {
		baseScore += utils.TruncDecimal4(float64(r.PlayCnt) / float64(registerDays) * conf.ActivityRatio)
	}

	if r.Collectioned {
		baseScore += utils.TruncDecimal4(conf.Collectioned)
	}

	// reecntDays
	if r.IsRecent {
		baseScore += utils.TruncDecimal4(conf.RecentPlayed)
	}

	return baseScore
}

func calcDiffDays(t time.Time) int {
	loc := time.FixedZone("UTC-6", -6*60*60)

	now := time.Now().In(loc)
	t = t.In(loc)

	tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
	nDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return int(nDay.Sub(tDay).Hours() / 24)
}
```