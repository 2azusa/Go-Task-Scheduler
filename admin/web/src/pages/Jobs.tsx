import { useEffect, useState, useCallback } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { Button } from "@/components/ui/button";
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
import { Job } from "@/types/api";
import { Plus, Play, Trash2, Pencil } from "lucide-react";
import { toast } from "sonner";
import { JobEdit } from "@/components/business/JobEdit";

const Jobs = () => {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [editingJob, setEditingJob] = useState<Partial<Job> | null>(null);

  // useEffect(() => {
  //   loadJobs();
  // }, [page]);

  // const loadJobs = async () => {
  //   setLoading(true);
  //   try {
  //     const response = await api.searchJobs({ page, pageSize: 10 });
  //     if (response.code === 200) {
  //       setJobs(response.data.list || []);
  //       setTotal(response.data.total);
  //     }
  //   } catch (error) {
  //     toast.error("加载任务列表失败");
  //   } finally {
  //     setLoading(false);
  //   }
  // };

  const loadJobs = useCallback(async () => {
    setLoading(true);
    try {
      const response = await api.searchJobs({ page, pageSize: 10 });
      if (response.code === 200) {
        setJobs(response.data.list || []);
        setTotal(response.data.total);
      }
    } catch (error) {
      toast.error("加载任务列表失败");
    } finally {
      setLoading(false);
    }
  }, [page]);
  useEffect(() => {
    loadJobs();
  }, [loadJobs]);

  const handleDelete = async (id: number) => {
    try {
      const response = await api.deleteJobs([id]);
      if (response.code === 200) {
        toast.success("删除成功");
        loadJobs();
      } else {
        toast.error(response.msg || "删除失败");
      }
    } catch (error) {
      toast.error("删除失败");
    }
  };

  const getJobTypeLabel = (type: number) => {
    return type === 1 ? "命令任务" : "HTTP任务";
  };

  const getStatusBadge = (status?: number) => {
    if (!status) return <Badge variant="secondary">未知</Badge>;
    if (status === 1) return <Badge variant="default">运行中</Badge>;
    return <Badge variant="secondary">已停止</Badge>;
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-foreground">任务管理</h1>
            <p className="text-muted-foreground">管理和监控所有定时任务</p>
          </div>
          <Button onClick={() => setEditingJob({})}>
            <Plus className="mr-2 h-4 w-4" />
            创建任务
          </Button>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>任务列表</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>任务名称</TableHead>
                  <TableHead>类型</TableHead>
                  <TableHead>Cron表达式</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>节点</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center">
                      加载中...
                    </TableCell>
                  </TableRow>
                ) : jobs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground">
                      暂无任务
                    </TableCell>
                  </TableRow>
                ) : (
                  jobs.map((job) => (
                    <TableRow key={job.id}>
                      <TableCell>{job.id}</TableCell>
                      <TableCell className="font-medium">{job.name}</TableCell>
                      <TableCell>{getJobTypeLabel(job.jobType)}</TableCell>
                      <TableCell className="font-mono text-sm">{job.spec}</TableCell>
                      <TableCell>{getStatusBadge(job.status)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {job.runOn || "自动分配"}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button size="sm" variant="outline">
                            <Play className="h-4 w-4" />
                          </Button>
                          <Button size="sm" variant="outline" onClick={() => setEditingJob(job)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleDelete(job.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
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
      {editingJob && (
        <JobEdit
          job={editingJob}
          onClose={() => setEditingJob(null)}
          onSuccess={() => {
            setEditingJob(null);
            loadJobs();
          }}
        />
      )}
    </MainLayout>
  );
};

export default Jobs;