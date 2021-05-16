## overview

[![codecov](https://codecov.io/gh/zhyoulun/gls/branch/master/graph/badge.svg?token=10FUXUMWAN)](https://codecov.io/gh/zhyoulun/gls)

## v1

- 原则：不考虑性能，按协议进行标准化的实现
- 如果发现有可以优化性能的地方，加上todo优化即可
- 尽量少的封装
- 尽量使用golang原生函数
- 尽量少的使用反射，便于重构
- 学习使用
- 日志：chunk/chunk stream/packet级别的日志，写成trace的，需要额外使用debug包打到csv文件中，方便调试

## 依赖关系

0. core, utils
1. av, amf
2. flv
3. rtmp
4. stream
5. server
6. cmd

## 调试工具

- wireshark
  - 过滤条件：rtmp and tcp.port>xx
- print csv excel
