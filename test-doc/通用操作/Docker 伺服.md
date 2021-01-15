在服务器上伺服思源笔记最简单的方案是通过 Docker 部署，镜像名称是 `b3log/siyuan`，目前没有版本标签，只有最新版。
{: id="20201227173504-opkavin"}

搭建若遇到问题请到[思源笔记 Issues](https://github.com/siyuan-note/siyuan/issues) 上反馈，谢谢。
{: id="20201227200109-1swwwt1"}

## 文件结构
{: id="20201227174700-39dg7ur"}

整体程序位于 `/opt/siyuan/` 下，基本上就是 Electron 安装包 resources 文件夹下的结构：
{: id="20201227174134-hptryqy"}

* {: id="20201227174714-t7ew8rk"}appearance：图标、主题、多语言
* {: id="20201227174726-zztkpj2"}guide：帮助文档
* {: id="20201227174744-z7qo4j2"}stage：界面和静态资源
* {: id="20201227174842-tr8u7q0"}kernel：内核程序
{: id="20201227174705-0hl54vz"}

## 启动入口
{: id="20201227174908-s19h988"}

构建 Docker 镜像时已经设置了入口：`ENTRYPOINT [ "/opt/siyuan/kernel" ]`，所以使用 `docker run b3log/siyuan` 即可启动，但这样是不能正常使用的，还需要设置一些参数，具体请参考((20201225172241-ifgc4hm "这里"))。
{: id="20201227180709-rglrztt"}

必传的参数有：
{: id="20201227181257-ewvzfu0"}

* {: id="20201227201514-c7brkss"}`--conf` 指定配置文件夹路径，配置文件夹在宿主机上通过 `-v` 挂载到到容器中
* {: id="20201227201521-fs7hlwf"}`--resident` 指定为 true 常驻内存
{: id="20201227201453-yh75kqd"}

下面是一条启动命令示例：
{: id="20201227181436-2o8ssj3"}

`docker run -v conf_dir_host:conf_dir_container -v data_dir_host:data_dir_container -p 6806:6806 b3log/siyuan --resident=true --conf=conf_dir_container`
{: id="20201227181459-iml8ebp"}

* {: id="20201227193950-dp2hioi"}`conf_dir_host`：宿主机上的配置文件夹路径
* {: id="20201227194032-68h45ue"}`conf_dir_container`：容器内配置文件夹路径，和后面 `--conf` 指定成一样的
* {: id="20201227194227-k2r2xan"}`data_dir_host`：宿主机数据文件夹路径
* {: id="20201227194257-5idyo8v"}`data_dir_container`：容器内数据文件夹路径
* {: id="20201227194126-ys9bhgh"}配置文件 conf.json 内 box path 字段需要在 `data_dir_container` 路径下
{: id="20201227174657-0k1ruhd"}

为了简化，建议将 conf、data 文件夹路径在宿主机和容器上分别配置为一致的，比如：
{: id="20201227194509-hdnzkry"}

* {: id="20201227194543-y71qqsd"}`conf_dir_host` 和 `conf_dir_container` 配置为 /siyuan/conf
* {: id="20201227194623-2um0t40"}`data_dir_host` 和 `data_dir_container` 配置为 /siyuan/data
{: id="20201227194542-ejfp46x"}

对应的启动命令示例：
{: id="20201227194830-86gbz55"}

`docker run -v /siyuan/conf:/siyuan/conf -v /siyuan/data:/siyuan/data -p 6806:6806 b3log/siyuan --resident=true --conf=/siyuan/conf`
{: id="20201227194658-m3auzob"}

对应的 conf.json 中 box 配置示例：
{: id="20201227194736-2jetrvt"}

```json
{
   "url": "http://127.0.0.1:6806/siyuan/siyuan/思源笔记用户指南/",
   "name": "思源笔记用户指南",
   "auth": "",
   "user": "",
   "password": "",
   "path": "/siyuan/data/思源笔记用户指南"
}
```
{: id="20201227194842-2bfxv59"}

除了手动准备 conf.json，也可以通过调用内核 API 来打开/关闭文件夹。
{: id="20201227195523-uipfd0v"}

## 内核 API
{: id="20201227194925-7ipoiv6"}

### 打开文件夹
{: id="20201227195443-zxgp2sw"}

POST `/notebook/mount`，参数：
{: id="20201227195224-cnwhfri"}

* {: id="20201227195626-jsv1r80"}`url`：固定传入 `http://127.0.0.1:6806/siyuan/`，即 box.url
* {: id="20201227195644-i7xcm1g"}`path`：内核数据文件夹下的某个文件夹路径，即 box.path
{: id="20201227195500-v08m84n"}

### 关闭文件夹
{: id="20201227195737-xbkf95m"}

POST `/notebook/unmount`，参数：
{: id="20201227195742-df6gznf"}

* {: id="20201227195805-rq7h6m9"}`url`：固定传入 `http://127.0.0.1:6806/siyuan/`，即 box.url
{: id="20201227195758-qsyk4py"}


{: id="20201227173504-847cs1q" type="doc"}
