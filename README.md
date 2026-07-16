# ip-inspector

一个轻量的 Linux IP 诱捕与自动封禁工具。

ip-inspector 监听指定端口，将连接来源加入 `ipset` 黑名单，并通过 `iptables` 丢弃后续流量。支持同时监听多个端口，封禁记录使用 SQLite 持久化，默认保留 24 小时。

## 特性

- 支持同时监听多个自定义端口
- 基于 `ipset` 与 `iptables` 封禁来源 IP
- SQLite 持久化封禁记录
- 默认封禁 24 小时
- 支持 systemd 守护运行

## 环境要求

- Linux
- root 权限
- iptables 与 ipset 支持
- Go 1.23.4 或更高版本（源码构建）

## 安装

```bash
git clone https://github.com/Rehtt/ip-inspector.git
cd ip-inspector
sudo make install PORTS=3306,6379,8080
```

安装后，程序位于 `/usr/local/sbin/ip-inspector`，数据保存在 `/var/lib/ip-inspector`。
`PORTS` 为必填参数，安装程序会将其写入 systemd 服务配置；未指定或包含非法端口时安装会立即终止。

直接运行时，通过 `-ports` 传入逗号分隔的端口列表：

```bash
sudo ip-inspector -ports 3306,6379,8080
```

端口参数为必填项。重复端口会自动忽略；单个端口绑定失败时会跳过，所有端口均失败时程序退出。

查看运行状态与日志：

```bash
systemctl status ip-inspector
journalctl -u ip-inspector -f
```

## 配置

systemd 服务监听安装时通过 `PORTS` 指定的端口。如需修改，编辑 `/etc/systemd/system/ip-inspector.service` 中的 `-ports` 参数，然后重新加载并重启服务：

```bash
sudo systemctl daemon-reload
sudo systemctl restart ip-inspector
```

封禁时长与 ipset 名称仍在 [`main.go`](main.go) 中定义。

> [!WARNING]
> 程序会修改主机防火墙规则，请在确认规则影响后运行。请确保至少一个监听端口可用。
