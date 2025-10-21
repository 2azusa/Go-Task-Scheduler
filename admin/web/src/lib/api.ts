import type {
  ApiResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  User,
  UserSearchRequest,
  Job,
  JobLog,
  JobSearchRequest,
  JobLogSearchRequest,
  NodeSearchResult,
  NodeSearchRequest,
  Script,
  ScriptSearchRequest,
  SystemStatistics,
  WeekStatistics,
  ServerInfo,
  PageResult,
} from "@/types/api";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";

/**
 * API 客户端，用于与后端服务进行交互
 */
class ApiClient {
  /**
   * 从 localStorage 获取认证 token
   * @returns 返回 token 字符串，如果不存在则返回 null
   */
  private getToken(): string | null {
    return localStorage.getItem("token");
  }

  /**
   * 发送一个 API 请求
   * @template T - 期望的响应数据类型
   * @param {string} endpoint - API 的端点路径
   * @param {RequestInit} [options={}] - fetch 请求的配置选项
   * @returns {Promise<ApiResponse<T>>} 返回一个 Promise，解析为 API 响应
   * @throws {Error} 如果网络请求失败或响应状态码不是 2xx
   */
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const token = this.getToken();
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // 如果 token 存在且不是登录或注册接口，则添加 Authorization 头
    if (token && !endpoint.includes("/login") && !endpoint.includes("/register")) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const config: RequestInit = {
      ...options,
      headers,
    };

    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("API 请求失败，状态码:", response.status, "响应:", errorText);
      throw new Error(`请求失败，状态码 ${response.status}`);
    }

    const responseClone = response.clone();
    try {
      return await response.json();
    } catch (error) {
      console.error("Failed to parse JSON response:", error);
      const text = await responseClone.text();
      console.error("Raw response text:", text);
      throw error;
    }
  }

  // Auth APIs
  /**
   * 用户登录
   * @param {LoginRequest} data - 用户的登录凭据
   * @returns {Promise<ApiResponse<LoginResponse>>} 返回一个 Promise，该 Promise 解析为包含 token 的登录响应
   */
  async login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    return this.request<LoginResponse>("/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 注册一个新用户
   * @param {RegisterRequest} data - 用户的注册信息
   * @returns {Promise<ApiResponse<User>>} 返回一个 Promise，该 Promise 解析为新创建的用户信息
   */
  async register(data: RegisterRequest): Promise<ApiResponse<User>> {
    return this.request<User>("/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // User APIs
  /**
   * 获取当前登录用户的信息
   * @returns {Promise<ApiResponse<User>>} 返回一个 Promise，该 Promise 解析为当前用户的数据
   */
  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<User>("/user/find");
  }

  /**
   * 根据指定条件搜索用户
   * @param {UserSearchRequest} data - 用户的搜索条件
   * @returns {Promise<ApiResponse<PageResult<User>>>} 返回一个 Promise，该 Promise 解析为用户的分页列表
   */
  async searchUsers(
    data: UserSearchRequest
  ): Promise<ApiResponse<PageResult<User>>> {
    return this.request<PageResult<User>>("/user/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 更新用户信息
   * @param {Partial<User> & { id: number }} data - 要更新的用户信息 (至少需要包含 id)
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在用户信息更新后解析
   */
  async updateUser(
    data: Partial<User> & { id: number }
  ): Promise<ApiResponse<void>> {
    return this.request<void>("/user/update", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 更新当前用户的密码
   * @param {string} password - 当前密码
   * @param {string} newPassword - 新密码
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在密码更新后解析
   */
  async updatePassword(
    password: string,
    newPassword: string
  ): Promise<ApiResponse<void>> {
    return this.request<void>("/user/change_pw", {
      method: "POST",
      body: JSON.stringify({ password, newPassword }),
    });
  }

  /**
   * 删除一个或多个用户
   * @param {number[]} ids - 要删除的用户 ID 数组
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在用户被删除后解析
   */
  async deleteUsers(ids: number[]): Promise<ApiResponse<void>> {
    return this.request<void>("/user/del", {
      method: "POST",
      body: JSON.stringify({ ids }),
    });
  }

  // Job APIs
  /**
   * 根据指定条件搜索任务
   * @param {JobSearchRequest} data - 任务的搜索条件
   * @returns {Promise<ApiResponse<PageResult<Job>>>} 返回一个 Promise，该 Promise 解析为任务的分页列表
   */
  async searchJobs(data: JobSearchRequest): Promise<ApiResponse<PageResult<Job>>> {
    return this.request<PageResult<Job>>("/job/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 根据 ID 查找任务
   * @param {number} id - 要查找的任务 ID
   * @returns {Promise<ApiResponse<Job>>} 返回一个 Promise，该 Promise 解析为找到的任务
   */
  async findJobById(id: number): Promise<ApiResponse<Job>> {
    const params = new URLSearchParams({ id: id.toString() });
    return this.request<Job>(`/job/find?${params.toString()}`);
  }

  /**
   * 创建一个新任务
   * @param {Partial<Job>} data - 要创建的任务的详细信息
   * @returns {Promise<ApiResponse<Job>>} 返回一个 Promise，该 Promise 解析为新创建的任务
   */
  async createJob(data: Partial<Job>): Promise<ApiResponse<Job>> {
    return this.request<Job>("/job/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 更新一个现有任务
   * @param {Partial<Job>} data - 任务的更新信息
   * @returns {Promise<ApiResponse<Job>>} 返回一个 Promise，该 Promise 解析为更新后的任务
   */
  async updateJob(data: Partial<Job>): Promise<ApiResponse<Job>> {
    return this.request<Job>("/job/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 删除一个或多个任务
   * @param {number[]} ids - 要删除的任务 ID 数组
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在任务被删除后解析
   */
  async deleteJobs(ids: number[]): Promise<ApiResponse<void>> {
    return this.request<void>("/job/del", {
      method: "POST",
      body: JSON.stringify({ ids }),
    });
  }

  /**
   * 获取任务的执行日志
   * @param {JobLogSearchRequest} data - 获取任务日志的查询条件
   * @returns {Promise<ApiResponse<PageResult<JobLog>>>} 返回一个 Promise，该 Promise 解析为任务日志的分页列表
   */
  async getJobLogs(
    data: JobLogSearchRequest
  ): Promise<ApiResponse<PageResult<JobLog>>> {
    return this.request<PageResult<JobLog>>("/job/log", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 在指定节点上单次执行任务
   * @param {number} jobId - 要执行的任务 ID
   * @param {string} nodeUuid - 执行任务的节点 UUID
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在任务执行被触发后解析
   */
  async executeJobOnce(
    jobId: number,
    nodeUuid: string
  ): Promise<ApiResponse<void>> {
    return this.request<void>("/job/once", {
      method: "POST",
      body: JSON.stringify({ jobId, nodeUuid }),
    });
  }

  /**
   * 终止一个正在运行的任务
   * @param {number} jobId - 要终止的任务 ID
   * @param {string} nodeUuid - 任务所在的节点 UUID
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在任务终止命令发送后解析
   */
  async killJob(jobId: number, nodeUuid: string): Promise<ApiResponse<void>> {
    return this.request<void>("/job/kill", {
      method: "POST",
      body: JSON.stringify({ jobId, nodeUuid }),
    });
  }

  // Node APIs
  /**
   * 根据指定条件搜索节点
   * @param {NodeSearchRequest} data - 节点的搜索条件
   * @returns {Promise<ApiResponse<PageResult<NodeSearchResult>>>} 返回一个 Promise，该 Promise 解析为节点的分页列表
   */
  async searchNodes(
    data: NodeSearchRequest
  ): Promise<ApiResponse<PageResult<NodeSearchResult>>> {
    return this.request<PageResult<NodeSearchResult>>("/node/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 删除一个节点
   * @param {string} uuid - 要删除节点的 UUID
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在节点被删除后解析
   */
  async deleteNode(uuid: string): Promise<ApiResponse<void>> {
    return this.request<void>("/node/del", {
      method: "POST",
      body: JSON.stringify({ uuid }),
    });
  }

  // Script APIs
  /**
   * 根据指定条件搜索脚本
   * @param {ScriptSearchRequest} data - 脚本的搜索条件
   * @returns {Promise<ApiResponse<PageResult<Script>>>} 返回一个 Promise，该 Promise 解析为脚本的分页列表
   */
  async searchScripts(
    data: ScriptSearchRequest
  ): Promise<ApiResponse<PageResult<Script>>> {
    return this.request<PageResult<Script>>("/script/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 根据 ID 查找脚本
   * @param {number} id - 要查找的脚本 ID
   * @returns {Promise<ApiResponse<Script>>} 返回一个 Promise，该 Promise 解析为找到的脚本
   */
  async findScriptById(id: number): Promise<ApiResponse<Script>> {
    const params = new URLSearchParams({ id: id.toString() });
    return this.request<Script>(`/script/find?${params.toString()}`);
  }

  /**
   * 创建一个新脚本
   * @param {Script} data - 要创建的脚本的详细信息
   * @returns {Promise<ApiResponse<Script>>} 返回一个 Promise，该 Promise 解析为新创建的脚本
   */
  async createScript(data: Script): Promise<ApiResponse<Script>> {
    return this.request<Script>("/script/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 更新一个现有脚本
   * @param {Partial<Script>} data - 脚本的更新信息
   * @returns {Promise<ApiResponse<Script>>} 返回一个 Promise，该 Promise 解析为更新后的脚本
   */
  async updateScript(data: Partial<Script>): Promise<ApiResponse<Script>> {
    return this.request<Script>("/script/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * 删除一个或多个脚本
   * @param {number[]} ids - 要删除的脚本 ID 数组
   * @returns {Promise<ApiResponse<void>>} 返回一个 Promise，在脚本被删除后解析
   */
  async deleteScripts(ids: number[]): Promise<ApiResponse<void>> {
    return this.request<void>("/script/del", {
      method: "POST",
      body: JSON.stringify({ ids }),
    });
  }

  // Statistics APIs
  /**
   * 获取当日的系统统计数据
   * @returns {Promise<ApiResponse<SystemStatistics>>} 返回一个 Promise，该 Promise 解析为当日的系统统计数据
   */
  async getTodayStatistics(): Promise<ApiResponse<SystemStatistics>> {
    return this.request<SystemStatistics>("/statis/today");
  }

  /**
   * 获取本周的系统统计数据
   * @returns {Promise<ApiResponse<WeekStatistics>>} 返回一个 Promise，该 Promise 解析为本周的统计数据
   */
  async getWeekStatistics(): Promise<ApiResponse<WeekStatistics>> {
    return this.request<WeekStatistics>("/statis/week");
  }

  /**
   * 获取系统信息
   * @param {string} [uuid] - 可选的节点 UUID。如果提供，则获取指定节点的信息；否则获取管理服务器的信息
   * @returns {Promise<ApiResponse<ServerInfo>>} 返回一个 Promise，该 Promise 解析为服务器信息
   */
  async getSystemInfo(uuid?: string): Promise<ApiResponse<ServerInfo>> {
    let url = "/statis/system";
    if (uuid) {
      const params = new URLSearchParams();
      params.append("uuid", uuid);
      url += `?${params.toString()}`;
    }
    return this.request<ServerInfo>(url);
  }
}

export const api = new ApiClient();