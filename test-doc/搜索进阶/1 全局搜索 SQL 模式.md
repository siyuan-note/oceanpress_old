## `blocks` 表
{: id="20201117103851-l9cahuc"}

该表用于存储内容块数据。
{: id="20201224120447-bx9fwlf"}

|  字段 | 类型 | 说明                                      |
| --------: | :------: | --------------------------------------------- |
|      id |  text  | 内容块 ID                                |
|     box |  text  | 笔记本名                                |
|    path |  text  | 内容块所在文档路径                 |
| tree_id |  text  | 抽象语法树 ID，和根节点 ID 相同 |
| content |  text  | 内容块 Markdown                          |
|    type |  text  | 内容块类型                             |
|    time |  text  | 日期时间                                |
{: id="20201224120447-cs6ur9x"}

示例：
{: id="20201224120447-ub63xgw"}

```sql
SELECT * FROM blocks WHERE content LIKE '%内容块%'
```
{: id="20201224120447-ieeh0ht"}

`type` 内容块类型值列表请参考((20201224092621-71qvlyb "这里"))。
{: id="20201224120447-g34nwns"}

## 默认查询条件
{: id="20201224120447-lezmdb3"}

* {: id="20201224120447-u4tbyb3"}如果不指定 `LIMIT`，则最多只返回前 32 条结果
{: id="20201224120447-ij3ivd5"}


{: id="20201222093044-rx4zjoy" type="doc"}
