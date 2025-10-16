-- -------------------------------------------------------------
-- Pulse - A Distributed Task Scheduling System
-- Database Schema
--
-- Author: Deng Zihao
-- Version: 1.0
-- -------------------------------------------------------------


CREATE DATABASE IF NOT EXISTS `pulse` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `pulse`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for jobs
-- ----------------------------
DROP TABLE IF EXISTS `jobs`;
CREATE TABLE `jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '任务ID, 主键',
  `name` varchar(64) NOT NULL COMMENT '任务名称',
  `command` text NOT NULL COMMENT '执行命令或HTTP URL',
  `script_id` varchar(256) DEFAULT NULL COMMENT '预设脚本ID列表 (JSON格式)',
  `timeout` bigint(20) DEFAULT 0 COMMENT '任务执行超时时间(秒), 0为不限制',
  `retry_times` int(4) DEFAULT 0 COMMENT '任务失败重试次数',
  `retry_interval` bigint(10) DEFAULT 0 COMMENT '任务失败重试间隔(秒)',
  `type` tinyint(1) NOT NULL COMMENT '任务类型 (1: 命令任务, 2: HTTP任务)',
  `http_method` tinyint(1) DEFAULT NULL COMMENT 'HTTP方法 (1: GET, 2: POST)',
  `notify_type` tinyint(1) NOT NULL COMMENT '通知类型',
  `status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否分配节点 (0: 未分配, 1: 已分配)',
  `notify_to` varchar(256) DEFAULT NULL COMMENT '通知对象用户ID列表 (JSON格式)',
  `spec` varchar(64) NOT NULL COMMENT 'cron定时表达式',
  `run_on` varchar(128) DEFAULT NULL COMMENT '运行节点UUID',
  `note` varchar(512) DEFAULT '' COMMENT '备注信息',
  `created` bigint(20) NOT NULL COMMENT '创建时间 (Unix时间戳)',
  `updated` bigint(20) DEFAULT 0 COMMENT '更新时间 (Unix时间戳)',
  PRIMARY KEY (`id`),
  KEY `idx_job_name` (`name`),
  KEY `idx_job_status` (`status`),
  KEY `idx_job_run_on` (`run_on`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';

-- ----------------------------
-- Table structure for job_logs
-- ----------------------------
DROP TABLE IF EXISTS `job_logs`;
CREATE TABLE `job_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '日志ID, 主键',
  `name` varchar(64) NOT NULL COMMENT '任务名称',
  `job_id` int(11) NOT NULL COMMENT '任务ID',
  `command` varchar(512) DEFAULT NULL COMMENT '执行命令',
  `ip` varchar(32) DEFAULT NULL COMMENT '节点IP',
  `hostname` varchar(32) DEFAULT NULL COMMENT '节点主机名',
  `node_uuid` varchar(128) NOT NULL COMMENT '节点唯一标识',
  `success` tinyint(1) NOT NULL COMMENT '是否成功 (0: 失败, 1: 成功)',
  `output` text COMMENT '执行输出',
  `spec` varchar(64) NOT NULL COMMENT 'cron定时表达式',
  `retry_times` int(4) DEFAULT 0 COMMENT '重试次数',
  `start_time` bigint(20) NOT NULL COMMENT '开始时间 (Unix时间戳)',
  `end_time` bigint(20) DEFAULT 0 COMMENT '结束时间 (Unix时间戳)',
  PRIMARY KEY (`id`),
  KEY `idx_job_log_name` (`name`),
  KEY `idx_job_log_id` (`job_id`),
  KEY `idx_job_log_node` (`node_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务日志表';

-- ----------------------------
-- Table structure for nodes
-- ----------------------------
DROP TABLE IF EXISTS `nodes`;
CREATE TABLE `nodes` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '节点ID, 主键',
  `pid` varchar(16) NOT NULL COMMENT 'Agent进程ID',
  `ip` varchar(32) DEFAULT '' COMMENT '节点IP地址',
  `hostname` varchar(64) DEFAULT '' COMMENT '节点主机名',
  `uuid` varchar(128) NOT NULL COMMENT '节点唯一标识',
  `version` varchar(64) DEFAULT '' COMMENT 'Agent版本号',
  `status` tinyint(1) DEFAULT NULL COMMENT '状态 (1: 连接成功, 2: 连接失败)',
  `up` bigint(20) NOT NULL COMMENT '上线时间 (Unix时间戳)',
  `down` bigint(20) DEFAULT 0 COMMENT '下线时间 (Unix时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_node_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='执行节点表';

-- ----------------------------
-- Table structure for scripts
-- ----------------------------
DROP TABLE IF EXISTS `scripts`;
CREATE TABLE `scripts` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '脚本ID, 主键',
  `name` varchar(256) NOT NULL COMMENT '脚本名称',
  `command` text NOT NULL COMMENT '脚本执行的命令',
  `created` bigint(20) NOT NULL COMMENT '创建时间 (Unix时间戳)',
  `updated` bigint(20) DEFAULT 0 COMMENT '更新时间 (Unix时间戳)',
  PRIMARY KEY (`id`),
  KEY `idx_script_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预设脚本表';

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '用户ID, 主键',
  `username` varchar(128) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码',
  `email` varchar(64) DEFAULT '' COMMENT '邮箱',
  `role` tinyint(1) DEFAULT 1 COMMENT '角色 (1: 普通用户, 2: 管理员)',
  `created` bigint(20) NOT NULL COMMENT '创建时间 (Unix时间戳)',
  `updated` bigint(20) DEFAULT 0 COMMENT '更新时间 (Unix时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

SET FOREIGN_KEY_CHECKS = 1;