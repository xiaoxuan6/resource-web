# 简述
    将 github 私有库中的笔记，渲染到 web 网站上。

# 使用方式

克隆项目

```bash
git clone https://github.com/xiaoxuan6/resource-web.git
```

复制 `.env.example` 为 `.env`, 修改里面的参数为自己的配置。

> **Note**
> 
> 私有库中的笔记文件名格式必须为：`xxx.md`
> 
> 私有库中的笔记内容格式必须为：
> 
>       [描述1](https://www.baidu.com)<br>
>       [描述2](https://www.baidu.com)<br>

然后运行

```go
go run main.go
```

## Docker部署

环境要求：Git、Docker、Docker-Compose

克隆项目

```bash
git clone https://github.com/xiaoxuan6/resource-web.git
```

进入 `resource-web` 文件夹，运行项目

```bash
docker-compose up -d
```

部署成功后，通过 `ip + 端口号` 访问，默认端口为：`8080`

# 相关
[cli 模式](https://github.com/xiaoxuan6/rsearch)