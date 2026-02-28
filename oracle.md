# 项目规范

这个项目是 golang + html 的混合项目。

第一原则是如无必要，勿增实体。每次基于最小实现改动。

## 设计

按照 https://github.com/likec4/likec4 的规范，生成项目架构，导出到 likec4.c4

## 前端开发

html 部分放置在web目录中，使用 tailwindcss 作为前端框架，需使用响应式设计，适配移动端和PCweb端。

## 后端开发

使用 go.mod 对应的版本的语言开发，不要使用过时的API。

web使用GoFrame作为主要开发框架。前端路由在`web/api/routes.go`中。


## 测试

## 部署

## 功能完成后

更改内容写入 history.md
如果 web/api 层面有变动，则按照openAPI3的标准，更新 openAPI3.yaml
开发完成后，需要更新对应markdown文档