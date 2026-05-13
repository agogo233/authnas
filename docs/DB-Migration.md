# Database Migration | 数据库迁移

## 迁移步骤

1. **停止 AuthNas**
   ```bash
   docker compose rm -s authnas
   ```

2. **准备目标数据库**
   确保目标数据库可连接（PostgreSQL 或 SQLite）

3. **配置环境变量**
   设置 `MIGRATE_TO_DB_*` 变量（目标数据库配置）

4. **执行迁移**
   ```bash
   docker compose run authnas migrate
   ```

5. **更新配置**
   将 `DB_*` 改为目标数据库配置，移除 `MIGRATE_TO_DB_*`

6. **重启服务**
   ```bash
   docker compose up -d authnas
   ```

> [!TIP]
> 迁移对原数据库**非破坏性**，但目标数据库会被覆盖。
