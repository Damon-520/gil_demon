# 项目名称 
   
GIL_teacher 服务
> 这是一个基于 Go 语言开发的 API 服务框架，提供标准化的项目结构和常用功能组件。

## 运行条件

> 列出运行该项目所必须的条件和相关依赖

* Go 版本 1.22+
* Make
* Protocol Buffers
* PostgreSQL
* Elasticsearch
* Wire (依赖注入工具)

## 目录介绍

``` text
├── app/                   # 项目内部实现
│   ├── conf               # 配置结构体
│   ├── consts             # 存放项目中的常量定义
│   ├── core               # 核心功能模块，可能包括基础设施、工具类等
│   ├── http               # HTTP 请求处理，控制器层
│   ├── dao                # 数据访问层，数据库表的定义和操作
│   ├── model              # 数据模型层，定义数据结构
│   │    └── api           # API 相关的数据模型，每个接口对应一个定义文件，包含输入和输出定义
│   │    └── dto           # 数据传输对象，用于在服务之间传递数据
│   ├── router             # 路由配置
│   ├── server             # 服务器相关功能
│   └── service            # 业务服务层
│   └── middleware         # 中间件
│   └── third_party        # 项目专用三方库
│   └── util               # 项目自定义工具库
├── bin                    # 编译后的可执行文件
│   └── gil_teacher        # 启动的二进制文件
├── script                 # 定时任务
├── main                   # 应用程序的入口文件
│   └── gil_teacher        # 具体服务的入口
│       ├── main.go        # 启动文件
│       ├── wire.go        # 使用 Wire 进行依赖注入的配置文件
│       └── wire_gen.go    # Wire 生成的依赖注入代码
├── configs                # 配置文件目录
│   ├── gray/              # 灰度环境配置
│   ├── local/             # 本地开发环境配置
│   ├── online/            # 线上环境配置
│   ├── test/              # 测试环境配置
│   ├── rpc.yaml           # RPC 服务相关配置
│   └── user_privacy.json  # 用户隐私配置（JSON 格式）
├── docs/                  # 项目的文档
│   ├── README.md          # 项目说明文档
│   └── update_20250214_activity.sql  # 测试活动数据库脚本
├── log/                    # 日志目录
├── libs/                   # 公共库目录
├── mock/                   # Mock 相关库
├── third_party/            # 第三方库
├── proto/                  # proto文件
│   ├── gil_teacher/        # 所有路由定义
│   │   ├── api             # 路由定义
│   │   ├── base            # 基础类型
│   │   ├── demo            # 示例类型
│   │   └── user            # 用户类型
│   ├── gen/go/proto/gil_teacher/   # 生成文件
│   │   ├── api             # 路由文件
│   │   │    ├── xxx_pb.go  # 出入参结构体
│   │   │    ├── xxx_grpc.pb.go # grpc路由方法代码
│   │   │    └── xxx_grpc.pb.gw.go # http转grpc代码
│   │   ├── base            # 基础文件
│   │   ├── demo          # 示例文件
│   │   └── user          # 用户文件
│   └── openapi/               # swagger文件

```

## 运行说明

> 项目使用 Make 进行构建和运行管理

1. 生成各项依赖代码

```shell
make all
```

1. 编译项目

```shell
make build
```

1. 运行服务（默认使用 local 环境配置）

```shell
bin/gil_teacher 
```

### 环境配置

项目支持多环境配置，配置文件位于 `configs` 目录：

* local: 本地开发环境
* test: 测试环境
* gray: 灰度环境
* online: 生产环境

可通过环境变量指定配置：

```shell
ENV=test bin/gil_teacher
```

## 新增或修改路由

> 当新增或修改路由protoc后，需要重新生成代码：

```shell
make proto
```

## 测试说明

> 导入 doc/update_20250214_activity.sql 文件
> 配置 数据库用户密码和地址

1. 配置数据库连接

修改对应环境配置文件中的数据库配置：

* 用户名
* 密码
* 数据库地址
* 数据库名称

## 技术架构

本项目采用清晰的分层架构：

* Controller 层：处理 HTTP 和 GRPC 请求和响应
* Service 层：实现核心业务逻辑
* DAO 层：数据访问层，处理数据库操作
* Core 层：提供基础设施和工具类
* Server 层：提供服务注册和启动

主要技术栈：
- Go 1.22+
- Wire (依赖注入)
- Elasticsearch
- Pos
- Protocol Buffers

## 协作者

请在参与项目开发时，将你的信息添加到下面的列表：

* [zengwei](https://gitlab.xiaoluxue.cn/zengwei) - Contribution

## License

[选择合适的开源协议]

## 测试结果

1W并发数量，执行10S结果
![img.png](docs/img.png)
