# Gil Dict SDK

这是一个用于获取远程数据字典信息的 Golang SDK。

## 安装

```bash
go get gitlab.xiaoluxue.cn/be-app/gil_dict_sdk
```

## 使用示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"gitlab.xiaoluxue.cn/be-app/gil_dict_sdk/cache"
	"gitlab.xiaoluxue.cn/be-app/gil_dict_sdk/demo/dlog"

	"gitlab.xiaoluxue.cn/be-app/gil_dict_sdk"
	"time"
)

func main() {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       0,
	})

	ctx := context.Background()

	redisCache := cache.NewRedisCache(ctx, redisClient)

	logger_ := dlog.NewLogger(dlog.Config{
		Path:         "./logs",
		Level:        "debug",
		Rotationtime: time.Hour,
		Maxage:       24 * time.Hour,
	})

	clog := dlog.NewContextLogger(logger_)

	// 创建客户端实例
	client := gil_dict_sdk.NewDictClient(
		ctx, redisCache, clog, gil_dict_sdk.DictClientOption{
			Domain:           "http://127.0.0.1:8080",
			Timeout:          30 * time.Second,
			RetryCount:       3,
			RetryWaitTime:    500 * time.Millisecond,
			RetryMaxWaitTime: 300 * time.Second,
		},
	)

	// 测试获取多个字典类型
	fmt.Println("测试获取多个字典类型...")
	response, err := client.GetDictByTypes(ctx, []string{"phase", "subject", "grade"}, "sys")
	if err != nil {
		log.Fatalf("获取字典数据失败: %v", err)
	}

	fmt.Printf("%+v\n", response)

	region, err := client.GetAllRegion(ctx)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v\n", region)

}

```

## 功能特性

- 支持 Redis 缓存集成
- 支持自定义日志配置
- 支持设置超时时间
- 支持设置重试次数和重试间隔
- 支持获取多个字典类型
- 支持获取单个字典类型
- 支持获取地区信息
- 使用 go-resty 进行 HTTP 请求
- 自动处理 JSON 序列化和反序列化

## 配置选项

创建客户端时可以配置以下选项：

- `Domain`: API 服务域名
- `Timeout`: 请求超时时间
- `RetryCount`: 重试次数
- `RetryWaitTime`: 重试等待时间
- `RetryMaxWaitTime`: 最大重试等待时间

## 缓存支持

SDK 支持 Redis 缓存，可以通过 `cache.NewRedisCache` 创建缓存实例。缓存配置包括：

- Redis 地址
- Redis 密码
- Redis 数据库编号

## 日志支持

SDK 支持自定义日志配置，可以通过 `dlog.NewLogger` 创建日志实例。日志配置包括：

- 日志路径
- 日志级别
- 日志轮转时间
- 日志保留时间

## 错误处理

SDK 会返回标准的 Go 错误，建议在使用时进行适当的错误处理。主要错误类型包括：

- 网络请求错误
- 缓存操作错误
- 数据解析错误
- 参数验证错误

## 开发环境要求

- Go 1.22 或更高版本
- Redis 服务器
- 支持 HTTP/HTTPS 的网络环境

## 贡献指南

欢迎提交 Pull Request 或创建 Issue 来帮助改进这个项目。在提交代码之前，请确保：

1. 代码符合 Go 标准规范
2. 添加了必要的测试用例
3. 更新了相关文档

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

# gil_dict_sdk



## Getting started

To make it easy for you to get started with GitLab, here's a list of recommended next steps.

Already a pro? Just edit this README.md and make it your own. Want to make it easy? [Use the template at the bottom](#editing-this-readme)!

## Add your files

- [ ] [Create](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#create-a-file) or [upload](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#upload-a-file) files
- [ ] [Add files using the command line](https://docs.gitlab.com/ee/gitlab-basics/add-file.html#add-a-file-using-the-command-line) or push an existing Git repository with the following command:

```
cd existing_repo
git remote add origin https://gitlab.xiaoluxue.cn/be-app/gil_dict_sdk.git
git branch -M main
git push -uf origin main
```

## Integrate with your tools

- [ ] [Set up project integrations](https://gitlab.xiaoluxue.cn/be-app/gil_dict_sdk/-/settings/integrations)

## Collaborate with your team

- [ ] [Invite team members and collaborators](https://docs.gitlab.com/ee/user/project/members/)
- [ ] [Create a new merge request](https://docs.gitlab.com/ee/user/project/merge_requests/creating_merge_requests.html)
- [ ] [Automatically close issues from merge requests](https://docs.gitlab.com/ee/user/project/issues/managing_issues.html#closing-issues-automatically)
- [ ] [Enable merge request approvals](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/)
- [ ] [Set auto-merge](https://docs.gitlab.com/ee/user/project/merge_requests/merge_when_pipeline_succeeds.html)

## Test and Deploy

Use the built-in continuous integration in GitLab.

- [ ] [Get started with GitLab CI/CD](https://docs.gitlab.com/ee/ci/quick_start/index.html)
- [ ] [Analyze your code for known vulnerabilities with Static Application Security Testing (SAST)](https://docs.gitlab.com/ee/user/application_security/sast/)
- [ ] [Deploy to Kubernetes, Amazon EC2, or Amazon ECS using Auto Deploy](https://docs.gitlab.com/ee/topics/autodevops/requirements.html)
- [ ] [Use pull-based deployments for improved Kubernetes management](https://docs.gitlab.com/ee/user/clusters/agent/)
- [ ] [Set up protected environments](https://docs.gitlab.com/ee/ci/environments/protected_environments.html)

***

# Editing this README

When you're ready to make this README your own, just edit this file and use the handy template below (or feel free to structure it however you want - this is just a starting point!). Thanks to [makeareadme.com](https://www.makeareadme.com/) for this template.

## Suggestions for a good README

Every project is different, so consider which of these sections apply to yours. The sections used in the template are suggestions for most open source projects. Also keep in mind that while a README can be too long and detailed, too long is better than too short. If you think your README is too long, consider utilizing another form of documentation rather than cutting out information.

## Name
Choose a self-explaining name for your project.

## Description
Let people know what your project can do specifically. Provide context and add a link to any reference visitors might be unfamiliar with. A list of Features or a Background subsection can also be added here. If there are alternatives to your project, this is a good place to list differentiating factors.

## Badges
On some READMEs, you may see small images that convey metadata, such as whether or not all the tests are passing for the project. You can use Shields to add some to your README. Many services also have instructions for adding a badge.

## Visuals
Depending on what you are making, it can be a good idea to include screenshots or even a video (you'll frequently see GIFs rather than actual videos). Tools like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation
Within a particular ecosystem, there may be a common way of installing things, such as using Yarn, NuGet, or Homebrew. However, consider the possibility that whoever is reading your README is a novice and would like more guidance. Listing specific steps helps remove ambiguity and gets people to using your project as quickly as possible. If it only runs in a specific context like a particular programming language version or operating system or has dependencies that have to be installed manually, also add a Requirements subsection.

## Usage
Use examples liberally, and show the expected output if you can. It's helpful to have inline the smallest example of usage that you can demonstrate, while providing links to more sophisticated examples if they are too long to reasonably include in the README.

## Support
Tell people where they can go to for help. It can be any combination of an issue tracker, a chat room, an email address, etc.

## Roadmap
If you have ideas for releases in the future, it is a good idea to list them in the README.

## Contributing
State if you are open to contributions and what your requirements are for accepting them.

For people who want to make changes to your project, it's helpful to have some documentation on how to get started. Perhaps there is a script that they should run or some environment variables that they need to set. Make these steps explicit. These instructions could also be useful to your future self.

You can also document commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something. Having instructions for running tests is especially helpful if it requires external setup, such as starting a Selenium server for testing in a browser.

## Authors and acknowledgment
Show your appreciation to those who have contributed to the project.

## License
For open source projects, say how it is licensed.

## Project status
If you have run out of energy or time for your project, put a note at the top of the README saying that development has slowed down or stopped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.
