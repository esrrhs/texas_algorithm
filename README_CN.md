# 德州扑克算法（支持鬼牌）

高性能 Java 德州扑克算法库，支持最多 **2 张鬼牌**（Wild Card），提供两个独立子系统：

| 子系统 | 内存占用 | 功能 |
|--------|----------|------|
| **查表算法** | 约几十 MB | 5–7 张牌的最大牌面评估、大小排序、牌型识别 |
| **胜率评估** | 约 200 MB | 2 张手牌 + 0–4 张公牌的 1v1 胜率估算 |

---

## Maven 依赖

```xml
<dependency>
    <groupId>com.github.esrrhs</groupId>
    <artifactId>texas_algorithm</artifactId>
    <version>1.0.14</version>
</dependency>
```

---

## 快速上手

### 1. 加载数据表

```java
// 加载查表数据（约几十 MB，手牌评估必须先调用）
TexasAlgorithmUtil.load();

// 加载胜率数据（约 200 MB，胜率估算必须先调用）
TexasAlgorithmUtil.loadProbility();
```

### 2. 获取最优 5 张牌

```java
// 从 2 张手牌 + 5 张公牌中找出最大的 5 张牌
String best = TexasAlgorithmUtil.getMax("黑2,黑3", "方2,方A,黑7,黑5,鬼");
```

### 3. 比较牌面大小

```java
// 获取 7 张牌的绝对排名（数值越大越强）
int rank = TexasAlgorithmUtil.getWinPosition("方4,方A,鬼,黑A,黑3,黑5,黑6");

// 直接比较两手 7 张牌：正数表示 str1 赢，0 表示平局，负数表示 str2 赢
int cmp = TexasAlgorithmUtil.compare("方4,方A,鬼,黑A,黑3,黑5,黑6", "黑2,红3,方7,梅9,方K,黑Q,红J");
```

### 4. 获取牌型

```java
// 返回牌型常量（详见下方牌型表）
int type = TexasAlgorithmUtil.getWinType("方4,方A,鬼,黑A,黑3,黑5,黑6");
```

### 5. 估算胜率（1v1）

```java
// 估算 2 张手牌 + 若干公牌情况下的 1v1 胜率
// 返回 [0, 1] 的浮点数
float p = TexasAlgorithmUtil.getHandProbability("方3,鬼", "黑2,黑4,黑5,黑K");
```

---

## 牌面表示

牌面用中文字符串表示，多张牌用英文逗号分隔。

### 花色

| 表示 | 花色 |
|------|------|
| `方` | ♦ 方块 |
| `梅` | ♣ 梅花 |
| `红` | ♥ 红心 |
| `黑` | ♠ 黑桃 |
| `鬼` | 鬼牌（万能替代任意牌） |

### 点数

`2` `3` `4` `5` `6` `7` `8` `9` `10` `J` `Q` `K` `A`

### 示例

```
方A    = 方块 A
黑K    = 黑桃 K
红10   = 红心 10
鬼     = 鬼牌（可替代任意一张牌）
```

---

## 牌型说明

`getWinType()` 返回 `TexasCardUtil` 中的整型常量：

| 常量 | 值 | 牌型 |
|------|----|------|
| `TEXAS_CARD_TYPE_GAOPAI` | 1 | 高牌 |
| `TEXAS_CARD_TYPE_DUIZI` | 2 | 对子 |
| `TEXAS_CARD_TYPE_LIANGDUI` | 3 | 两对 |
| `TEXAS_CARD_TYPE_SANTIAO` | 4 | 三条 |
| `TEXAS_CARD_TYPE_SHUNZI` | 5 | 顺子 |
| `TEXAS_CARD_TYPE_TONGHUA` | 6 | 同花 |
| `TEXAS_CARD_TYPE_HULU` | 7 | 葫芦 |
| `TEXAS_CARD_TYPE_SITIAO` | 8 | 四条 |
| `TEXAS_CARD_TYPE_TONGHUASHUN` | 9 | 同花顺 |
| `TEXAS_CARD_TYPE_KINGTONGHUASHUN` | 10 | 皇家同花顺 |

---

## API 参考

### `TexasAlgorithmUtil`

```java
// 加载 / 初始化
void load()                                   // 加载查表数据到内存（默认当前目录）
void load(String dirPath)                     // 从指定目录加载查表数据
void load(File dir)                           // 从指定目录加载查表数据
void loadProbility()                          // 加载胜率数据到内存（默认当前目录）
void loadProbility(String dirPath)            // 从指定目录加载胜率数据
void loadProbility(File dir)                  // 从指定目录加载胜率数据
boolean isLoaded()                            // 查表数据是否已成功加载
boolean isProbabilityLoaded()                 // 胜率数据是否已成功加载

// 最优手牌
String getMax(String hand, String pub, ...)   // 2 张手牌 + 3–5 张公牌中的最优 5 张
List<Byte> getMax(List<Byte> pokes, ...)      // 同上，接受 byte 列表参数

// 手牌评估（需先调用 load()）
int    getWinPosition(String cards)           // 在所有组合中的绝对排名
double getWinProbability(String cards)        // 相对于同规模全部组合的胜率比例
int    getWinType(String cards)               // 牌型常量
long   getWinMax(String cards)                // 最优 5 张牌的编码 key
int    compare(String str1, String str2)      // 比较两手 7 张牌

// 胜率估算（需先调用 loadProbility()）
float  getHandProbability(String hand, String pub)  // 1v1 胜率估算
```

---

## 运行单元测试

```bash
# 运行 JUnit 5 单元测试（支持 Java 8、11、17、21）
mvn test
```

若需运行完整基准对照测试：
1. 将 `texas_algorithm.rar` 解压到项目根目录。
2. 运行 `TestUtil.main()` 或直接执行 `mvn test`。

---

## 重新生成数据表

1. 将 `texas_algorithm.rar` 解压到项目根目录。
2. 添加 JVM 参数 `-Xmx8000m`（需要约 8 GB 堆内存）。
3. 运行 `TexasAlgorithmUtil.main()`。
4. 在 8 核机器上约需 **10 小时**完成。

---

## 算法详解

### 查表算法

查表算法解决的问题是：给定 N 张牌（5–7 张），找出其中最优的 5 张牌面，并给出在所有可能组合中的绝对排名。

#### 第一步 — 穷举所有组合

从 52 张普通牌 + 2 张鬼牌中选 7 张，共约 1 亿种组合 C(54, 7)。每种组合编码为一个 `long` 类型的 key，得到约 1 亿长度的数组。6 张和 5 张的情况同理生成。

#### 第二步 — 多线程快速排序

对 1 亿条记录按牌力从小到大排序，排序依据是 7 选 5 后的最大牌大小。使用多线程快速排序，在 8 核机器上约需 10 小时。若将查表结果替换原始比牌逻辑作为排序比较器，可缩短至约 2 小时。

#### 第三步 — 输出原始文件

排序完成后，按顺序输出到 `texas_data.txt`（约 12 GB），每条记录包含：编码 key、排名序号、最优 5 张牌的编码、牌型、可读牌面字符串。由于排名是阶梯型的（多手牌强度相同但顺序不同），输出时需额外做一次比牌处理。

#### 第四步 — 去色压缩

1 亿条记录直接加载会撑爆内存。利用花色冗余性进行压缩：

- **同花类牌型**（同花、同花顺、皇家同花顺）：至少 5 张牌同花色，将花色分布统一归一化为 `♦♦♦♦♦♣♠`，大幅减少 key 的种类。
- **非同花牌型**：花色对牌力毫无影响，将所有花色统一替换为 `♦`。

压缩后生成两个文件（`texas_data_color.txt` 和 `texas_data_normal.txt`），合计约 18 MB，加载到内存约几十 MB。

#### 第五步 — 查询

给定 7 张牌：
1. 对手牌进行同花归一化，在同花表中查找。
2. 对手牌进行去色处理，在普通表中查找。
3. 两表都命中时，取排名更高的结果。

---

### 胜率评估算法

给定 2 张手牌和 0–4 张公牌，估算 1v1 情况下的胜率，而无需穷举所有剩余牌组合。

#### 第一步 — 生成胜率表

利用已有的 7 张牌排名表，对所有 N 张牌组合（2 ≤ N ≤ 6），遍历包含这 N 张牌的所有 7 张牌组合，计算平均胜率、最高胜率、最低胜率。生成 5 个输出文件，原始大小约 2 GB。

#### 第二步 — 去色压缩

对胜率表同样进行花色归一化，生成两张表（原始表和去色表），合计约 300 MB，运行时内存占用约 200 MB。

#### 第三步 — 查询逻辑

给定手牌 `H` 和公牌 `C`：

1. **P1** — 仅用公牌 `C` 查胜率表，得到仅用这些公牌组成 7 张牌时的平均胜率，以及 `P1_max` 和 `P1_min`。
2. **P2** — 用手牌 + 公牌一起查胜率表，得到玩家当前手牌的近似平均胜率（存在轻微误差，因手牌被重复统计）。
3. **差值插值** — 根据 P2 相对于 P1 的位置，在 `P1_min` 到 `P1_max` 之间线性插值，得出最终胜率估算值。

#### 误差说明

| 场景 | 真实胜率 | 估算结果 | 误差 |
|------|----------|----------|------|
| 一般情况 | 0.50 | 0.60 | ≤ 0.10 |

若采用精确穷举（固定手牌和公牌、枚举所有剩余公牌和对手手牌），2 张手牌 + 4 张公牌的情况需要超过 20 天才能计算完成，且数据量已超出实际限制。估算方案在可接受的误差范围内解决了这一问题。

---

## 相关项目

- [majiang_algorithm](https://github.com/esrrhs/majiang_algorithm) — 麻将算法
- [teenpatti_algorithm](https://github.com/esrrhs/teenpatti_algorithm) — 印度炸金花算法
