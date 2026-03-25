# adnx_dns v1.0.0

Release name: `adnx_dns`
Version: `1.0.0`

## 建议发布说明

- 初始版本
- 支持 GoDaddy 主域名定时同步到本地 MySQL
- 支持按 IPv4 绑定 A 记录
- 支持按 IP 或子域名解绑
- 支持本地禁用主域名
- 所有业务接口要求 `api_token`
- 增加请求节流，命中限流直接返回错误

## 建议命令

```bash
git init
git add .
git commit -m "release: adnx_dns v1.0.0"
git remote add origin <your-github-repo>
git push -u origin main
git tag v1.0.0
git push origin v1.0.0
bash scripts/build_release.sh
```

然后把 `dist/` 下的压缩包上传到 GitHub Release 即可。
