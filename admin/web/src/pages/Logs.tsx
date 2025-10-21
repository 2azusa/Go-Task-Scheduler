import { useEffect, useState, useCallback } from "react";
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
import { api } from "@/lib/api";
import { JobLog } from "@/types/api";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

const Logs = () => {
  const [logs, setLogs] = useState<JobLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const loadLogs = useCallback(async () => {
    setLoading(true);
    try {
      const response = await api.getJobLogs({ page, pageSize: 10 });
      if (response.code === 200) {
        setLogs(response.data.list || []);
        setTotal(response.data.total);
      }
    } catch (error) {
      toast.error("加载日志列表失败");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  const getStatusBadge = (success: boolean) => {
    if (success) return <Badge variant="default">成功</Badge>;
    return <Badge variant="destructive">失败</Badge>;
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">任务日志</h1>
          <p className="text-muted-foreground">查看所有任务的执行日志</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>日志列表</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>任务ID</TableHead>
                  <TableHead>任务名称</TableHead>
                  <TableHead>输出</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>执行时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center">
                      加载中...
                    </TableCell>
                  </TableRow>
                ) : logs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-muted-foreground">
                      暂无日志
                    </TableCell>
                  </TableRow>
                ) : (
                  logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell>{log.id}</TableCell>
                      <TableCell>{log.jobId}</TableCell>
                      <TableCell>{log.name}</TableCell>
                      <TableCell className="max-w-md truncate font-mono text-sm">
                        {log.output}
                      </TableCell>
                      <TableCell>{getStatusBadge(log.success)}</TableCell>
                      <TableCell>
                        {new Date(log.startTime * 1000).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>

            {total > 10 && (
              <div className="mt-4 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  共 {total} 条记录
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(page - 1)}
                    disabled={page === 1}
                  >
                    上一页
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(page + 1)}
                    disabled={page * 10 >= total}
                  >
                    下一页
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
};

export default Logs;
