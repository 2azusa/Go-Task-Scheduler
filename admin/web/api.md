好的，这是根据您提供的 TypeScript 接口和 API 客户端代码整理的 API 文档。这份文档旨在让大型语言模型（LLM）能够理解如何调用 `api` 对象中的各个方法来与后端服务进行交互。

---

## **API 客户端文档**

### **概述**

本文档描述了一个用于与后端服务交互的 `ApiClient` 实例，该实例已导出为 `api`。所有方法都返回一个 `Promise`，该 `Promise` 会直接解析为响应体中的 `data` 字段。客户端会自动处理认证 Token 的附加和业务错误码的抛出。

- **成功**: `Promise` 解析为期望的数据结构。
- **失败**: `Promise` 会被拒绝（reject），并附带一个 `Error` 对象，其 `message` 属性包含后端返回的错误信息。

### **通用数据结构**

- **`Delete`**: 用于批量删除操作的负载。
  - `ids`: `number[]` - 一个包含待删除项数字 ID 的数组。
- **`PageInfo`**: 用于分页查询的基础接口。
  - `page?`: `number` - 要检索的页码。
  - `pageSize?`: `number` - 每页要检索的项目数。
- **`PageResult<T>`**: 分页查询结果的通用结构。
  - `list`: `T[]` - 当前页的项目列表。
  - `total`: `number` - 符合条件的项目总数。
  - `page`: `number` - 当前页码。
  - `pageSize`: `number` - 每页的项目数。

---

### **1. 认证 (Auth)**

#### **1.1 用户登录**

- **方法**: `api.login(data)`
- **描述**: 使用用户名和密码对用户进行身份验证。
- **参数**:
  - `data` (`Login`):
    - `username`: `string` - 用户名（长度 2-20）。
    - `password`: `string` - 密码（字母数字，长度 4-20）。
- **返回**: `Promise<{ token: string; user: User; }>`
  - `token`: 用于后续请求的认证令牌。
  - `user`: 登录用户的详细信息 (`User` 对象)。

#### **1.2 用户注册**

- **方法**: `api.register(data)`
- **描述**: 创建一个新用户账户。
- **参数**:
  - `data` (`Register`):
    - `username`: `string` - 用户名（长度 2-20）。
    - `password`: `string` - 密码（字母数字，长度 4-20）。
    - `email`: `string` - 用户的电子邮件地址。
    - `role?`: `Role` - 用户的角色 (`1` for Normal, `2` for Admin)。默认为 `1`。
- **返回**: `Promise<User>` - 包含新创建用户信息的 `User` 对象。

---

### **2. 用户管理 (User)**

#### **2.1 获取当前用户信息**

- **方法**: `api.getCurrentUser()`
- **描述**: 根据当前请求的认证 Token 获取用户信息。
- **参数**: 无。
- **返回**: `Promise<User>` - 当前登录用户的 `User` 对象。

#### **2.2 根据 ID 查找用户**

- **方法**: `api.findUserById(data)`
- **描述**: 根据指定的唯一 ID 查找用户信息。
- **参数**:
  - `data` (`UserFind`):
    - `id?`: `number` - 要查找的用户的 ID。
- **返回**: `Promise<User>` - 查找到的用户的 `User` 对象。

#### **2.3 搜索用户**

- **方法**: `api.searchUsers(data)`
- **描述**: 根据多个可选条件进行分页搜索。
- **参数**:
  - `data` (`UserSearch`): 继承自 `PageInfo`。
    - `id?`: `number` - 按用户 ID 过滤。
    - `username?`: `string` - 按用户名过滤（支持模糊匹配）。
    - `email?`: `string` - 按电子邮件地址过滤。
    - `role?`: `Role` - 按用户角色过滤 (`1` or `2`)。
- **返回**: `Promise<PageResult<User>>` - 包含用户列表和分页信息的结果。

#### **2.4 更新用户信息**

- **方法**: `api.updateUser(data)`
- **描述**: 更新指定用户的信息。只有提供的字段才会被更新。
- **参数**:
  - `data` (`UserUpdate`):
    - `username?`: `string` - 新的用户名。
    - `email?`: `string` - 新的电子邮件地址。
    - `role?`: `Role` - 用户的新角色 (`1` or `2`)。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

#### **2.5 修改密码**

- **方法**: `api.updatePassword(data)`
- **描述**: 修改当前登录用户的密码。
- **参数**:
  - `data` (`ChangePassword`):
    - `password`: `string` - 当前密码。
    - `newPassword`: `string` - 新密码。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

#### **2.6 删除用户**

- **方法**: `api.deleteUsers(data)`
- **描述**: 根据 ID 批量删除用户。
- **参数**:
  - `data` (`Delete`):
    - `ids`: `number[]` - 待删除用户的 ID 数组。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

---

### **3. 任务管理 (Job)**

#### **3.1 搜索任务**

- **方法**: `api.searchJobs(data)`
- **描述**: 根据多个可选条件分页搜索计划任务。
- **参数**:
  - `data` (`JobSearch`): 继承自 `PageInfo`。
    - `id?`: `number` - 按任务 ID 过滤。
    - `name?`: `string` - 按任务名称过滤。
    - `runOn?`: `string` - 按任务运行的节点名称过滤。
    - `type?`: `JobType` - 按任务类型过滤 (`1` for Cmd, `2` for Http)。
    - `status?`: `JobStatus` - 按任务分配状态过滤 (`0` for Not Assigned, `1` for Assigned)。
- **返回**: `Promise<PageResult<Job>>` - 包含任务列表和分页信息的结果。

#### **3.2 根据 ID 查找任务**

- **方法**: `api.findJobById(data)`
- **描述**: 获取指定 ID 的任务详情。
- **参数**:
  - `data` (`JobFind`):
    - `id`: `number` - 要查找的任务的 ID。
- **返回**: `Promise<Job>` - 查找到的任务 `Job` 对象。

#### **3.3 创建或更新任务**

- **方法**: `api.saveJob(data)`
- **描述**: 如果 `data` 中包含 `id`，则更新现有任务；否则，创建新任务。
- **参数**:
  - `data` (`JobUpdate`):
    - `id?`: `number` - 要更新的任务 ID。
    - `name`: `string` - 任务名称。
    - `jobType`: `JobType` - 任务类型 (`1`: Cmd, `2`: Http)。
    - `command`: `string` - 要执行的命令或 HTTP URL。
    - `spec`: `string` - cron 表达式。
    - `allocation`: `Allocation` - 分配方式 (`1`: Manual, `2`: Auto)。
    - `runOn?`: `string` - 手动分配时必须指定的节点 UUID。
    - `timeout?`: `number` - 执行超时（秒），0 为无超时。
    - `retryTimes?`: `number` - 失败重试次数。
    - `retryInterval?`: `number` - 重试间隔（秒）。
    - `httpMethod?`: `HttpMethod` - HTTP 任务方法 (`1`: GET, `2`: POST)。
    - `sciptId?`: `number[]` - 关联的脚本 ID 数组。
    - `notifyType`: `NotifyType` - 通知类型 (`1`: Mail, `2`: WebHook)。
    - `notifyTo`: `number[]` - 需要通知的用户 ID 数组。
    - `note?`: `string` - 备注。
- **返回**: `Promise<Job>` - 创建或更新后的任务 `Job` 对象。

#### **3.4 删除任务**

- **方法**: `api.deleteJobs(data)`
- **描述**: 根据 ID 批量删除任务。
- **参数**:
  - `data` (`Delete`):
    - `ids`: `number[]` - 待删除任务的 ID 数组。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

#### **3.5 获取任务日志**

- **方法**: `api.getJobLogs(data)`
- **描述**: 分页搜索任务的执行日志。
- **参数**:
  - `data` (`JobLogSearch`): 继承自 `PageInfo`。
    - `name?`: `string` - 按任务名称过滤。
    - `jobId?`: `number` - 按任务 ID 过滤。
    - `nodeUuid?`: `number` - 按节点 UUID 过滤。
    - `success?`: `boolean` - 按执行状态过滤 (`true` 成功, `false` 失败)。
- **返回**: `Promise<PageResult<JobLog>>` - 包含任务日志列表和分页信息的结果。

#### **3.6 立即执行一次任务**

- **方法**: `api.executeJobOnce(data)`
- **描述**: 在指定节点上触发一次性任务执行。
- **参数**:
  - `data` (`JobOnce`):
    - `jobId`: `number` - 要执行的任务 ID。
    - `nodeUuid`: `string` - 目标节点的 UUID。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

#### **3.7 终止正在运行的任务**

- **方法**: `api.killJob(data)`
- **描述**: 终止在特定节点上正在运行的任务实例。
- **参数**:
  - `data` (`JobKill`):
    - `jobId`: `number` - 要终止的任务 ID。
    - `nodeUuid`: `string` - 任务所在节点的 UUID。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

---

### **4. 节点管理 (Node)**

#### **4.1 搜索节点**

- **方法**: `api.searchNodes(data)`
- **描述**: 根据多个可选条件分页搜索工作节点。
- **参数**:
  - `data` (`NodeSearch`): 继承自 `PageInfo`。
    - `ip?`: `string` - 按节点 IP 地址过滤。
    - `uuid?`: `string` - 按节点 UUID 过滤。
    - `up?`: `number` - 按节点启动时间戳过滤。
    - `status?`: `NodeStatus` - 按节点连接状态过滤 (`1`: Online, `2`: Offline)。
- **返回**: `Promise<PageResult<Node>>` - 包含节点列表和分页信息的结果。

#### **4.2 删除节点**

- **方法**: `api.deleteNode(data)`
- **描述**: 根据 ID 批量删除节点。**注意：参数类型是 `Delete`，其 `ids` 字段应对应节点的数字 ID，而非 UUID**。
- **参数**:
  - `data` (`Delete`):
    - `ids`: `number[]` - 待删除节点的 ID 数组。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

---

### **5. 脚本管理 (Script)**

#### **5.1 搜索脚本**

- **方法**: `api.searchScripts(data)`
- **描述**: 根据可选条件分页搜索可重用脚本。
- **参数**:
  - `data` (`ScriptSearch`): 继承自 `PageInfo`。
    - `id?`: `number` - 按脚本 ID 过滤。
    - `name?`: `string` - 按脚本名称过滤。
- **返回**: `Promise<PageResult<Script>>` - 包含脚本列表和分页信息的结果。

#### **5.2 根据 ID 查找脚本**

- **方法**: `api.findScriptById(data)`
- **描述**: 获取指定 ID 的脚本详情。
- **参数**:
  - `data` (`ScriptFind`):
    - `id`: `number` - 要查找的脚本的 ID。
- **返回**: `Promise<Script>` - 查找到的脚本 `Script` 对象。

#### **5.3 创建或更新脚本**

- **方法**: `api.saveScript(data)`
- **描述**: 如果 `data` 中包含 `id`，则更新现有脚本；否则，创建新脚本。
- **参数**:
  - `data` (`ScriptUpdate`):
    - `id?`: `number` - 要更新的脚本 ID。
    - `name?`: `string` - 脚本名称。
    - `command?`: `string` - 脚本内容。
- **返回**: `Promise<Script>` - 创建或更新后的脚本 `Script` 对象。

#### **5.4 删除脚本**

- **方法**: `api.deleteScripts(data)`
- **描述**: 根据 ID 批量删除脚本。
- **参数**:
  - `data` (`Delete`):
    - `ids`: `number[]` - 待删除脚本的 ID 数组。
- **返回**: `Promise<object>` - 一个表示操作成功的空对象。

---

### **6. 统计 (Statistics)**

#### **6.1 获取今日统计**

- **方法**: `api.getTodayStatistics()`
- **描述**: 获取今日任务执行和节点状态的摘要统计。
- **参数**: 无。
- **返回**: `Promise<TodayStatistics>` - 包含今日统计数据的对象。

#### **6.2 获取周统计**

- **方法**: `api.getWeekStatistics()`
- **描述**: 获取过去一周每日任务执行成功与失败的统计数据。
- **参数**: 无。
- **返回**: `Promise<WeekStatistics>` - 包含周统计数据的对象。

#### **6.3 获取系统信息**

- **方法**: `api.getSystemInfo(data)`
- **描述**: 获取系统信息。如果提供了 `uuid`，则返回指定节点的信息；否则，返回管理服务器的信息。
- **参数**:
  - `data` (`GetSystem`):
    - `uuid?`: `number` - 目标节点的 UUID。
- **返回**: `Promise<SystemInfo>` - 包含 CPU、内存、磁盘和操作系统信息的对象。