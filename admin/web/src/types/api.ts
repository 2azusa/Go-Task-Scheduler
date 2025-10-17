// API Response Types
export interface ApiResponse<T> {
  code: number;
  data: T;
  msg: string;
}

export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

// User Types
export interface User {
  id: number;
  username: string;
  email: string;
  role: number; // 1: 普通用户, 2: 管理员
  created: number;
  updated: number;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
  role?: number;
}

// Job Types
export interface Job {
  id: number;
  name: string;
  command: string;
  script_id?: number[];
  timeout?: number;
  retry_times?: number;
  retry_interval?: number;
  job_type: number; // 1: 命令任务, 2: HTTP任务
  http_method?: number; // 1: GET, 2: POST
  notify_type: number;
  notify_to?: number[];
  spec: string; // Cron 表达式
  run_on?: string; // 节点 UUID
  note?: string;
  allocation: number; // 1: 手动分配, 2: 自动分配
  status?: number;
  created?: number;
  updated?: number;
}

export interface JobLog {
  id: number;
  name: string;
  job_id: number;
  command: string;
  ip: string;
  hostname: string;
  node_uuid: string;
  success: boolean;
  output: string;
  spec: string;
  retry_times: number;
  start_time: number;
  end_time: number;
}

export interface JobSearchRequest {
  page?: number;
  page_size?: number;
  id?: number;
  name?: string;
  run_on?: string;
  job_type?: number;
  status?: number;
}

// Node Types
export interface Node {
  id: number;
  pid: string;
  ip: string;
  hostname: string;
  uuid: string;
  version: string;
  status: number; // 1: 连接成功, 2: 连接失败
  up: number;
  down: number;
  job_count: number;
}

export interface NodeSearchRequest {
  page?: number;
  page_size?: number;
  ip?: string;
  uuid?: string;
  up?: number;
  status?: number;
}

// Script Types
export interface Script {
  id?: number;
  name: string;
  command: string;
  created?: number;
  updated?: number;
}

export interface ScriptSearchRequest {
  page?: number;
  page_size?: number;
  id?: number;
  name?: string;
}

// Statistics Types
export interface SystemStatistics {
  normal_node_count: number;
  fail_node_count: number;
  job_exc_success_count: number;
  job_running_count: number;
  job_exc_fail_count: number;
}

export interface DateCount {
  date: string;
  count: string;
}

export interface WeekStatistics {
  success_date_count: DateCount[];
  fail_date_count: DateCount[];
}
