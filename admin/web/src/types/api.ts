// ===================================================================
//  通用及 API 结构类型
// ===================================================================

/**
 * API 响应的通用外层结构。
 */
export interface ApiResponse<T> {
  code: number;
  data: T;
  msg: string;
}

/**
 * 分页查询结果的通用结构。
 */
export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}

// ===================================================================
//  用户与认证 (User & Auth)
// ===================================================================

export enum UserRole {
  Normal = 1, // 普通用户
  Admin = 2, // 管理员
}

export interface User {
  id: number;
  username: string;
  password: number;
  email: string;
  role: UserRole;
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
  role?: UserRole;
}

export interface UserSearchRequest {
  page?: number;
  pageSize?: number;
  id?: number;
  username?: string;
  email?: string;
  role?: number;
}

// ===================================================================
//  任务 (Job)
// ===================================================================

export enum JobType {
  Command = 1,
  HTTP = 2,
}

export enum HttpMethod {
  GET = 1,
  POST = 2,
  PUT = 3,
  DELETE = 4,
}

export enum JobAllocation {
  Manual = 1, // 手动分配
  Auto = 2, // 自动分配
}

export enum NotifyType {
  Mail = 1, // 发送邮件
  WebHook = 2, // 飞书消息卡片
}

export interface Job {
  id: number;
  name: string;
  command: string;
  scriptId?: string;
  timeout?: number;
  retryTimes?: number;
  retryInterval?: number;
  jobType: JobType;
  httpMethod?: HttpMethod;
  httpUrl?: string;
  notifyType: NotifyType;
  notifyTo?: string;
  spec: string;
  runOn?: string;
  note?: string;
  allocation: JobAllocation;
  status?: number;
  created?: number;
  updated?: number;
}

export interface JobLog {
  id: number;
  name: string;
  jobId: number;
  command: string;
  ip: string;
  hostname: string;
  nodeUuid: string;
  success: boolean;
  output: string;
  spec: string;
  retryTimes: number;
  startTime: number;
  endTime: number;
}

export interface JobLogSearchRequest {
  page?: number;
  pageSize?: number;
  name?: string;
  jobId?: number;
  nodeUuid?: string;
  success?: boolean;
}

export interface JobSearchRequest {
  page?: number;
  pageSize?: number;
  id?: number;
  name?: string;
  runOn?: string;
  jobType?: number;
  status?: number;
}

// ===================================================================
//  节点 (Node)
// ===================================================================

export enum NodeStatus {
  Connected = 1,
  Disconnected = 2,
}

export interface Node {
  id: number;
  pid: string;
  ip: string;
  hostname: string;
  uuid: string;
  version: string;
  status: NodeStatus;
  up: number;
  down: number;
}

export interface NodeSearchResult {
  node: Node;
  jobCount: number;
}

export interface NodeSearchRequest {
  page?: number;
  pageSize?: number;
  ip?: string;
  uuid?: string;
  up?: number;
  status?: number;
}

// ===================================================================
//  脚本 (Script)
// ===================================================================

export interface Script {
  id?: number;
  name: string;
  command: string;
  created?: number;
  updated?: number;
}

export interface ScriptSearchRequest {
  page?: number;
  pageSize?: number;
  id?: number;
  name?: string;
}

// ===================================================================
//  统计 (Statistics)
// ===================================================================

export interface SystemStatistics {
  normalNodeCount: number;
  failNodeCount: number;
  jobExcSuccessCount: number;
  jobRunningCount: number;
  jobExcFailCount: number;
}

export interface DateCount {
  date: string;
  count: number;
}

export interface WeekStatistics {
  successDateCount: DateCount[];
  failDateCount: DateCount[];
}

export interface ServerInfo {
  os: Os;
  cpu: Cpu;
  ram: Ram;
  disk: Disk;
}

interface Os {
  goos: string;
  numCpu: number;
  compiler: string;
  goVersion: string;
  numGoroutine: number;
}

interface Cpu {
  cpus: number[];
  cores: number;
}

interface Ram {
  usedMb: number;
  totalMb: number;
  usedPercent: number;
}

interface Disk {
  usedMb: number;
  usedGb: number;
  totalMb: number;
  totalGb: number;
  usedPercent: number;
}
