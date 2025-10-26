// =================================================================
// Common - 通用接口
// =================================================================

/**
 * 定义用于查询的分页信息。
 */
export interface PageInfo {
    /**
     * 要检索的页码。
     */
    page?: number;
    /**
     * 每页要检索的项目数。
     */
    pageSize?: number;
}

/**
 * 定义删除操作的负载。
 */
export interface Delete {
    /**
     * 一个包含待删除数字 ID 的数组。
     */
    ids: number[];
}

/**
 * 所有 API 响应的通用包装器。
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


// =================================================================
// User - 用户相关接口
// =================================================================

/**
 * 定义用户可以拥有的角色。
 */
export enum Role {
    /**
     * 具有普通权限的标准用户。
     */
    Normal = 1,
    /**
     * 具有提升权限的管理员。
     */
    Admin = 2,
}

/**
 * 用户登录的负载。
 */
export interface Login {
    /**
     * 用户名，长度必须在 2 到 20 个字符之间。
     */
    username: string;
    /**
     * 密码，必须是字母数字，长度在 4 到 20 个字符之间。
     */
    password: string;
}

/**
 * 用户注册的负载。
 */
export interface Register {
    /**
     * 用户名，长度必须在 2 到 20 个字符之间。
     */
    username: string;
    /**
     * 密码，必须是字母数字，长度在 4 到 20 个字符之间。
     */
    password: string;
    /**
     * 用户的电子邮件地址。
     */
    email: string;
    /**
     * 用户的角色。如果未提供，则默认为 'Normal' (1)。
     */
    role?: Role;
}

/**
 * 用户更改密码的负载。
 */
export interface ChangePassword {
    /**
     * 当前密码。必须是字母数字，长度在 4 到 20 个字符之间。
     */
    password: string;
    /**
     * 新密码。必须是字母数字，长度在 4 到 20 个字符之间。
     */
    newPassword: string;
}

/**
 * 更新用户信息的负载。只有提供的字段才会被更新。
 */
export interface UserUpdate {
    /**
     * 新的用户名，长度必须在 2 到 20 个字符之间。
     */
    username?: string;
    /**
     * 新的电子邮件地址。
     */
    email?: string;
    /**
     * 用户的新角色。
     */
    role?: Role;
}

/**
 * 查找用户的负载。如果省略 'id'，则根据认证令牌返回当前用户。
 */
export interface UserFind {
    /**
     * 要查找用户的唯一 ID。
     */
    id?: number;
}

/**
 * 定义用于搜索用户的可选查询参数。
 */
export interface UserSearch extends PageInfo {
    /**
     * 按用户 ID 过滤。
     */
    id?: number;
    /**
     * 按用户名过滤（支持部分匹配）。
     */
    username?: string;
    /**
     * 按电子邮件地址过滤。
     */
    email?: string;
    /**
     * 按用户角色过滤。
     */
    role?: Role;
}

/**
 * 表示一个用户的信息。
 */
export interface User {
    id: number;
    username: string;
    email: string;
    role: Role;
    created: number;
    updated: number;
    password?: string;
}

// =================================================================
// Job - 任务相关接口
// =================================================================

/**
 * 定义任务如何分配到节点。
 */
export enum Allocation {
    /**
     * 任务被手动分配到特定节点。
     */
    ManualAllocation = 1,
    /**
     * 系统自动将任务分配到可用节点。
     */
    AutoAllocation   = 2,
}

/**
 * 定义 HTTP 类型任务的 HTTP 方法。
 */
export enum HttpMethod {
	HttpMethodGet = 1, // GET 请求
	HttpMethodPost = 2 // POST 请求
}

/**
 * 定义任务的分配状态。
 */
export enum JobStatus {
    /**
     * 任务尚未分配到节点。
     */
    JobStatusNotAssigned = 0,
    /**
     * 任务已被分配到节点。
     */
    JobStatusAssigned = 1,
}

/**
 * 定义任务完成提醒的通知类型。
 */
export enum NotifyType {
	NotifyTypeMail = 1, // 发送邮件
	NotifyTypeWebHook = 2 // 发送飞书消息卡片
}

/**
 * 定义任务的类型。
 */
export enum JobType {
    /**
     * 命令行/shell 脚本任务。
     */
    JobTypeCmd = 1,
    /**
     * HTTP 请求任务。
     */
	JobTypeHttp = 2,
}

/**
 * 创建或更新任务的负载。更新时，仅修改提供的字段。
 */
export interface JobUpdate {
    /**
     * 要更新任务的唯一 ID。如果省略，将创建一个新任务。
     */
    id?: number;
    /**
     * 任务的名称。
     */
    name: string;
    /**
     * 任务的类型（例如，命令行或 HTTP）。
     */
    jobType: JobType;
    /**
     * 需要执行的命令或要调用的 HTTP URL。
     */
    command: string;
    /**
     * 定义任务计划的 cron 表达式。
     */
    spec: string;
    /**
     * 指定任务如何分配到节点。
     */
    allocation: Allocation;
    /**
     * 运行任务的节点的 UUID。手动分配时必须指定。
     */
    runOn?: string;
    /**
     * 执行超时时间（秒）。默认为 0（无超时）。
     */
    timeout?: number;
    /**
     * 任务失败后的重试次数。默认为 0。
     */
    retryTimes?: number;
    /**
     * 两次重试之间的间隔时间（秒）。如果为 0，'retryTimes' 将被强制设为 1。
     */
    retryInterval?: number;
    /**
     * 用于 HTTP 任务的 HTTP 方法。
     */
    httpMethod?: HttpMethod;
    /**
     * 与此任务关联的预设脚本 ID 数组。
     */
    scriptId?: number[];
    /**
     * 任务结果的通知方式。
     */
    notifyType: NotifyType;
    /**
     * 任务完成后需要通知的用户 ID 数组。
     */
    notifyTo: number[];
    /**
     * 任务的可选备注或描述。
     */
    note?: string;
}

/**
 * 定义用于搜索任务的可选查询参数。
 */
export interface JobSearch extends PageInfo {
    id?: number;
    name?: string;
    /**
     * 任务运行的节点名称。
     */
    runOn?: string;
    type?: JobType;
    /**
     * 按任务是否已分配到节点进行过滤。
     */
    status?: JobStatus;
}

/**
 * 定义用于搜索任务日志的可选查询参数。
 */
export interface JobLogSearch extends PageInfo {
    name?: string;
    jobId?: number;
    nodeUuid?: string;
    /**
     * 按执行状态过滤（true 为成功，false 为失败）。
     */
    success?: boolean;
}

/**
 * 按 ID 查找特定任务的负载。
 */
export interface JobFind {
    id: number;
}

/**
 * 在特定节点上触发一次性执行任务的负载。
 */
export interface JobOnce {
    jobId: number;
    nodeUuid: string;
}

/**
 * 终止正在运行的任务实例的负载。
 */
export interface JobKill {
    jobId: number;
    nodeUuid: string;
}

/**
 * 表示一个计划任务的信息。
 */
export interface Job {
    id: number;
    name: string;
    /**
     * 任务类型 (1: Shell, 2: HTTP)。
     */
    jobType: JobType;
    command: string;
    /**
     * cron 调度表达式。
     */
    spec: string;
    /**
     * 任务当前状态 (0: 未分配, 1: 已分配)。
     */
    status: JobStatus;
    note: string;
    /**
     * 任务分配到的节点的 UUID。
     */
    runOn: string;
    /**
     * HTTP 任务的 HTTP 方法 (1: GET, 2: POST)。
     */
    httpMethod: HttpMethod | null;
    timeout: number;
    retryTimes: number;
    retryInterval: number;
    /**
     * 通知类型（例如，1 表示邮件）。
     */
    notifyType: NotifyType;
    /**
     * 需要通知的用户 ID 数组。
     */
    notifyTo: number[] | null;
    /**
     * 关联的脚本 ID 数组。
     */
    scriptId: number[] | null;
    created: number;
    updated: number;
    /**
     * 任务运行所在节点的主机名。
     */
    hostname: string;
    /**
     * 任务运行所在节点的 IP 地址。
     */
    ip: string;

    cmd: string[] | null;
}

/**
 * 表示任务执行的日志条目。
 */
export interface JobLog {
    id: number;
    job_id: number;
    name: string;
    node_uuid: string;
    hostname: string;
    ip: string;
    spec: string;
    command: string;
    /**
     * 任务执行的输出。
     */
    output: string;
    /**
     * 指示执行是否成功。
     */
    success: boolean;
    /**
     * 本次执行的重试次数。
     */
    retry_times: number;
    start_time: number;
    end_time: number;
}

// =================================================================
// Node - 节点相关接口
// =================================================================

/**
 * 定义节点的连接状态。
 */
export enum NodeStatus {
    /**
     * 节点已连接并成功运行。
     */
    NodeConnSuccess = 1,
    /**
     * 节点连接失败或已丢失。
     */
	NodeConnFail = 2,
}

/**
 * 定义用于搜索节点的可选查询参数。
 */
export interface NodeSearch extends PageInfo {
    ip?: string;
    uuid?: string;
    /**
     * 按节点启动时间戳过滤。
     */
    up?: number;
    status?: NodeStatus;
}

/**
 * 按 UUID 查找特定节点的负载。
 */
export interface NodeFind {
    uuid: string;
}

/**
 * 表示一个工作节点。
 */
export interface Node {
    id: number;
    hostname: string;
    ip: string;
    pid: string;
    /**
     * 节点的状态 (1: 在线, 2: 下线)。
     */
    status: NodeStatus;
    uuid: string;
    version: string;
    /**
     * 当前分配给此节点的任务数。
     */
    jobCount: number;
    /**
     * 节点上线时的时间戳。
     */
    up: number;
    /**
     * 节点下线时的时间戳。
     */
    down: number;
}

// =================================================================
// Script - 脚本相关接口
// =================================================================

/**
 * 创建或更新脚本的负载。只有提供的字段会被更新。
 */
export interface ScriptUpdate {
    /**
     * 要更新脚本的唯一 ID。如果省略，将创建一个新脚本。
     */
    id?: number;
    name?: string;
    command?: string;
}

/**
 * 按 ID 查找特定脚本的负载。
 */
export interface ScriptFind {
    id: number;
}

/**
 * 定义用于搜索脚本的可选查询参数。
 */
export interface ScriptSearch extends PageInfo {
    id?: number;
    name?: string;
}

/**
 * 表示一个可重用的脚本。
 */
export interface Script {
    id: number;
    name: string;
    command: string;
    created: number;
    updated: number;

    cmd: string[] | null;
}

// =================================================================
// Statistics - 统计相关接口
// =================================================================

/**
 * 获取系统信息的负载。
 * 如果提供了 uuid，则返回该节点的信息；否则，返回管理服务器的信息。
 */
export interface GetSystem {
    /**
     * 目标节点的 UUID。
     */
    uuid?: string;
}

// =================================================================
// Auth - 认证接口响应
// =================================================================

/**
 * `POST /register` (用户注册) 的响应。
 */
export type RegisterResponse = ApiResponse<User>;

/**
 * `POST /login` (用户认证) 的响应。
 */
export type LoginResponse = ApiResponse<{
    token: string;
    user: User;
}>;


// =================================================================
// User - 用户接口响应
// =================================================================

/**
 * `POST /user/change_pw` (修改密码) 的响应。
 */
export type ChangePasswordResponse = ApiResponse<object>;

/**
 * `GET /user/find` (按 ID 或 token 查找用户) 的响应。
 */
export type GetUserResponse = ApiResponse<User>;

/**
 * `POST /user/search` (搜索用户) 的响应。
 */
export type SearchUserResponse = ApiResponse<PageResult<User>>;

/**
 * `POST /user/del` (删除用户) 的响应。
 */
export type DelUserResponse = ApiResponse<object>;

/**
 * `POST /user/update` (更新用户) 的响应。
 */
export type UpdateUserResponse = ApiResponse<{
    username: string;
    email: string;
    role: Role;
}>;

// =================================================================
// Statis - 统计接口响应
// =================================================================

export interface Cpu {
    /**
     * CPU 使用率百分比数组。
     */
    cpus: number[];
    /**
     * CPU 核心数。
     */
    cores: number;
}

export interface Ram {
    usedMb: number;
    totalMB: number;
    usedPercent: number;
}

export interface Disk {
    totalGb: number;
    totalMb: number;
    usedGb: number;
    usedMb: number;
    usedPercent: number;
}

export interface Os {
    /**
     * 操作系统（例如，"linux"）。
     */
    goos: string;
    numCpu: number;
    compiler: string;
    goVersion: string;
    numGoroutine: number;
}

/**
 * 表示节点或服务器的详细系统信息。
 */
export interface SystemInfo {
    cpu: Cpu;
    disk: Disk;
    os: Os;
    ram: Ram;
}

export interface DateCount {
    count: string;
    date: string;
}

/**
 * 表示过去一周任务执行的统计数据。
 */
export interface WeekStatistics {
    failDateCount: DateCount[];
    successDateCount: DateCount[];
}

/**
 * 表示今日统计数据的摘要。
 */
export interface TodayStatistics {
    normalNodeCount: number;
    failNodeCount: number;
    jobExcSuccessCount: number;
    jobRunningCount: number;
    jobExcFailCount: number;
}

/**
 * `GET /statis/today` (获取今日统计数据) 的响应。
 */
export type TodayStatisticsResponse = ApiResponse<TodayStatistics>;

/**
 * `GET /statis/week` (获取每周统计数据) 的响应。
 */
export type WeekStatisticsResponse = ApiResponse<WeekStatistics>;

/**
 * `GET /statis/system` (获取系统信息) 的响应。
 */
export type SystemInfoResponse = ApiResponse<SystemInfo>;


// =================================================================
// Node - 节点接口响应
// =================================================================

/**
 * `POST /node/search` (搜索节点) 的响应。
 */
export type SearchNodeResponse = ApiResponse<PageResult<Node>>;

/**
 * `POST /node/del` (删除节点) 的响应。
 */
export type DelNodeResponse = ApiResponse<object>;


// =================================================================
// Job - 任务接口响应
// =================================================================

/**
 * `POST /job/add` 或 `POST /job/update` (添加或更新任务) 的响应。
 */
export type AddJobResponse = ApiResponse<Job>;

/**
 * `GET /job/find` (按 ID 查找任务) 的响应。
 */
export type FindJobResponse = ApiResponse<Job>;

/**
 * `POST /job/search` (搜索任务) 的响应。
 */
export type SearchJobResponse = ApiResponse<PageResult<Job>>;

/**
 * `POST /job/del` (删除任务) 的响应。
 */
export type DelJobOnceResponse = ApiResponse<object>;

/**
 * `POST /job/log` (搜索任务日志) 的响应。
 */
export type JobLogResponse = ApiResponse<PageResult<JobLog>>;

/**
 * `POST /job/once` (立即执行任务) 的响应。
 */
export type RunJobOnceResponse = ApiResponse<object>;

/**
 * `POST /job/kill` (终止正在运行的任务) 的响应。
 */
export type KillJobOnceResponse = ApiResponse<object>;


// =================================================================
// Script - 脚本接口响应
// =================================================================

/**
 * `POST /script/add` 或 `POST /script/update` (添加或更新脚本) 的响应。
 */
export type AddScriptResponse = ApiResponse<Script>;

/**
 * `GET /script/find` (按 ID 查找脚本) 的响应。
 */
export type FindScriptResponse = ApiResponse<Script>;

/**
 * `POST /script/search` (搜索脚本) 的响应。
 */
export type SearchScriptResponse = ApiResponse<PageResult<Script>>;

/**
 * `POST /script/del` (删除脚本) 的响应。
 */
export type DelScriptResponse = ApiResponse<object>;
