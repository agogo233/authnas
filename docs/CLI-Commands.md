# CLI Commands | CLI 命令

## 访问方式

```bash
docker compose run authnas <command> [options]
```

## 命令列表

### serve

启动 AuthNas 服务（默认命令）。

```bash
docker compose run authnas serve
```

### migrate

数据库迁移，详见 [数据库迁移](DB-Migration.md)。

### generate password-reset

为用户生成密码重置链接。

```bash
docker compose run authnas generate password-reset <username>
```
