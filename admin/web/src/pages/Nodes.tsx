import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { Node, NodeStatus } from "@/types/api";
import { Activity, Server, Trash2, Search, RotateCcw } from "lucide-react";

const Nodes = () => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  // --- 新增筛选状态 ---
  const [searchIp, setSearchIp] = useState("");
  const [filterStatus, setFilterStatus] = useState("all"); // "all", "1", "2"

  useEffect(() => {
    fetchNodes();
  }, []);

  const fetchNodes = async () => {
    setIsLoading(true);
    try {
      const nodesData = await api.searchNodes({ 
        page: 1, 
        pageSize: 100, // 获取足够多的节点
        ip: searchIp || undefined,
        status: filterStatus !== "all" ? Number(filterStatus) as NodeStatus : undefined,
      });
      setNodes(nodesData.list || []);
    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
      else toast.error("Failed to load nodes due to an unknown error");
    } finally {
      setIsLoading(false);
    }
  };
  
  // --- 新增处理函数 ---
  const handleDelete = async (nodeId: number, nodeHostname: string) => {
    if (window.confirm(`Are you sure you want to remove the node "${nodeHostname}"?`)) {
      try {
        await api.deleteNode({ ids: [nodeId] });
        toast.success(`Node "${nodeHostname}" has been removed.`);
        fetchNodes(); // 成功后刷新列表
      } catch (error) {
        if (error instanceof Error) toast.error(error.message);
        else toast.error("Failed to remove node.");
      }
    }
  };
  
  const handleReset = () => {
    setSearchIp("");
    setFilterStatus("all");
    // 使用 setTimeout 确保状态更新后再触发搜索
    setTimeout(() => fetchNodes(), 0);
  };
  
  // --- UI 辅助函数 ---
  const getStatusBadge = (status: NodeStatus) => {
    return status === NodeStatus.NodeConnSuccess ? (
      <Badge className="bg-success/10 text-success border-success/20 hover:bg-success/20">Online</Badge>
    ) : (
      <Badge className="bg-destructive/10 text-destructive border-destructive/20 hover:bg-destructive/20">Offline</Badge>
    );
  };

  const formatDate = (timestamp: number) => {
    if (!timestamp) return 'N/A';
    return new Date(timestamp * 1000).toLocaleString();
  };

  if (isLoading && nodes.length === 0) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold mb-2">Nodes</h1>
        <p className="text-muted-foreground">Monitor and manage your worker nodes</p>
      </div>

      {/* --- 新增筛选区域 --- */}
      <Card>
        <CardContent className="pt-6 flex flex-col sm:flex-row gap-4">
            <div className="relative flex-grow">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input placeholder="Search by IP address..." value={searchIp} onChange={(e) => setSearchIp(e.target.value)} className="pl-9" />
            </div>
            <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="sm:w-[180px]"><SelectValue placeholder="Filter by Status" /></SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">All Statuses</SelectItem>
                    <SelectItem value={String(NodeStatus.NodeConnSuccess)}>Online</SelectItem>
                    <SelectItem value={String(NodeStatus.NodeConnFail)}>Offline</SelectItem>
                </SelectContent>
            </Select>
            <div className="flex gap-2">
                <Button variant="outline" onClick={handleReset} className="w-full sm:w-auto"><RotateCcw className="h-4 w-4 mr-2"/>Reset</Button>
                <Button onClick={fetchNodes} className="w-full sm:w-auto"><Search className="h-4 w-4 mr-2"/>Search</Button>
            </div>
        </CardContent>
      </Card>
      
      {isLoading && <div className="text-center"><Activity className="h-6 w-6 animate-spin text-primary mx-auto" /></div>}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {!isLoading && nodes.length === 0 ? (
          <Card className="col-span-full"><CardContent className="py-12 text-center"><Server className="h-12 w-12 mx-auto mb-4 text-muted-foreground" /><p className="text-muted-foreground">No nodes match your criteria.</p></CardContent></Card>
        ) : (
          nodes.map((node) => (
            // [更新] 卡片变为可点击，导航到详情页
            <Card 
                key={node.id} 
                className="relative overflow-hidden hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => navigate(`/nodes/${node.uuid}`, { state: { node } })} // 传递 node 对象
            >
              <div className={`absolute top-0 left-0 right-0 h-1 ${node.status === NodeStatus.NodeConnSuccess ? "bg-success" : "bg-destructive"}`} />
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2"><Server className="h-5 w-5 text-primary" /><CardTitle className="text-lg">{node.hostname}</CardTitle></div>
                  {getStatusBadge(node.status)}
                </div>
                <CardDescription className="font-mono text-xs pt-1">{node.ip}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between"><span className="text-muted-foreground">UUID:</span><span className="font-mono text-xs">{node.uuid.slice(0, 8)}...</span></div>
                  <div className="flex justify-between"><span className="text-muted-foreground">Version:</span><span>{node.version}</span></div>
                  <div className="flex justify-between"><span className="text-muted-foreground">Jobs:</span><span className="font-semibold">{node.jobCount}</span></div>
                  <div className="flex justify-between"><span className="text-muted-foreground">PID:</span><span className="font-mono text-xs">{node.pid}</span></div>
                  {node.up > 0 && (<div className="flex justify-between"><span className="text-muted-foreground">Last Seen:</span><span className="text-xs">{formatDate(node.up)}</span></div>)}
                </div>
                <div className="pt-2 border-t border-border">
                  {/* [更新] 删除按钮功能激活 */}
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    className="w-full text-destructive hover:text-destructive"
                    onClick={(e) => {
                        e.stopPropagation(); // 阻止点击事件冒泡到父级 Card
                        handleDelete(node.id, node.hostname);
                    }}
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Remove Node
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
};

export default Nodes;