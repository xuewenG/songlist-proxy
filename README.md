# Songlist Proxy

## 项目功能

Songlist Proxy 是一个使用 Go 开发的，用于代理 [歌单 - Songlist](https://bot.starlwr.com/songlist/edit) 的项目。

因为原项目部署在 Cloudflare，所以部分用户可能会遇到访问不了的问题，可以通过本项目在优化线路的服务器上代理原项目，达到优化访问速度的目的。

## 部署方式

### 代理前端页面

首先，使用 Traefik 或 Nginx 代理原项目的前端页面。在代理页面时，同时需要替换前端页面使用的 API 域名，例如在 Traefik 中可以使用 rewritebody 中间件来进行替换：

```yaml
http:
  middlewares:
    rewrite_api_starlwr_com_to_list_example_com:
      plugin:
        rewritebody:
          lastModified: true
          rewrites:
            - regex: api.starlwr.com
              # 替换成你使用的域名
              replacement: list.example.com
```

### 代理后端接口

然后，使用本项目代理原项目的 API。可使用 Docker Compose 来部署：

```yaml
services:
  songlist-proxy:
    image: ixuewen/songlist-proxy
    container_name: songlist-proxy
    restart: always
    volumes:
      # 提供配置文件到容器中，可替换为你自己的本地路径
      - ./songlist-proxy/config.yaml:/app/config.yaml
```

配置文件说明如下：

```yaml
app:
  # 端口号
  port: 80
bili:
  # 用户的 UID
  uid: 123456789
  # 原项目存在偶尔获取不到用户头像地址的问题，这里填写用户的默认头像地址
  avatar: https://img.example.com/avatar.png
  # 保持不变即可
  url: bot.starlwr.com
```
