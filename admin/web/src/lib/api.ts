import type {
  ApiResponse,
  PageResult,
  User,
  Job,
  JobLog,
  Node,
  Script,
  TodayStatistics,
  WeekStatistics,
  SystemInfo,
  LoginResponse as ApiLoginResponse,
  Login,
  Register,
  ChangePassword,
  UserSearch,
  UserUpdate,
  Delete,
  JobSearch,
  JobUpdate,
  JobLogSearch,
  JobOnce,
  JobKill,
  NodeSearch,
  ScriptSearch,
  ScriptUpdate,

  GetSystem,
  UserFind,
  JobFind,
  ScriptFind,
} from "@/types/api"

// const API_BASE_URL = "/api";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api"

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
   * @returns {Promise<T>} 返回一个 Promise，解析为 API 响应中的 data 字段
   * @throws {Error} 如果网络请求失败、响应状态码不是 2xx 或业务码不为 200
   */
  // ===================================================================================
  // [更正 1] 修改 request 方法的返回类型签名，去掉 ApiResponse 包装
  // ===================================================================================
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const token = this.getToken();
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...options.headers,
    };

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
      console.error(`API 请求失败，状态码: ${response.status}`, { endpoint, errorText });
      throw new Error(`网络请求错误 (状态码: ${response.status})`);
    }

    // 克隆响应以防解析失败后需要再次读取
    const responseClone = response.clone();
    try {
      const apiResponse: ApiResponse<T> = await response.json();

      // 在这里集中处理业务状态码
      if (apiResponse.code === 200) {
        // 成功，直接返回核心数据 data，这匹配了新的 Promise<T> 返回类型
        return apiResponse.data;
      } else {
        // 业务逻辑失败，抛出后端提供的错误信息
        throw new Error(apiResponse.msg || "API 返回未知错误");
      }
    } catch (error) {
      // 捕获上面抛出的错误或 JSON 解析错误
      console.error("API 响应处理失败:", error);
      if (error instanceof SyntaxError) {
        // 如果是 JSON 解析错误，尝试打印原始文本
        const rawText = await responseClone.text();
        console.error("原始响应文本:", rawText);
        throw new Error("无法解析服务器响应");
      }
      throw error; // 重新抛出错误，让调用方可以捕获
    }
  }

  // ===================================================================================
  // [更正 2] 更新所有公共 API 方法的返回类型，使其与 request 的新返回类型一致
  // ===================================================================================

  // Auth APIs
  async login(data: Login): Promise<ApiLoginResponse["data"]> {
    return this.request<ApiLoginResponse["data"]>("/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async register(data: Register): Promise<User> {
    return this.request<User>("/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // User APIs
  async getCurrentUser(): Promise<User> {
    return this.request<User>("/user/find", {
      method: "POST",
      body: JSON.stringify({}),
    });
  }

  async findUserById(data: UserFind): Promise<User> {
    return this.request<User>("/user/find", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async searchUsers(data: UserSearch): Promise<PageResult<User>> {
    return this.request<PageResult<User>>("/user/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateUser(data: UserUpdate): Promise<object> {
    return this.request<object>("/user/update", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updatePassword(data: ChangePassword): Promise<object> {
    return this.request<object>("/user/change_pw", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteUsers(data: Delete): Promise<object> {
    return this.request<object>("/user/del", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // Job APIs
  async searchJobs(data: JobSearch): Promise<PageResult<Job>> {
    return this.request<PageResult<Job>>("/job/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async findJobById(data: JobFind): Promise<Job> {
    return this.request<Job>(`/job/find?`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // 传入id则更新指定任务信息
  // 无id则创建新任务
  async saveJob(data: JobUpdate): Promise<Job> {
    return this.request<Job>("/job/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteJobs(data: Delete): Promise<object> {
    return this.request<object>("/job/del", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getJobLogs(data: JobLogSearch): Promise<PageResult<JobLog>> {
    return this.request<PageResult<JobLog>>("/job/log", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async executeJobOnce(data: JobOnce): Promise<object> {
    return this.request<object>("/job/once", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async killJob(data: JobKill): Promise<object> {
    return this.request<object>("/job/kill", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // Node APIs
  async searchNodes(data: NodeSearch): Promise<PageResult<Node>> {
    return this.request<PageResult<Node>>("/node/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteNode(data: Delete): Promise<object> {
    return this.request<object>("/node/del", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // Script APIs
  async searchScripts(data: ScriptSearch): Promise<PageResult<Script>> {
    return this.request<PageResult<Script>>("/script/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async findScriptById(data: ScriptFind): Promise<Script> {
    return this.request<Script>(`/script/find?`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // 传入id则更新指定脚本
  // 无id则创建新脚本
  async saveScript(data: ScriptUpdate): Promise<Script> {
    return this.request<Script>("/script/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteScripts(data: Delete): Promise<object> {
    return this.request<object>("/script/del", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // Statistics APIs
  async getTodayStatistics(): Promise<TodayStatistics> {
    return this.request<TodayStatistics>("/statis/today");
  }

  async getWeekStatistics(): Promise<WeekStatistics> {
    return this.request<WeekStatistics>("/statis/week");
  }

  // 传入uuid返回指定节点的系统信息
  // 无uuid传入则返回当前管理服务器的系统信息
  async getSystemInfo(data: GetSystem): Promise<SystemInfo> {
    return this.request<SystemInfo>("/statis/system", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }
}

export const api = new ApiClient();