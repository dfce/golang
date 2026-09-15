# 在 gorm 中创建临时表
  因为 GORM 默认使用 连接池，第二个操作可能跑到另一个连接上。 会抛错relation "tmp_game_score" does not exist
```go
// 1. 使用同一个事务
// CREATE TEMP TABLE tmp_table (...) ON COMMIT [ DELETE ROWS | DROP | PRESERVE ROWS ];
// ON COMMIT 的三种模式:
    // DELETE ROWS	    每次事务提交时清空表（TRUNCATE）。
    // DROP	            每次事务提交时直接删除表。
    // PRESERVE ROWS	  默认行为，事务提交后数据仍保留，直到连接关闭。

createSql := fmt.Sprintf(`
    CREATE TEMP TABLE %s (
				id BIGINT PRIMARY KEY,
				score NUMERIC(12,6),
				is_like BOOLEAN
			) ON COMMIT DROP;
    CREATE INDEX idx_tmp_game_score_id ON tmp_game_score (id);   
`, "tmp_game_score")
tx := db.Session(&gorm.Session{NewDB: true}).Begin()
defer tx.Rollback()

tx.Exec(createSQL)
tx.Exec("执行其他逻辑")

tx.Commit()
```

```sql
CREATE TABLE IF NOT EXISTS tmp_game_score_1 (
  id BIGINT PRIMARY KEY,
  score NUMERIC(12,6),
  is_like BOOLEAN
);
TRUNCATE TABLE tmp_game_score_1;
```