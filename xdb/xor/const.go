package xor

// KWRand 常量关键字，在使用 Model 查询数据的时候，可以放在 order by 语句中，让结果随机排序：
// 如 score> 0 order by X:RAND()
// 在 Model 执行过程中，关键字 “X:RAND()” 会被替换为数据库方言的对于函数名。
//
// 大多数情况下可以直接使用 OrderByRand() 配置项
const KWRand = `X:RAND()`
