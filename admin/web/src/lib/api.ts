import type {
  ApiResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  User,
  Job,
  JobLog,
  JobSearchRequest,
  Node,
  NodeSearchRequest,
  Script,
  ScriptSearchRequest,
  SystemStatistics,
  WeekStatistics,
  PageResult,
} from "@/types/api";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8089";

class ApiClient {
  private getToken(): string | null {
    return localStorage.getItem("token");
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const token = this.getToken();
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    if (token && !endpoint.includes("/login") && !endpoint.includes("/register")) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    return response.json();
  }

  // Auth APIs
  async login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    return this.request<LoginResponse>("/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async register(data: RegisterRequest): Promise<ApiResponse<User>> {
    return this.request<User>("/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // User APIs
  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<User>("/user/find");
  }

  async searchUsers(data: any): Promise<ApiResponse<PageResult<User>>> {
    return this.request<PageResult<User>>("/user/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updatePassword(password: string, new_password: string): Promise<ApiResponse<void>> {
    return this.request<void>("/user/pw", {
      method: "POST",
      body: JSON.stringify({ password, new_password }),
    });
  }

  // Job APIs
  async searchJobs(data: JobSearchRequest): Promise<ApiResponse<PageResult<Job>>> {
    return this.request<PageResult<Job>>("/job/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async createJob(data: Partial<Job>): Promise<ApiResponse<Job>> {
    return this.request<Job>("/job/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateJob(data: Partial<Job>): Promise<ApiResponse<Job>> {
    return this.request<Job>("/job/edit", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteJobs(ids: number[]): Promise<ApiResponse<void>> {
    return this.request<void>("/job/del", {
      method: "POST",
      body: JSON.stringify({ ids }),
    });
  }

  async getJobLogs(data: any): Promise<ApiResponse<PageResult<JobLog>>> {
    return this.request<PageResult<JobLog>>("/job/log", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async executeJobOnce(job_id: number, node_uuid: string): Promise<ApiResponse<void>> {
    return this.request<void>("/job/once", {
      method: "POST",
      body: JSON.stringify({ job_id, node_uuid }),
    });
  }

  // Node APIs
  async searchNodes(data: NodeSearchRequest): Promise<ApiResponse<PageResult<Node>>> {
    return this.request<PageResult<Node>>("/node/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteNode(uuid: string): Promise<ApiResponse<void>> {
    return this.request<void>("/node/del", {
      method: "POST",
      body: JSON.stringify({ uuid }),
    });
  }

  // Script APIs
  async searchScripts(data: ScriptSearchRequest): Promise<ApiResponse<PageResult<Script>>> {
    return this.request<PageResult<Script>>("/script/search", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async createScript(data: Script): Promise<ApiResponse<Script>> {
    return this.request<Script>("/script/add", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateScript(data: Partial<Script>): Promise<ApiResponse<Script>> {
    return this.request<Script>("/script/edit", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async deleteScripts(ids: number[]): Promise<ApiResponse<void>> {
    return this.request<void>("/script/del", {
      method: "POST",
      body: JSON.stringify({ ids }),
    });
  }

  // Statistics APIs
  async getTodayStatistics(): Promise<ApiResponse<SystemStatistics>> {
    return this.request<SystemStatistics>("/statis/today");
  }

  async getWeekStatistics(): Promise<ApiResponse<WeekStatistics>> {
    return this.request<WeekStatistics>("/statis/week");
  }

  async getSystemInfo(): Promise<ApiResponse<any>> {
    return this.request<any>("/statis/system");
  }
}

export const api = new ApiClient();
