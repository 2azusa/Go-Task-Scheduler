-- -------------------------------------------------------------
-- Pulse - A Distributed Task Scheduling System
-- Test Data Insertion
--
-- Author: Deng Zihao
-- Version: 1.0
-- -------------------------------------------------------------

USE `pulse`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Records for users
-- ----------------------------
INSERT INTO `users` (`id`, `username`, `password`, `email`, `role`, `created`, `updated`) VALUES
(1, 'admin', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'admin@example.com', 2, UNIX_TIMESTAMP(), 0),
(2, 'testuser', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'test@example.com', 1, UNIX_TIMESTAMP(), 0),
(3, 'john.doe', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'john.doe@example.com', 1, UNIX_TIMESTAMP(), 0),
(4, 'jane.smith', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'jane.smith@example.com', 1, UNIX_TIMESTAMP(), 0),
(5, 'alice', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'alice@example.com', 1, UNIX_TIMESTAMP(), 0),
(6, 'bob', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'bob@example.com', 1, UNIX_TIMESTAMP(), 0),
(7, 'charlie', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'charlie@example.com', 1, UNIX_TIMESTAMP(), 0),
(8, 'david', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'david@example.com', 1, UNIX_TIMESTAMP(), 0),
(9, 'emily', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'emily@example.com', 1, UNIX_TIMESTAMP(), 0),
(10, 'frank', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'frank@example.com', 1, UNIX_TIMESTAMP(), 0),
(11, 'grace', '$2a$10$TvjTFa5H4ScOAvj/qiqeb.lfIwVXSnB3EgU2CmfyupL3tk4irv/A6', 'grace@example.com', 1, UNIX_TIMESTAMP(), 0);

-- ----------------------------
-- Records for scripts
-- ----------------------------
INSERT INTO `scripts` (`id`, `name`, `command`, `created`, `updated`) VALUES
(1, 'List Files', 'ls -la', UNIX_TIMESTAMP(), 0),
(2, 'System Info', 'uname -a', UNIX_TIMESTAMP(), 0),
(3, 'Disk Usage', 'df -h', UNIX_TIMESTAMP(), 0),
(4, 'Memory Usage', 'free -m', UNIX_TIMESTAMP(), 0),
(5, 'Running Processes', 'ps aux', UNIX_TIMESTAMP(), 0),
(6, 'Uptime', 'uptime', UNIX_TIMESTAMP(), 0),
(7, 'Network Stats', 'netstat -tuln', UNIX_TIMESTAMP(), 0),
(8, 'Check Port 80', 'nc -zv localhost 80', UNIX_TIMESTAMP(), 0),
(9, 'Check Port 443', 'nc -zv localhost 443', UNIX_TIMESTAMP(), 0),
(10, 'Hello World', 'echo "Hello World"', UNIX_TIMESTAMP(), 0),
(11, 'Create Temp File', 'touch /tmp/tempfile', UNIX_TIMESTAMP(), 0);

-- ----------------------------
-- Records for nodes
-- ----------------------------
-- INSERT INTO `nodes` (`id`, `pid`, `ip`, `hostname`, `uuid`, `version`, `status`, `up`, `down`) VALUES
-- (1, '12345', '192.168.1.10', 'node-1', 'a1b2c3d4-e5f6-7890-1234-567890abcdef', '1.0.0', 1, UNIX_TIMESTAMP(), 0),
-- (2, '54321', '192.168.1.11', 'node-2', 'fedcba09-8765-4321-0987-654321fedcba', '1.0.1', 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- (3, '11111', '192.168.1.12', 'node-3', '11111111-1111-1111-1111-111111111111', '1.0.2', 1, UNIX_TIMESTAMP(), 0),
-- (4, '22222', '192.168.1.13', 'node-4', '22222222-2222-2222-2222-222222222222', '1.0.2', 1, UNIX_TIMESTAMP(), 0),
-- (5, '33333', '192.168.1.14', 'node-5', '33333333-3333-3333-3333-333333333333', '1.0.3', 1, UNIX_TIMESTAMP(), 0),
-- (6, '44444', '192.168.1.15', 'node-6', '44444444-4444-4444-4444-444444444444', '1.0.3', 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- (7, '55555', '192.168.1.16', 'node-7', '55555555-5555-5555-5555-555555555555', '1.1.0', 1, UNIX_TIMESTAMP(), 0),
-- (8, '66666', '192.168.1.17', 'node-8', '66666666-6666-6666-6666-666666666666', '1.1.0', 1, UNIX_TIMESTAMP(), 0),
-- (9, '77777', '192.168.1.18', 'node-9', '77777777-7777-7777-7777-777777777777', '1.1.1', 1, UNIX_TIMESTAMP(), 0),
-- (10, '88888', '192.168.1.19', 'node-10', '88888888-8888-8888-8888-888888888888', '1.1.1', 1, UNIX_TIMESTAMP(), 0),
-- (11, '99999', '192.168.1.20', 'node-11', '99999999-9999-9999-9999-999999999999', '1.2.0', 1, UNIX_TIMESTAMP(), 0);

-- ----------------------------
-- Records for jobs
-- ----------------------------
INSERT INTO `jobs` (`id`, `name`, `command`, `script_id`, `timeout`, `retry_times`, `retry_interval`, `type`, `http_method`, `notify_type`, `status`, `notify_to`, `spec`, `run_on`, `note`, `created`, `updated`) VALUES
(1, 'Daily Backup', '/scripts/backup.sh', NULL, 3600, 3, 60, 1, NULL, 1, 1, '[1,2]', '0 0 1 * * *', 'a1b2c3d4-e5f6-7890-1234-567890abcdef', 'Daily database backup job.', UNIX_TIMESTAMP(), 0),
(2, 'API Health Check', 'http://api.example.com/health', NULL, 30, 5, 10, 2, 1, 1, 1, '[1]', '*/5 * * * * *', 'a1b2c3d4-e5f6-7890-1234-567890abcdef', 'Checks the health of the main API every 5 seconds.', UNIX_TIMESTAMP(), 0),
(3, 'Update System', 'apt-get update && apt-get upgrade -y', NULL, 7200, 0, 0, 1, NULL, 1, 1, '[1]', '0 4 * * 0', NULL, 'Updates the system every Sunday at 4 AM.', UNIX_TIMESTAMP(), 0),
(4, 'POST to Webhook', 'http://webhook.example.com/data', NULL, 60, 2, 30, 2, 2, 1, 1, '[2]', '0 30 2 * * *', 'fedcba09-8765-4321-0987-654321fedcba', 'Sends a POST request to a webhook.', UNIX_TIMESTAMP(), 0),
(5, 'Clean Temp Files', 'rm -rf /tmp/*', NULL, 600, 0, 0, 1, NULL, 0, 1, NULL, '0 3 * * *', NULL, 'Cleans the temp directory every day at 3 AM.', UNIX_TIMESTAMP(), 0),
(6, 'Ping Google', 'ping -c 4 google.com', NULL, 120, 1, 30, 1, NULL, 1, 1, '[1]', '0 * * * * *', NULL, 'Pings Google every hour.', UNIX_TIMESTAMP(), 0),
(7, 'Check DB Connection', 'mysqladmin -u root -p ping', NULL, 10, 3, 10, 1, NULL, 1, 1, '[1]', '*/10 * * * * *', NULL, 'Checks the database connection every 10 minutes.', UNIX_TIMESTAMP(), 0),
(8, 'GET Request to API', 'http://myapi.com/data', NULL, 20, 0, 0, 2, 1, 0, 1, NULL, '0 0/30 * * * ?', NULL, 'Makes a GET request every 30 minutes.', UNIX_TIMESTAMP(), 0),
(9, 'Sync Files', 'rsync -avz /source/ /destination/', NULL, 1800, 1, 60, 1, NULL, 1, 1, '[1]', '0 2 * * *', NULL, 'Syncs files every day at 2 AM.', UNIX_TIMESTAMP(), 0),
(10, 'Restart Service', 'systemctl restart myservice', NULL, 300, 0, 0, 1, NULL, 1, 1, '[1]', '0 0 1 1 *', NULL, 'Restarts a service on the first day of every month.', UNIX_TIMESTAMP(), 0);

-- ----------------------------
-- Records for job_logs
-- ----------------------------
INSERT INTO `job_logs` (`id`, `name`, `job_id`, `command`, `ip`, `hostname`, `node_uuid`, `success`, `output`, `spec`, `retry_times`, `start_time`, `end_time`) VALUES
(1, 'Daily Backup', 1, '/scripts/backup.sh', '192.168.1.10', 'node-1', 'a1b2c3d4-e5f6-7890-1234-567890abcdef', 1, 'Backup completed successfully.', '0 0 1 * * *', 0, UNIX_TIMESTAMP() - 3600, UNIX_TIMESTAMP() - 3540),
(2, 'API Health Check', 2, 'http://api.example.com/health', '192.168.1.10', 'node-1', 'a1b2c3d4-e5f6-7890-1234-567890abcdef', 0, 'Connection timed out.', '*/5 * * * * *', 3, UNIX_TIMESTAMP() - 600, UNIX_TIMESTAMP() - 570),
(3, 'API Health Check', 2, 'http://api.example.com/health', '192.168.1.10', 'node-1', 'a1b2c3d4-e5f6-7890-1234-567890abcdef', 1, '{"status": "ok"}', '*/5 * * * * *', 0, UNIX_TIMESTAMP() - 300, UNIX_TIMESTAMP() - 299),
(4, 'POST to Webhook', 4, 'http://webhook.example.com/data', '192.168.1.11', 'node-2', 'fedcba09-8765-4321-0987-654321fedcba', 1, 'Webhook received data.', '0 30 2 * * *', 0, UNIX_TIMESTAMP() - 86400, UNIX_TIMESTAMP() - 86395),
(5, 'Clean Temp Files', 5, 'rm -rf /tmp/*', '192.168.1.12', 'node-3', '11111111-1111-1111-1111-111111111111', 1, 'Temp files cleaned.', '0 3 * * *', 0, UNIX_TIMESTAMP() - 172800, UNIX_TIMESTAMP() - 172790),
(6, 'Ping Google', 6, 'ping -c 4 google.com', '192.168.1.13', 'node-4', '22222222-2222-2222-2222-222222222222', 1, 'PING google.com (142.250.191.78): 56 data bytes', '0 * * * * *', 0, UNIX_TIMESTAMP() - 3600, UNIX_TIMESTAMP() - 3590),
(7, 'Check DB Connection', 7, 'mysqladmin -u root -p ping', '192.168.1.14', 'node-5', '33333333-3333-3333-3333-333333333333', 0, 'mysqladmin: connect to server at \'localhost\' failed', '*/10 * * * * *', 2, UNIX_TIMESTAMP() - 600, UNIX_TIMESTAMP() - 590),
(8, 'GET Request to API', 8, 'http://myapi.com/data', '192.168.1.15', 'node-6', '44444444-4444-4444-4444-444444444444', 1, '{"data": "some data"}', '0 0/30 * * * ?', 0, UNIX_TIMESTAMP() - 1800, UNIX_TIMESTAMP() - 1799),
(9, 'Sync Files', 9, 'rsync -avz /source/ /destination/', '192.168.1.16', 'node-7', '55555555-5555-5555-5555-555555555555', 1, 'Files synced successfully.', '0 2 * * *', 0, UNIX_TIMESTAMP() - 86400, UNIX_TIMESTAMP() - 86300),
(10, 'Restart Service', 10, 'systemctl restart myservice', '192.168.1.17', 'node-8', '66666666-6666-6666-6666-666666666666', 1, 'Service restarted.', '0 0 1 1 *', 0, UNIX_TIMESTAMP() - 2592000, UNIX_TIMESTAMP() - 2591900),
(11, 'Update System', 11, 'apt-get update && apt-get upgrade -y', '192.168.1.18', 'node-9', '77777777-7777-7777-7777-777777777777', 1, 'System updated.', '0 4 * * 0', 0, UNIX_TIMESTAMP() - 604800, UNIX_TIMESTAMP() - 604200);


SET FOREIGN_KEY_CHECKS = 1;
