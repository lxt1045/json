# 本次维护工作记录

> 目的: 记录本次修复/升级的背景、问题定位与修改点, 供下次再次维护本项目时复用, 避免重新思考。

## 背景

本库 (`github.com/lxt1045/json`) 是一个基于 `unsafe.Pointer` + 内存池的高性能 JSON 库。
早期在 Go 1.18 之前的版本下用例全部可以通过, 但在新版 Go (本次环境 = Go 1.25) 下执行 `go test ./...`
时会立即 fatal。本次任务的四条 TODO:

1. 修复新版 Go 下的测试 panic。
2. 查漏补缺、修基础功能 BUG。
3. 修复历史原因导致已经跑不起来的 `*_test.go`。
4. 根据现有代码补充实用功能。

## 核心问题与修复

### 1. fatal: `invalid pointer found on stack`

**现象**: `TestStruct` / `TestStructST/qq1` 等用例在 `Unmarshal` 过程中 runtime 直接 fatal,
栈帧指向 `parseObj` → `func_marshal_unmarshal.structMFuncs.funcXX` → `PoolStore.Idx`。

**根因**:

- `PoolStore.pointerPool` / `slicePool` 是 `unsafe.Pointer` 字段。
- 入口 `UnmarshalString` 若该结构体没有指针字段, `tag.ptrCache == nil`, 因此
  `store.pointerPool` 保持零值 (nil)。
- 但 `structMFuncs` / `sliceMFuncs` 内部无条件执行:
  ```go
  store.pointerPool = pointerOffset(store.pointerPool, *sonPidx)
  store.slicePool   = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
  ```
  `pointerOffset(nil, 0x180)` 得到 `0x180` —— 一个"像指针"的小整数, 回塞进
  `unsafe.Pointer` 类型字段。
- Go 1.18+ 的栈扫描对 unsafe.Pointer 值更加严格, 这种"像指针但不指向任何分配"的值会触发 fatal。

**修复** (`func_marshal_unmarshal.go`):

所有对 `store.pointerPool` / `store.slicePool` 的偏移操作加了非 nil 守卫:

```go
if store.pointerPool != nil {
    store.pointerPool = pointerOffset(store.pointerPool, *sonPidx)
}
if store.slicePool != nil {
    store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
}
```

关键是: **基址为 nil 就保持 nil**, 任何后续调用 `store.Idx(*pidx)` 只会在
`pidx != nil` (即存在指针字段, 此时必然有 `ptrCache != nil`, `pointerPool != nil`)
时发生, 不会再构造出无效指针。

### 2. panic: `Nested loops are not yet supported`

**现象**: `TestStructST` 中 `type ST struct { ST *ST }` 自引用结构体, 直接 panic。

**根因**: `NewStructTagInfo` 构建 tag 时发现祖先链上已有同名类型就直接 panic。
当前所有 fUnm/fM 都是在"构建时"就生成好的闭包, 递归类型天然无法一次构建完。

**修复** (`struct_tag.go`):

1. `NewStructTagInfo` 的祖先检查改为: 命中祖先时直接返回祖先 tag, 不再递归构建。
2. `setFuncs` 的 `reflect.Struct` 分支在递归检测命中时, **不复用** `structMFuncs` 的默认闭包,
   而是安装一个运行时代理:
   - 在代理被调用时 (此时外层 `LoadTagNodeSlow` 早已返回, 祖先 tag 已完整构建完毕),
     把 `store.tag` 切换到祖先 tag;
   - 为嵌套层单独 `Get()` 一个 pointerPool, 避免偏移累加。

参考代码: `struct_tag.go` `recursiveAncestor != nil` 分支。

### 3. `tireTree.Get2` 在 `-race` 下 fatal (checkptr)

**现象**: `go test -race` 下 `Test_binTree` fatal
`checkptr: converted pointer straddles multiple allocations`。

**根因**: 旧实现
```go
p := (*[1 << 20]tireTreeNode)(unsafe.Pointer(&root.tree[0]))
```
把 `[][128]tireTreeNode` 展平为一维数组, checkptr 判定为跨分配访问。

**修复** (`tire_tree.go`): 改用二维索引 `tree[row][col]`, 语义一致, Go 编译器内联后汇编等价。

### 4. 字符串序列化转义不完整 (正确性 BUG)

**现象**: `stringMm` / `marshalKey` 只处理了 `"` 和 `\\`, 未处理 `\n`, `\r`, `\t`, `\b`, `\f`,
以及 `\x00-\x1f` 的控制字符。对含这些字符的字符串做 Marshal 输出的 JSON 是**非法**的,
`encoding/json` 也无法反解。

**修复** (`func_marshal_unmarshal.go`, `marshal.go`):

- 重写 `stringMm`, 按 RFC 8259 一次遍历转义 `"`, `\\`, `\b`, `\f`, `\n`, `\r`, `\t`
  以及其余控制字符 (`\u00XX`)。
- `marshalKey` 对字符串 key 复用 `stringMm`, 消除两处不一致实现。
- 新增 `extra_test.go::TestStringEscape`, 用"标准库能否反解析"作为断言准则。

### 5. 其它清理

- 删除 `type_builder.go` 里的 `main()` (库包不该有 main)。
- 删除 `NewStructTagInfo` 中已彻底跳不到的 "Nested loops are not yet supported" panic。
- 把 pool 安全相关的沉淀到 `.cursor/rules/unsafe-pointer-safety.mdc`, 防止后续 regression。

## 新增功能

`extra.go` 补齐了日常使用里最常缺的两个 API, 与 `encoding/json` 签名保持一致:

| API | 说明 |
|-----|------|
| `MarshalIndent(v, prefix, indent) ([]byte, error)` | 带缩进的 Marshal, 内部对 `Marshal` 输出做流式缩进 |
| `AppendIndent(dst, src, prefix, indent) []byte`    | 对外暴露的缩进函数, 可以直接作用于已有的 JSON 字节 |
| `Valid(data []byte) bool`                          | 仅做语法扫描, O(n) 时间, 不分配内存 |
| `ValidString(s string) bool`                       | 字符串版本 |

`extra_test.go` 覆盖了上述 API 的正/反用例, 以及字符串转义的各种特殊字符。

## 测试情况

环境: Go 1.25, Windows amd64。

- `go test -count=1 ./...` ✅ 全绿
- `go test -race -count=1 ./...` ✅ 全绿 (包含 checkptr)
- `TestStructST/qq1` ✅
- `TestStringEscape` / `TestValidString` / `TestMarshalIndent` ✅

## 给下次维护者的提示

1. 遇到 "invalid pointer found on stack" 或 checkptr fatal, 先 grep
   `pointerOffset\(store\.(pointerPool|slicePool)` 和 `unsafe\.Pointer\(&.*\[0\]\)`,
   大概率是这类老问题复发。
2. 新增 `PoolStore` 字段若带指针, 必须同时在每个 `Idx` / `Get` 调用点考虑"未初始化为 nil"的情况。
3. 新类型支持 (例如 map 的更多 value 类型) 建议先写 round-trip 测试 (Marshal → stdjson.Unmarshal),
   再加实现, 能避免大量看不见的转义/语法 BUG。
4. 性能回归: 基准在 `api_bench_test.go` / `bench_test.go`, 改动前后至少做一次
   `go test -bench=. -benchmem -run=^$ -count=3` 对比。
5. `.cursor/rules/` 里的规则必须一并更新, 特别是当出现新的 unsafe 使用模式时。
