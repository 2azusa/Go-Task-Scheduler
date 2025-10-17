import { useEffect, useState } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { api } from "@/lib/api";
import { Node } from "@/types/api";
import { toast } from "sonner";

const Nodes = () => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadNodes();
  }, []);

  const loadNodes = async () => {
    setLoading(true);
    try {
      const response = await api.searchNodes({ page: 1, page_size: 100 });
      if (response.code === 200) {
        setNodes(response.data.list || []);
      }
    } catch (error) {
      toast.error("加载节点列表失败");
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: number) => {
    if (status === 1) {
      return <Badge className="bg-success text-success-foreground">在线</Badge>;
    }
    return <Badge variant="destructive">离线</Badge>;
  };

  const formatUptime = (timestamp: number) => {
    const hours = Math.floor((Date.now() / 1000 - timestamp) / 3600);
    const days = Math.floor(hours / 24);
    if (days > 0) return `${days}天`;
    return `${hours}小时`;
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">节点管理</h1>
          <p className="text-muted-foreground">监控所有工作节点状态</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>节点列表</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>状态</TableHead>
                  <TableHead>主机名</TableHead>
                  <TableHead>IP地址</TableHead>
                  <TableHead>UUID</TableHead>
                  <TableHead>版本</TableHead>
                  <TableHead>运行时间</TableHead>
                  <TableHead className="text-right">任务数</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center">
                      加载中...
                    </TableCell>
                  </TableRow>
                ) : nodes.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground">
                      暂无节点
                    </TableCell>
                  </TableRow>
                ) : (
                  nodes.map((node) => (
                    <TableRow key={node.id}>
                      <TableCell>{getStatusBadge(node.status)}</TableCell>
                      <TableCell className="font-medium">{node.hostname}</TableCell>
                      <TableCell>{node.ip}</TableCell>
                      <TableCell className="font-mono text-sm">{node.uuid}</TableCell>
                      <TableCell>{node.version}</TableCell>
                      <TableCell>{formatUptime(node.up)}</TableCell>
                      <TableCell className="text-right">{node.job_count}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
};

export default Nodes;
