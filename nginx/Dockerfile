FROM nginx:alpine

# 复制自定义的 nginx 配置文件
COPY nginx.conf /etc/nginx/nginx.conf

# 创建 SSL 目录（如果需要的话）
RUN mkdir -p /etc/nginx/ssl

# 设置时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 80 443

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]