# vela-beat/process
vela 获取 系统 进程信息的方法

## vela.ps.all(condition...)
- 利用condition过滤找出相关进程信息
```lua
    vela.ps.all("exe re *java*" , "pid > 10").pipe(function(process)
        print(process) 
    end)
```

## vela.ps.snapshot(bool)
- 快照监控进程是否新增或者删除 参数:是否开启上报事件
- snapshot.sync() //同步快照数据到中心端
- snapshot.on_create(pipe) // 程序新建
- snapshot.on_delete(pipe) // 程序关闭 
- snapshot.on_update(pipe) // 程序更新
- snapshot.poll(time)      // 定时监控
```lua
    local cnd  = vela.cndf("exe re " ,
        "/usr/local/ssoc/ssc" ,
        "/usr/local/zabbix_new/sbin/zabbix_agent_new",
        "/opt/google/chrome/chrome",
    )
    local snap = vela.ps.snapshot(true) -- 报告中心端
    snap.ignore(cnd)
    snap.sync()
    snap.on_create(kfk , function(process) end)
    snap.on_update(kfk , function(process) end)
    snap.on_delete(kfk , function(process) end)
    snap.poll(5)
```


## 其他方法
- process = vela.ps.pid(int)
- summary = vela.ps.exe(string)
- summary = vela.ps.cmd(string)
- summary = vela.ps.name(string)
- summary = vela.ps.user(string)
- summary = vela.ps.ppid(string)

## process 
- 相关的进程信息 支持以下字段的读取
- name
- pid
- ppid
- cmd
- cwd
- exe
- state
- args
- memory
- rss
- rss_pct
- share
- stime //start time 