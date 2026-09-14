# Texas Hold'em Algorithm (with Joker Support)

[中文文档](README_CN.md)

A high-performance Java library for Texas Hold'em poker with Joker (wild card) support. Supports up to **2 Jokers** and provides two independent subsystems:

| Subsystem | Memory | Capability |
|-----------|--------|------------|
| **Lookup Table** | ~tens of MB | Best-hand evaluation, rank, type for 5–7 cards |
| **Win Probability** | ~200 MB | 1v1 win probability estimate for 2 hole + 0–4 community cards |

---

## Maven Dependency

```xml
<dependency>
    <groupId>com.github.esrrhs</groupId>
    <artifactId>texas_algorithm</artifactId>
    <version>1.0.14</version>
</dependency>
```

---

## Quick Start

### 1. Load Tables

```java
// Load lookup table (~tens of MB, required for hand evaluation)
TexasAlgorithmUtil.load();

// Load probability table (~200 MB, required for win probability estimation)
TexasAlgorithmUtil.loadProbility();
```

### 2. Get Best 5-Card Hand

```java
// Get the best 5-card hand from 2 hole cards + 5 community cards
// Returns a string of the best 5 cards
String best = TexasAlgorithmUtil.getMax("黑2,黑3", "方2,方A,黑7,黑5,鬼");
```

### 3. Compare Hands (Hand Ranking)

```java
// Get the absolute rank of a 7-card hand (higher = stronger)
// Used to compare two hands: whichever has a higher position wins
int rank = TexasAlgorithmUtil.getWinPosition("方4,方A,鬼,黑A,黑3,黑5,黑6");

// Compare two 7-card hands directly: returns positive if str1 > str2, 0 if equal, negative if str1 < str2
int cmp = TexasAlgorithmUtil.compare("方4,方A,鬼,黑A,黑3,黑5,黑6", "黑2,红3,方7,梅9,方K,黑Q,红J");
```

### 4. Get Hand Type

```java
// Returns an integer constant for the hand type (see Hand Types below)
int type = TexasAlgorithmUtil.getWinType("方4,方A,鬼,黑A,黑3,黑5,黑6");
```

### 5. Estimate Win Probability (1v1)

```java
// Estimate win probability given 2 hole cards + some community cards
// Returns a float in [0, 1]
float p = TexasAlgorithmUtil.getHandProbability("方3,鬼", "黑2,黑4,黑5,黑K");
```

---

## Card Notation

Cards are expressed as Chinese-character strings separated by commas.

### Suits

| Notation | Suit | English |
|----------|------|---------|
| `方` | ♦ | Diamond |
| `梅` | ♣ | Club |
| `红` | ♥ | Heart |
| `黑` | ♠ | Spade |
| `鬼` | Joker | Wild card |

### Values

`2` `3` `4` `5` `6` `7` `8` `9` `10` `J` `Q` `K` `A`

### Examples

```
方A    = Diamond Ace
黑K    = Spade King
红10   = Heart Ten
鬼     = Joker (wild card, can substitute any card)
```

---

## Hand Types

The `getWinType()` method returns one of these constants from `TexasCardUtil`:

| Constant | Value | Hand |
|----------|-------|------|
| `TEXAS_CARD_TYPE_GAOPAI` | 1 | High Card |
| `TEXAS_CARD_TYPE_DUIZI` | 2 | One Pair |
| `TEXAS_CARD_TYPE_LIANGDUI` | 3 | Two Pair |
| `TEXAS_CARD_TYPE_SANTIAO` | 4 | Three of a Kind |
| `TEXAS_CARD_TYPE_SHUNZI` | 5 | Straight |
| `TEXAS_CARD_TYPE_TONGHUA` | 6 | Flush |
| `TEXAS_CARD_TYPE_HULU` | 7 | Full House |
| `TEXAS_CARD_TYPE_SITIAO` | 8 | Four of a Kind |
| `TEXAS_CARD_TYPE_TONGHUASHUN` | 9 | Straight Flush |
| `TEXAS_CARD_TYPE_KINGTONGHUASHUN` | 10 | Royal Flush |

---

## API Reference

### `TexasAlgorithmUtil`

```java
// Load / unload
void load()                                   // Load lookup tables into memory (current directory)
void load(String dirPath)                     // Load lookup tables from specified directory
void load(File dir)                           // Load lookup tables from specified directory
void loadProbility()                          // Load probability tables into memory (current directory)
void loadProbility(String dirPath)            // Load probability tables from specified directory
void loadProbility(File dir)                  // Load probability tables from specified directory
boolean isLoaded()                            // Check if lookup tables are loaded
boolean isProbabilityLoaded()                 // Check if probability tables are loaded

// Best hand
String getMax(String hand, String pub, ...)   // Best 5 cards from 2 hole + 3–5 community
List<Byte> getMax(List<Byte> pokes, ...)      // Same, accepting byte lists

// Hand evaluation (requires load())
int    getWinPosition(String cards)           // Absolute rank among all combinations
double getWinProbability(String cards)        // Win ratio vs. all same-size combinations
int    getWinType(String cards)               // Hand type constant
long   getWinMax(String cards)                // Encoded key of best 5-card hand
int    compare(String str1, String str2)      // Compare two 7-card hands

// Win probability estimate (requires loadProbility())
float  getHandProbability(String hand, String pub)  // 1v1 win probability estimate
```

---

## Running Unit Tests

```bash
# Run JUnit 5 tests (compatible with Java 8, 11, 17, 21)
mvn test
```

To run the full benchmark and lookup verification:
1. Extract `texas_algorithm.rar` into the project root directory.
2. Run `TestUtil.main()` or execute `mvn test`.

---

## How to Regenerate the Data Tables

1. Extract `texas_algorithm.rar` into the project root.
2. Run `TexasAlgorithmUtil.main()` with JVM flag `-Xmx8000m` (requires ~8 GB heap).
3. Generation takes approximately **10 hours** on an 8-core machine.

---

## Algorithm Details

### Lookup Table Algorithm

The lookup table answers "given N cards (5–7), what is the best possible 5-card hand and its absolute rank?"

#### Step 1 — Enumerate All Combinations

All C(54, 7) combinations from a 52-card deck plus 2 Jokers are enumerated (~100 million combinations). Each 7-card hand is encoded into a `long` key. The same process is repeated for 6-card and 5-card hands.

#### Step 2 — Multi-threaded Sort

The ~100 million entries are sorted by hand strength using multi-threaded quicksort. On an 8-core machine this takes ~10 hours. If the lookup table itself is used as the comparator instead of the brute-force evaluator, this can be reduced to ~2 hours.

#### Step 3 — Output Raw File

The sorted array is written to `texas_data.txt` (~12 GB), recording the encoding key, rank order, best-5-card value, hand type, and human-readable card string for each entry.

#### Step 4 — Suit Normalization (Color Removal)

100 million entries cannot fit in practical memory. The key insight is that suit information is redundant for non-flush hands:

- **Flush hands** (Flush, Straight Flush, Royal Flush): at least 5 cards share a suit. The suit distribution is normalized to `♦♦♦♦♦♣♠`, reducing the keyspace dramatically.
- **Non-flush hands**: suits are irrelevant; all suits are collapsed to `♦`.

This produces two compact files (`texas_data_color.txt` and `texas_data_normal.txt`) with a combined size of ~18 MB, loading to tens of MB in memory.

#### Step 5 — Query

Given 7 cards:
1. Look up the normalized flush key in the color table.
2. Look up the suit-stripped key in the normal table.
3. If both match, return the entry with the higher rank.

---

### Win Probability Estimation Algorithm

Given 2 hole cards and 0–4 community cards, this estimates 1v1 win probability without exhaustively enumerating all remaining card combinations.

#### Step 1 — Probability Table Generation

Using the 7-card rank table from above, for every N-card combination (2 ≤ N ≤ 6), the average win probability is computed by iterating over all 7-card supersets containing those N cards. This produces 5 output files (~2 GB total before compression).

#### Step 2 — Suit Normalization

The same suit-normalization trick is applied, reducing the 5 files to ~300 MB (two tables each: original and suit-stripped). Runtime memory usage is ~200 MB.

#### Step 3 — Query Logic

Given hole cards `H` and community cards `C`:

1. **P1** — Look up the community cards alone in the probability table.  
   This gives the average win probability for any hand using those community cards, plus `P1_max` and `P1_min`.

2. **P2** — Look up the combined hole + community cards.  
   This approximates the average win probability for the player's specific hand. (Minor inaccuracy: hole cards are counted twice.)

3. **Interpolation** — Using the relationship between `P2`, `P1`, `P1_max`, and `P1_min`, the final win probability is estimated by linear interpolation, assuming a uniform distribution of opponent hands.

#### Accuracy

| Scenario | True probability | Typical estimate | Error |
|----------|-----------------|------------------|-------|
| General | 0.50 | 0.60 | ≤ 0.10 |

Exhaustive calculation (fixing hole + community, enumerating all remaining cards and opponents) would take 20+ days for 2 hole + 4 community cards and exceeds practical data size limits. The estimation approach achieves acceptable accuracy within those constraints.

---

## Related Projects

- [majiang_algorithm](https://github.com/esrrhs/majiang_algorithm) — Mahjong algorithm
- [teenpatti_algorithm](https://github.com/esrrhs/teenpatti_algorithm) — Teen Patti (Indian poker) algorithm
