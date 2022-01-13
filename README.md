# mc_mod_sync
mc mod 同步工具

# 使用

## 打包 mod

```
python path/app.py gen
```

会生成一个 mod.zip 压缩包，将压缩包解压存放在一个能被访问的web服务器上即可

## 下载 mod

进入 mc mod 文件夹，使用指令
```
python path/app.py get http://url/update.json
```