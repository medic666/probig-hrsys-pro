# 重打包调试运行操作指南
操作根目录：`/home/wyc/projects/probig`

## 1. 仅修改前端 → 重新打包运行
```bash
# 构建前端（输出到 server/embed/dist）
cd client && bunx vite build

# 重新编译 Go（嵌入新前端资源）
cd ../server && go build -o probig-server .

# 清理旧数据库（修改数据模型时执行）+ 启动服务
rm -f hr.db && ./probig-server
```

## 2. 仅修改后端 → 重新编译运行
```bash
cd server
go build -o probig-server . && ./probig-server
```

## 3. 修改数据模型后必须删除数据库
> 说明：SQLite AutoMigrate 无法兼容全部字段变更，删库重建最稳妥
```bash
cd server
rm -f hr.db && go build -o probig-server . && ./probig-server
```

## 4. 调试开发：前后端分离（前端热更新）
**终端1：启动后端**
```bash
cd server && rm -f hr.db && go run main.go
```
**终端2：启动前端开发服务（自带热更新，API代理至 :8080）**
```bash
cd client && bun dev
```
访问地址：http://localhost:5173

## 5. 一键命令：清理数据库 + 完整重建并启动
```bash
cd /home/wyc/projects/probig
cd client && bunx vite build && cd ../server && rm -f hr.db && go build -o probig-server . && ./probig-server
```

# 场景速查表
| 场景 | 执行命令 |
| ---- | ---- |
| 仅改前端 | `cd client && bunx vite build && cd ../server && go build -o probig-server .` |
| 仅改后端 | `cd server && go build -o probig-server . && ./probig-server` |
| 修改数据模型 | 在上述命令基础上增加 `rm -f hr.db` |
| 日常调试开发 | 前后端分离启动：`go run main.go` + `bun dev` |
