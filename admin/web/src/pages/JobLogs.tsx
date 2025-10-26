import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { JobLog, Node } from "@/types/api"; // 导入 Node 类型
import { Activity, Search, CheckCircle2, XCircle, RotateCcw } from "lucide-react";

const JobLogs = () => {
  const [logs, setLogs] = useState<JobLog[]>([]);
  const [isLoading, setIsLoading] = useState(false); // 控制表格内加载
  const [isInitialLoading, setIsInitialLoading] = useState(true); // 控制首次全页加载
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const navigate = useNavigate();

  // --- 新增 State ---
  const [allNodes, setAllNodes] = useState<Node[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterStatus, setFilterStatus] = useState("all"); // "all", "true", "false"
  const [filterNodeUuid, setFilterNodeUuid] = useState("all");

  // 仅在组件首次挂载时执行，用于获取初始数据和节点列表
  useEffect(() => {
    const initialLoad = async () => {
      setIsInitialLoading(true);
      try {
        // 并行获取初始日志和所有节点
        const [logData, nodesData] = await Promise.all([
          api.getJobLogs({ page: 1, pageSize: 20 }),
          api.searchNodes({ page: 1, pageSize: 1000 }) // 获取所有节点用于筛选
        ]);
        setLogs(logData.list || []);
        setTotal(logData.total);
        setAllNodes(nodesData.list || []);
      } catch (error) {
        if (error instanceof Error) toast.error(`Initial load failed: ${error.message}`);
      } finally {
        setIsInitialLoading(false);
      }
    };
    initialLoad();
  }, []);

  // 当页码变化时，重新获取数据
  useEffect(() => {
    // 避免在首次加载时重复获取
    if (!isInitialLoading) {
      fetchLogs();
    }
  }, [page]);
  
  const fetchLogs = async (resetPage = false) => {
    // 如果是新搜索，则将页码重置为1
    const currentPage = resetPage ? 1 : page;
    if (resetPage) setPage(1);

    setIsLoading(true);
    try {
      const success = filterStatus !== 'all' ? filterStatus === 'true' : undefined;
      const nodeUuid = filterNodeUuid !== 'all' ? filterNodeUuid : undefined;
      
      const logData = await api.getJobLogs({
        page: currentPage,
        pageSize: 20,
        name: searchQuery || undefined,
        success,
        nodeUuid: nodeUuid, // 适配 API 可能的类型差异
      });
      
      setLogs(logData.list || []);
      setTotal(logData.total);
    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
      else toast.error("Failed to load logs due to an unknown error");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSearch = () => fetchLogs(true);
  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') handleSearch();
  };
  const handleReset = () => {
    setSearchQuery("");
    setFilterStatus("all");
    setFilterNodeUuid("all");
    // 确保状态更新后再触发搜索
    setTimeout(() => fetchLogs(true), 0);
  };

  // --- UI 辅助函数 ---
  const formatDate = (ts: number) => ts ? new Date(ts * 1000).toLocaleString() : 'N/A';
  const formatDuration = (start: number, end: number) => {
    if (!start || !end || end < start) return 'N/A';
    const durationSec = end - start;
    if (durationSec < 1) return "< 1s";
    const minutes = Math.floor(durationSec / 60);
    const seconds = Math.floor(durationSec % 60);
    return minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`;
  };

  if (isInitialLoading) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold mb-2">Job Logs</h1>
        <p className="text-muted-foreground">View execution history and outputs</p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Execution Logs</CardTitle>
          <CardDescription>History of all job executions across all nodes.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div className="relative col-span-1 sm:col-span-2">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input placeholder="Search by job name..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} onKeyDown={handleKeyDown} className="pl-9" />
            </div>
            <Select value={filterStatus} onValueChange={setFilterStatus}>
              <SelectTrigger><SelectValue placeholder="Filter by Status" /></SelectTrigger>
              <SelectContent><SelectItem value="all">All Statuses</SelectItem><SelectItem value="true">Success</SelectItem><SelectItem value="false">Failed</SelectItem></SelectContent>
            </Select>
            <Select value={filterNodeUuid} onValueChange={setFilterNodeUuid}>
              <SelectTrigger><SelectValue placeholder="Filter by Node" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Nodes</SelectItem>
                {allNodes.map(node => <SelectItem key={node.uuid} value={node.uuid}>{node.hostname}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-end gap-2 mb-4">
            <Button variant="outline" onClick={handleReset} className="gap-2"><RotateCcw className="h-4 w-4" /> Reset</Button>
            <Button onClick={handleSearch} className="gap-2"><Search className="h-4 w-4" /> Search</Button>
          </div>

          <div className="border rounded-lg">
            <Table>
              <TableHeader><TableRow><TableHead>Status</TableHead><TableHead>Job Name</TableHead><TableHead>Node</TableHead><TableHead>Start Time</TableHead><TableHead>Duration</TableHead><TableHead>Output</TableHead></TableRow></TableHeader>
              <TableBody>
                {isLoading ? (
                  <TableRow><TableCell colSpan={6} className="text-center py-8"><Activity className="h-6 w-6 animate-spin text-primary mx-auto" /></TableCell></TableRow>
                ) : logs.length === 0 ? (
                  <TableRow><TableCell colSpan={6} className="text-center py-8 text-muted-foreground">No logs found for the current criteria.</TableCell></TableRow>
                ) : (
                  logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell>{log.success ? (<Badge className="bg-success/10 text-success border-success/20"><CheckCircle2 className="h-3 w-3 mr-1" />Success</Badge>) : (<Badge className="bg-destructive/10 text-destructive border-destructive/20"><XCircle className="h-3 w-3 mr-1" />Failed</Badge>)}</TableCell>
                      <TableCell className="font-medium">
                          <a onClick={() => navigate(`/jobs/${log.job_id}`)} className="hover:underline cursor-pointer text-primary">{log.name}</a>
                      </TableCell>
                      <TableCell>
                        <div>
                          <a onClick={() => navigate(`/nodes/${log.node_uuid}`)} className="text-sm font-semibold hover:underline cursor-pointer">{log.hostname}</a>
                          <div className="text-xs text-muted-foreground">{log.ip}</div>
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">{formatDate(log.start_time)}</TableCell>
                      <TableCell className="text-sm font-medium">{formatDuration(log.start_time, log.end_time)}</TableCell>
                      <TableCell>
                        <Dialog>
                            <DialogTrigger asChild><Button variant="outline" size="sm">View Output</Button></DialogTrigger>
                            <DialogContent className="max-w-3xl"><DialogHeader><DialogTitle>Log Output for: {log.name}</DialogTitle></DialogHeader><pre className="mt-2 p-4 bg-muted rounded-md text-xs overflow-auto max-h-[60vh]"><code>{log.output || "No output captured."}</code></pre></DialogContent>
                        </Dialog>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {total > 20 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-muted-foreground">Showing {Math.min((page - 1) * 20 + 1, total)} to {Math.min(page * 20, total)} of {total} logs</p>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" onClick={() => setPage(page - 1)} disabled={page === 1}>Previous</Button>
                <Button variant="outline" size="sm" onClick={() => setPage(page + 1)} disabled={page * 20 >= total}>Next</Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default JobLogs;