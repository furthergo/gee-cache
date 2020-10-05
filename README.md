[Gee-Cache](https://geektutu.com/post/geecache.html)学习

# Gee cache梳理

## 框架
* 分布式内存缓存
* O(1)的time和space
* LRU淘汰算法
* 多节点一致性Hash

# 大纲

1. 内存缓存
    * 双向链表加map，map用于kv存储；双向链表存储key，用于LRU淘汰
    * LRU实现
    * Getter用于缓存miss时获取数据
2. 分布式
    * HTTP服务器，用于节点间查询缓存，采用pb编码
    * 一致性Hash，用于计算响应请求的节点
        * 收到查询key的value的请求
        * 查询cache，未命中判断是否应当从其它节点获取
        * 根据key用一致性Hash算法计算应请求的节点，发送请求
        * 应从本机节点请求，调用Getter，并更新缓存
    * singleFlight，防止缓存击穿
    