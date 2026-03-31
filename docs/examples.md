# Usage examples

Concrete **arguments** and **structured results** for each tool. Paths are **relative to `root`** (here `"/data/proj"`). Line breaks in files are `\n`.

For JSON-RPC wrapping and errors, see [tool-contract.md](tool-contract.md).

---

## `server.ping`

**Input:** `{}`

**Output:**

```json
{ "ok": true }
```

---

## `count_lines_words_bytes`

**File** `notes.txt`:

```text
hello world
second line
```

**Arguments:**

```json
{ "path": "notes.txt", "root": "/data/proj" }
```

**Output:**

```json
{
  "path": "/data/proj/notes.txt",
  "lines": 2,
  "words": 4,
  "bytes": 24
}
```

**Edge: stricter cap**

```json
{ "path": "notes.txt", "root": "/data/proj", "max_bytes": 10 }
```

If the resolved file is larger than **10 bytes**, the call fails with `validate: file exceeds maximum size` (wrapped in `tool execution failed`).

---

## `top_terms`

**File** `tweet.txt`:

```text
Paris paris PARIS! café 123
```

**Arguments:**

```json
{
  "path": "tweet.txt",
  "root": "/data/proj",
  "top_n": 10,
  "min_len": 2,
  "case_sensitive": false,
  "max_lines": 1000
}
```

With **`case_sensitive: false`**, the line is lowercased before tokenizing, so **`paris`** counts **3**. **`café`** is letters-only token **`caf`** + **`é`** as Unicode letters → one token `café` if entire run is letters (actually café is c a f é - one token). **`123`** is a numeric token; **`min_len`** is in **runes**, so **`12`** length 2 passes if min_len 2.

**Output** (order: count desc, term asc):

```json
{
  "path": "/data/proj/tweet.txt",
  "top_n": 10,
  "terms": [
    { "term": "paris", "count": 3 },
    { "term": "123", "count": 1 },
    { "term": "café", "count": 1 }
  ]
}
```

(Exact token list depends on UTF-8 input; numeric vs letter grouping follows `unicode.IsLetter` / `IsNumber`.)

**Edge: short tokens skipped**

`min_len: 5` may drop **`123`** (3 runes) and short tokens.

---

## `unique_values`

**File** `cols.txt` (comma-separated, **not** RFC CSV — no quoted commas):

```text
1,alice,east
2,bob,east
3,carol,north
```

**Arguments** (unique values of column index **1**):

```json
{
  "path": "cols.txt",
  "root": "/data/proj",
  "field_index": 1,
  "delimiter": ",",
  "trim_space": true,
  "max_unique": 100,
  "max_lines": 1000000
}
```

**Output** (sorted ascending):

```json
{
  "path": "/data/proj/cols.txt",
  "field_index": 1,
  "values": ["alice", "bob", "carol"]
}
```

**Edge: short rows**

Line `only-two` with **`field_index: 2`** → row skipped (not enough fields).

**Edge: quoted CSV**

```text
1,"Smith, Jr.",east
```

**Not** supported — delimiter split breaks inside quotes. Pre-normalize or use a real CSV tool upstream.

**Edge: `max_unique`**

With thousands of new distinct values, scanning stops after **`max_unique`** (default 1000, hard max 10_000); `values` may be incomplete but sorted among collected set.

---

## `field_frequency`

**File** `hits.log` (field 0 = user id):

```text
u1,200
u2,404
u1,200
u3,500
u1,200
```

**Arguments:**

```json
{
  "path": "hits.log",
  "root": "/data/proj",
  "field_index": 0,
  "delimiter": ",",
  "top_n": 5,
  "include_empty": false,
  "max_lines": 1000000
}
```

**Output** (count desc, value asc):

```json
{
  "path": "/data/proj/hits.log",
  "field_index": 0,
  "top_n": 5,
  "frequencies": [
    { "value": "u1", "count": 3 },
    { "value": "u2", "count": 1 },
    { "value": "u3", "count": 1 }
  ]
}
```

**Edge: empty field**

File:

```text
u1,
u2,y
```

Field index **1**: line 1 has empty second field.

- **`include_empty: false`** (default): empty values not counted → only **`y`** appears.
- **`include_empty: true`**: `""` counted; ties sort with **empty string first** in UTF-8 order before **`y`**.

**Edge: memory**

Very high **cardinality** fields keep all keys in a map until sort — monitor heap for huge logs.

---

## See also

- [tool-contract.md](tool-contract.md) — limits, error codes, full contract.
- [troubleshooting.md](troubleshooting.md) — when examples fail in practice.
- [architecture.md](architecture.md) — how tools are implemented.
