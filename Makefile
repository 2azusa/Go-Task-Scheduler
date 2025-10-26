# Go makefile

#export env
#basic information
ProjectAdmin := "pulseadmin"
ProjectNode := "pulsenode"

PROJECTBASE 	:= $(shell pwd)
PROJECTBIN 	:= $(PROJECTBASE)/bin
AdminConf := "$(PROJECTBIN)/admin"
NodeConf := "$(PROJECTBIN)/node"
TIMESTAMP   := $(shell /bin/date "+%F %T")

#change to deploy environment
AdminFile := ./admin/cmd/main.go
NodeFile := ./node/cmd/main.go
WebFile := ./admin/web

#compile ldflags
LDFLAGS		:= -s -w \
			   -X 'main.BuildGitBranch=$(shell git describe --all)' \
			   -X 'main.BuildGitRev=$(shell git rev-list --count HEAD)' \
			   -X 'main.BuildGitCommit=$(shell git rev-parse HEAD)' \
			   -X 'main.BuildDate=$(shell /bin/date "+%F %T")'


linux-dev: clean install-web build-web
	@echo "install linux amd64 dev version"
	@if [ ! -d $(AdminConf)/logs ]; then \
        mkdir -p $(AdminConf)/logs; \
    fi
	@if [ ! -d $(NodeConf)/logs ]; then \
        mkdir -p $(NodeConf)/logs; \
    fi
	@cp -r ./admin/conf $(AdminConf)
	@cp -r ./node/conf $(NodeConf)
	@echo "building project pulseadmin..."
	@CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -v -o $(PROJECTBIN)/$(ProjectAdmin) $(AdminFile)
	@echo "building project pulsenode..."
	@CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -v -o $(PROJECTBIN)/$(ProjectNode) $(NodeFile)
	@chmod +x $(PROJECTBIN)/$(ProjectAdmin)
	@chmod +x $(PROJECTBIN)/$(ProjectNode)

	@mv $(WebFile)/dist bin/
	@echo "build success."

install-web:
	@echo "install web node_modules..."
	cd $(WebFile)&&npm install

build-web:
	@echo "building web..."
	cd $(WebFile)&&yarn build

run-web:
	@echo "running web..."
	cd $(WebFile) &&yarn dev

local-dev: clean install-web
	@echo "install local dev version"
	@go mod tidy
	@if [ ! -d $(NodeConf)/logs ]; then \
        mkdir -p $(NodeConf)/logs; \
    fi
	@if [ ! -d $(AdminConf)/logs ]; then \
            mkdir -p $(AdminConf)/logs; \
    fi
	@cp -r ./admin/conf $(AdminConf)
	@cp -r ./node/conf $(NodeConf)
	@echo "building project pulseadmin..."
	@CGO_ENABLED=0 go build -v  -o $(PROJECTBIN)/$(ProjectAdmin) $(AdminFile)
	@echo "building project pulsenode..."
	@CGO_ENABLED=0 go build -v  -o $(PROJECTBIN)/$(ProjectNode) $(NodeFile)
	@chmod +x $(PROJECTBIN)/$(PROJECTNAME)
	@chmod +x $(PROJECTBIN)/$(ProjectNode)
	@echo "build success."

gitpush: clean fmt
	git add .
	git commit -m "$(m) changed at $(TIMESTAMP)"
	git push
fmt:
	@go fmt $(PROJECTBASE)/...
	@echo "hello"
	@go mod tidy

clean:
	@#echo $(PROJECTBIN)
	@rm -rf $(PROJECTBIN)/* &>/dev/null
depend:
	go mod download
gitpull: fmt
	git add .
	git commit -m "$(m) changed at $(TIMESTAMP)"
	git pull
.PHONY: fmt clean git



.PHONY: start-dev stop-dev restart-dev status-dev
# 启动所有开发服务
start-dev:
	@echo "--- 正在启动开发环境服务 ---"
	@./manage-admin.sh start
	@./manage-nodes.sh start
	@echo "--------------------------"

# 停止所有开发服务
stop-dev:
	@echo "--- 正在停止开发环境服务 ---"
	@./manage-admin.sh stop
	@./manage-nodes.sh stop
	@echo "--------------------------"

# 重启所有开发服务
restart-dev:
	@echo "--- 正在重启开发环境服务 ---"
	@./manage-admin.sh restart
	@./manage-nodes.sh restart
	@echo "--------------------------"

# 检查所有开发服务的状态
status-dev:
	@echo "--- 正在检查开发环境状态 ---"
	@./manage-nodes.sh status
	@echo "--------------------------"