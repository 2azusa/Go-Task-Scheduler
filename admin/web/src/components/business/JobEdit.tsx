import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Job, HttpMethod } from "@/types/api";
import { api } from "@/lib/api";
import { toast } from "sonner";

interface JobEditProps {
  job: Partial<Job> | null;
  onClose: () => void;
  onSuccess: () => void;
}

const httpMethodMap: { [key: string]: HttpMethod } = {
  GET: HttpMethod.GET,
  POST: HttpMethod.POST,
  PUT: HttpMethod.PUT,
  DELETE: HttpMethod.DELETE,
};

const httpMethodReverseMap: { [key: number]: string } = {
  [HttpMethod.GET]: "GET",
  [HttpMethod.POST]: "POST",
  [HttpMethod.PUT]: "PUT",
  [HttpMethod.DELETE]: "DELETE",
};

export const JobEdit = ({ job, onClose, onSuccess }: JobEditProps) => {
  const [formData, setFormData] = useState<Partial<Job>>({});

  useEffect(() => {
    if (job) {
      setFormData({
        ...job,
        httpMethod: job.httpMethod
          ? httpMethodReverseMap[job.httpMethod]
          : "GET",
      } as any);
    } else {
      setFormData({
        jobType: 1,
        status: 1,
        allocation: 2, // Default to Auto
        retryTimes: 0,
        retryInterval: 0,
        timeout: 0,
      });
    }
  }, [job]);

  const handleChange = (name: keyof Job, value: any) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async () => {
    try {
      const apiCall = job?.id ? api.updateJob : api.createJob;
      const dataToSend = {
        ...formData,
        httpMethod: httpMethodMap[formData.httpMethod as unknown as string],
      };
      const response = await apiCall(dataToSend);
      if (response.code === 200) {
        toast.success(job?.id ? "更新成功" : "创建成功");
        onSuccess();
      } else {
        toast.error(response.msg || (job?.id ? "更新失败" : "创建失败"));
      }
    } catch (error) {
      toast.error(job?.id ? "更新失败" : "创建失败");
    }
  };

  return (
    <Dialog open={!!job} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{job?.id ? "编辑任务" : "创建任务"}</DialogTitle>
        </DialogHeader>
        <div className="grid max-h-[70vh] grid-cols-1 gap-4 overflow-y-auto p-1 md:grid-cols-2 md:gap-6">
          <div className="space-y-2">
            <Label htmlFor="name">任务名称</Label>
            <Input
              id="name"
              value={formData.name || ""}
              onChange={(e) => handleChange("name", e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="jobType">任务类型</Label>
            <Select
              value={String(formData.jobType)}
              onValueChange={(value) => handleChange("jobType", parseInt(value))}
            >
              <SelectTrigger>
                <SelectValue placeholder="选择任务类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">命令任务</SelectItem>
                <SelectItem value="2">HTTP任务</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="spec">Cron表达式</Label>
            <Input
              id="spec"
              value={formData.spec || ""}
              onChange={(e) => handleChange("spec", e.target.value)}
            />
          </div>
          {formData.jobType === 1 ? (
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="command">命令</Label>
              <Input
                id="command"
                value={formData.command || ""}
                onChange={(e) => handleChange("command", e.target.value)}
              />
            </div>
          ) : (
            <>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="httpUrl">URL</Label>
                <Input
                  id="httpUrl"
                  value={formData.httpUrl || ""}
                  onChange={(e) => handleChange("httpUrl", e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="httpMethod">Method</Label>
                <Select
                  value={formData.httpMethod as unknown as string || "GET"}
                  onValueChange={(value) => handleChange("httpMethod", value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择HTTP方法" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="GET">GET</SelectItem>
                    <SelectItem value="POST">POST</SelectItem>
                    <SelectItem value="PUT">PUT</SelectItem>
                    <SelectItem value="DELETE">DELETE</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </>
          )}
          <div className="space-y-2">
            <Label htmlFor="timeout">超时时间 (秒)</Label>
            <Input
              id="timeout"
              type="number"
              value={formData.timeout || 0}
              onChange={(e) =>
                handleChange("timeout", parseInt(e.target.value))
              }
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="retryTimes">重试次数</Label>
            <Input
              id="retryTimes"
              type="number"
              value={formData.retryTimes || 0}
              onChange={(e) =>
                handleChange("retryTimes", parseInt(e.target.value))
              }
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="retryInterval">重试间隔 (秒)</Label>
            <Input
              id="retryInterval"
              type="number"
              value={formData.retryInterval || 0}
              onChange={(e) =>
                handleChange("retryInterval", parseInt(e.target.value))
              }
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="allocation">分配方式</Label>
            <Select
              value={String(formData.allocation)}
              onValueChange={(value) =>
                handleChange("allocation", parseInt(value))
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="选择分配方式" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">手动</SelectItem>
                <SelectItem value="2">自动</SelectItem>
              </SelectContent>
            </Select>
          </div>
           <div className="space-y-2">
            <Label htmlFor="runOn">指定节点 (UUID)</Label>
            <Input
              id="runOn"
              value={formData.runOn || ""}
              onChange={(e) => handleChange("runOn", e.target.value)}
              disabled={formData.allocation === 2}
            />
          </div>
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="note">备注</Label>
            <Input
              id="note"
              value={formData.note || ""}
              onChange={(e) => handleChange("note", e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button onClick={handleSubmit}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};