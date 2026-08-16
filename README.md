# power-market

电力市场竞价模拟与出清价格分析（single-period uniform-price auction）。

程序读入发电机报价 CSV，按报价从低到高排列成机组顺序（merit order），对给定的系统负荷
（demand）进行市场出清，出清价为最后一台中标机组（边际机组）的报价，并输出结算分析：
各发电商中标电量、未满足负荷、用户成本与容量利用率。

## 出清规则

- 按价格升序（同价保持输入顺序）依次调用容量，直到满足负荷。
- 出清价 = 边际机组报价，所有中标电量按该统一价格结算。
- 总供给 >= 负荷：最后一台机组可能部分中标（容量被削减为剩余负荷），未满足负荷为 0。
- 总供给 < 负荷：全部机组满发中标，出清价为最高报价，未满足负荷 = 负荷 - 总供给。
- 容量非正的报价不会中标；负荷 <= 0 时不出清。

## 目录结构

- `main.go` — 命令行入口
- `internal/bid` — CSV 解析（`ParseBids`）与机组排序（`MeritOrder`）
- `internal/clear` — 市场出清（`Clear`、`TotalSupply`）
- `internal/analyze` — 结算分析（`AwardedByGen`、`ConsumerCost`、`Utilization`、`GeneratorNames`）
- `example/bids.csv` — 示例报价文件

## 输入格式

CSV 首行必须是表头 `generator,kind,capacity,price`；`capacity` 单位 MW，`price` 单位 $/MWh，
两者均须为非负有限数值。示例见 `example/bids.csv`（核电、水电、煤电、燃气、尖峰机组共 7 条，
总供给 960 MW）。

## 命令行用法

```
power-market -bids <path> -demand <float>
```

- `-bids` 报价 CSV 路径（必填）
- `-demand` 系统负荷 MW，必须大于 0（必填）

示例：

```
go run . -bids example/bids.csv -demand 500    # 出清价 31.75 $/MWh，未满足 0 MW
go run . -bids example/bids.csv -demand 1200   # 总供给不足，出清价为最高报价，未满足 240 MW
```

## 退出码

- `0` 出清成功并打印报告
- `1` 输入文件缺失或格式错误（打印错误信息，不 panic）
- `2` 缺少 `-bids`、缺少 `-demand` 或 `-demand <= 0`（打印用法到 stderr）

## 构建与测试

```
export GOTOOLCHAIN=local CGO_ENABLED=0
go vet ./...
go build ./...
go test ./...
```

Docker：

```
docker build -t power-market .
```
