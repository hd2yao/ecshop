# RocketMQ Broker Dockerfile
FROM apache/rocketmq:5.1.4

LABEL maintainer="Dysania <dysania@hd2yao>"

# RocketMQ Broker 端口
EXPOSE 10909 10911 10912

# 复制配置文件
COPY broker.conf /home/rocketmq/rocketmq-5.1.4/conf/broker.conf

# 启动 Broker
CMD ["sh", "mqbroker", "-n", "rocketmq-namesrv:9876", "-c", "/home/rocketmq/rocketmq-5.1.4/conf/broker.conf"]

